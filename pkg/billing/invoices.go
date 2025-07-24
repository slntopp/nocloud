package billing

import (
	"context"
	"errors"
	"fmt"
	"github.com/arangodb/go-driver"
	epb "github.com/slntopp/nocloud-proto/events"
	"github.com/slntopp/nocloud/pkg/nocloud/payments"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/whmcs_gateway"
	per "github.com/slntopp/nocloud/pkg/nocloud/periods"
	ps "github.com/slntopp/nocloud/pkg/pubsub"
	"github.com/slntopp/nocloud/pkg/pubsub/billing"
	"github.com/slntopp/nocloud/pkg/pubsub/services_registry"
	"github.com/wI2L/jsondiff"
	"golang.org/x/sync/errgroup"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	elpb "github.com/slntopp/nocloud-proto/events_logging"
	ipb "github.com/slntopp/nocloud-proto/instances"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type pair[T any] struct {
	f T
	s T
}

func equalFloats(a, b float64) bool {
	const equalFloatsEpsilon = 0.0001
	return a == b || math.Abs(a-b) < equalFloatsEpsilon
}

func invoicesEqual(a, b *pb.Invoice, ignoreNulls bool) bool {
	if a == nil || b == nil {
		return false
	}
	_a := proto.Clone(a).(*pb.Invoice)
	_b := proto.Clone(b).(*pb.Invoice)
	if ignoreNulls {
		if _a.Items == nil {
			_b.Items = nil
		}
		if _a.Transactions == nil {
			_b.Transactions = nil
		}
		if _a.Instances == nil {
			_b.Instances = nil
		}
	}
	prepare := func(i *pb.Invoice) {
		if i.Items == nil {
			i.Items = []*pb.Item{}
		}
		if i.Transactions == nil {
			i.Transactions = []string{}
		}
		if i.Instances == nil {
			i.Instances = []string{}
		}
		if i.Meta == nil {
			i.Meta = make(map[string]*structpb.Value)
		}
		if i.TaxOptions == nil {
			i.TaxOptions = &pb.TaxOptions{}
		}
	}
	prepare(_a)
	prepare(_b)
	if len(_a.Items) != len(_b.Items) {
		return false
	}
	for i := range _a.Items {
		if _a.Items[i] == nil && _b.Items[i] != nil || _b.Items[i] == nil && _a.Items[i] != nil {
			return false
		}
		if _a.Items[i] == nil {
			continue
		}
		i1 := _a.Items[i]
		i2 := _b.Items[i]
		if i1.Amount != i2.Amount || i1.Description != i2.Description || i1.Unit != i2.Unit || i1.ApplyTax != i2.ApplyTax {
			return false
		}
		if !equalFloats(i1.Price, i2.Price) {
			return false
		}
	}
	if (_a.Currency == nil && _b.Currency != nil) ||
		(_a.Currency != nil && _b.Currency == nil) ||
		(_a.Currency != nil && _b.Currency != nil && _a.Currency.GetId() != _b.Currency.GetId()) {
		return false
	}
	if _a.TaxOptions.TaxRate != _b.TaxOptions.TaxRate {
		return false
	}
	_a.Currency = nil
	_b.Currency = nil
	_a.Items = nil
	_b.Items = nil
	_a.Total = 0 // It calculated based on items anyway
	_b.Total = 0
	patch, err := jsondiff.Compare(_a, _b)
	if err != nil {
		return false
	}
	return len(patch) == 0
}

var forbiddenStatusConversions = make([]pair[pb.BillingStatus], 0)

func checkAdditionalProperties(conf InvoicesConf, inv graph.Invoice, acc graph.Account) error {
	if inv.Properties == nil {
		return nil
	}
	if inv.GetProperties().GetEmailVerificationRequired() && !acc.GetIsEmailVerified() {
		return fmt.Errorf("email is not verified")
	}
	if (inv.GetProperties().GetPhoneVerificationRequired() && !acc.GetIsPhoneVerified()) ||
		(conf.ForceRequirePhoneVerification && !acc.GetIsPhoneVerified()) {
		return fmt.Errorf("phone is not verified")
	}
	return nil
}

const instanceOwner = `
LET account = LAST( // Find Instance owner Account
    FOR node, edge, path IN 4
    INBOUND DOCUMENT(@@instances, @instance)
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner","owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@accounts)
        RETURN node
    )
RETURN account`

const invoicesByPaymentDate = `
FOR invoice IN @@invoices
FILTER invoice.payment && invoice.payment > 0
FILTER invoice.payment >= @date_from
FILTER invoice.payment < @date_to
RETURN invoice
`

const unpaidInvoicesByCreatedDate = `
FOR invoice IN @@invoices
FILTER invoice.payment == null || invoice.payment == 0
FILTER invoice.created >= @date_from
FILTER invoice.created < @date_to
RETURN invoice
`

func (s *BillingServiceServer) GetNewNumber(log *zap.Logger, invoicesQuery string, date time.Time, template string, resetMode string) (string, int, error) {
	log = log.Named("GetNewNumber")
	var dateFrom, dateTo int64
	switch resetMode {
	case "DAILY":
		dateFrom = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()).Unix()
		dateTo = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()).
			AddDate(0, 0, 1).Unix()
	case "MONTHLY":
		dateFrom = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).Unix()
		dateTo = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).
			AddDate(0, 1, 0).Unix()
	case "YEARLY":
		dateFrom = time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location()).Unix()
		dateTo = time.Date(date.Year()+1, 1, 1, 0, 0, 0, 0, date.Location()).Unix()
	default:
		log.Info("Reset mode is unknown. Using max range", zap.String("mode", resetMode))
		dateFrom = 0
		dateTo = int64(^uint64(0) >> 1) // max int64
	}

	bindVars := map[string]interface{}{
		"@invoices": schema.INVOICES_COL,
		"date_from": dateFrom,
		"date_to":   dateTo,
	}

	cur, err := s.db.Query(context.Background(), invoicesQuery, bindVars)
	if err != nil {
		log.Error("Failed to get invoices to define number", zap.Error(err))
		return "", 0, fmt.Errorf("failed to get invoices. %w", err)
	}
	defer cur.Close()
	number := 1
	for {
		result := map[string]interface{}{}
		invoice := &graph.Invoice{
			Invoice:           &pb.Invoice{},
			InvoiceNumberMeta: &graph.InvoiceNumberMeta{},
		}
		_, err := cur.ReadDocument(context.Background(), &result)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			log.Error("Failed to get invoices", zap.Error(err))
			return "", 0, fmt.Errorf("failed to decode invoices. %w", err)
		}
		if err = s.invoices.DecodeInvoice(result, invoice); err != nil {
			return "", 0, fmt.Errorf("failed to decode invoice. %w", err)
		}
		if invoice.NumericNumber >= number {
			number = invoice.NumericNumber + 1
		}
	}

	return s.invoices.ParseNumberIntoTemplate(template, number, date), number, nil
}

func (s *BillingServiceServer) GetInvoices(ctx context.Context, r *connect.Request[pb.GetInvoicesRequest]) (*connect.Response[pb.Invoices], error) {
	log := s.log.Named("GetInvoice")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg
	log.Debug("Request received", zap.String("requester", requester))

	acc := requester

	query := `FOR t IN @@invoices`
	vars := map[string]interface{}{
		"@invoices":   schema.INVOICES_COL,
		"@currencies": schema.CUR_COL,
	}

	if req.GetUuid() != "" {
		return s._HandleGetSingleInvoice(ctx, acc, req.GetUuid())
	}

	if req.Account != nil {
		acc = *req.Account
		node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
		if !s.ca.HasAccess(ctx, requester, node, access.Level_ADMIN) && requester != req.GetAccount() {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		query += ` FILTER t.account == @acc`
		vars["acc"] = acc
	} else {
		if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
			query += ` FILTER t.account == @acc && t.status != @statusDraft && t.status != @statusTerm`
			vars["acc"] = acc
			vars["statusDraft"] = pb.BillingStatus_DRAFT
			vars["statusTerm"] = pb.BillingStatus_TERMINATED
		}
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "payment" || key == "total" || key == "processed" || key == "created" || key == "returned" || key == "deadline" {
				values := value.GetStructValue().AsMap()
				if val, ok := values["from"]; ok {
					from := val.(float64)
					query += fmt.Sprintf(` FILTER t["%s"] >= %f`, key, from)
				}

				if val, ok := values["to"]; ok {
					to := val.(float64)
					query += fmt.Sprintf(` FILTER t["%s"] <= %f`, key, to)
				}
			} else if key == "number" {
				query += fmt.Sprintf(` FILTER t["%s"] LIKE "%s"`, key, "%"+value.GetStringValue()+"%")
			} else if key == "whmcs_invoice_id" {
				query += fmt.Sprintf(` FILTER t.meta["%s"] LIKE "%s"`, key, "%"+value.GetStringValue()+"%")
			} else if key == "search_param" {
				query += fmt.Sprintf(`
LET acc = DOCUMENT(@@accounts, t.account)
FILTER LOWER(t["number"]) LIKE LOWER("%s") || t._key LIKE "%s" || t.meta["whmcs_invoice_id"] LIKE "%s" || LOWER(acc.title) LIKE LOWER("%s") || LOWER(acc.data.email) LIKE LOWER("%s")`,
					"%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%")
				vars["@accounts"] = schema.ACCOUNTS_COL
			} else if key == "currency" {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER TO_NUMBER(t.currency.id) in @%s`, "currencyIds")
				vars["currencyIds"] = values
				log.Debug("Added currency filter", zap.Any("values", values), zap.String("query", query))
			} else if key == "instances" {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER LENGTH(INTERSECTION(@%s, t.instances)) > 0`, "instancesUuids")
				vars["instancesUuids"] = values
			} else {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER t["%s"] in @%s`, key, key)
				vars[key] = values
			}
		}
	}

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT t.%s %s`
		field, sort := req.GetField(), req.GetSort()

		if field == "whmcs_invoice_id" {
			field = `meta["whmcs_invoice_id"]`
		}

		if field == "total" {
			if sort == "asc" {
				sort = "desc"
			} else {
				sort = "asc"
			}
		}

		query += fmt.Sprintf(subQuery, field, sort)
	}

	if req.Page != nil && req.Limit != nil {
		if req.GetLimit() != 0 {
			limit, page := req.GetLimit(), req.GetPage()
			offset := (page - 1) * limit

			query += ` LIMIT @offset, @count`
			vars["offset"] = offset
			vars["count"] = limit
		}
	}
	query += ` RETURN merge(t, {uuid: t._key, currency: DOCUMENT(@@currencies, TO_STRING(TO_NUMBER(t.currency.id)))})`

	cursor, err := s.db.Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to retrieve invoices", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve invoices")
	}
	defer cursor.Close()

	conf := MakeInvoicesConf(log, &s.settingsClient)
	requirePhoneVer := conf.ForceRequirePhoneVerification

	var invoices []*pb.Invoice
	for {
		invoice := &pb.Invoice{}
		meta, err := cursor.ReadDocument(ctx, invoice)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			log.Error("Failed to retrieve invoices", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to retrieve invoices")
		}
		invoice.Uuid = meta.Key
		if invoice.Properties == nil {
			invoice.Properties = &pb.AdditionalProperties{}
		}
		if requirePhoneVer {
			invoice.Properties.PhoneVerificationRequired = true
		}
		invoices = append(invoices, invoice)
	}

	resp := connect.NewResponse(&pb.Invoices{Pool: invoices})
	return resp, nil
}

func (s *BillingServiceServer) CreateInvoice(ctx context.Context, req *connect.Request[pb.CreateInvoiceRequest]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("CreateInvoice")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	t := req.Msg.Invoice
	log.Debug("Request received", zap.Any("invoice", t), zap.String("requester", requester))
	invConf := MakeInvoicesConf(log, &s.settingsClient)
	defCurr := MakeCurrencyConf(log, &s.settingsClient).Currency

	includeTaxInPrices := invConf.TaxIncluded

	if t == nil {
		return nil, status.Error(codes.InvalidArgument, "Invoice is nil")
	}
	if t.GetStatus() == pb.BillingStatus_BILLING_STATUS_UNKNOWN {
		t.Status = pb.BillingStatus_DRAFT
	}
	if t.GetType() == pb.ActionType_ACTION_TYPE_UNKNOWN {
		t.Type = pb.ActionType_NO_ACTION
	}
	if t.TaxOptions == nil {
		t.TaxOptions = &pb.TaxOptions{}
	}
	t.TaxOptions.TaxIncluded = includeTaxInPrices

	rootNs := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	rootAccess := s.ca.HasAccess(ctx, requester, rootNs, access.Level_ROOT)
	if !rootAccess && (t.Account != requester || !hasInternalAccess(ctx)) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	if t.Account == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing account")
	}
	if t.Currency == nil {
		t.Currency = defCurr
	}
	if t.Transactions == nil {
		t.Transactions = []string{}
	}
	if t.Instances == nil {
		t.Instances = []string{}
	}
	for _, i := range t.GetInstances() {
		if exists, err := s.instances.Exists(ctx, i); err != nil || !exists {
			log.Error("Given linked instance dont exists", zap.String("instance", i), zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "Instance you want to link do not exist")
		}
	}

	acc, err := s.accounts.GetAccountOrOwnerAccountIfPresent(ctx, t.Account)
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get account")
	}
	tax := acc.GetTaxRate()
	t.TaxOptions.TaxRate = tax

	// Rounding invoice items
	cur, err := s.currencies.Get(ctx, t.Currency.GetId())
	if err != nil {
		log.Error("Failed to get currency", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get currency")
	}
	var newTotal, newSubtotal float64
	for _, item := range t.GetItems() {
		price := graph.Round(item.GetPrice(), cur.Precision, cur.Rounding)
		item.Price = price
		if item.GetApplyTax() {
			if includeTaxInPrices {
				priceRaw := price / (1 + t.GetTaxOptions().GetTaxRate())
				newTotal += price * float64(item.GetAmount())
				newSubtotal += priceRaw * float64(item.GetAmount())
			} else {
				priceTaxed := price + price*t.GetTaxOptions().GetTaxRate()
				newTotal += priceTaxed * float64(item.GetAmount())
				newSubtotal += price * float64(item.GetAmount())
			}
		} else {
			newTotal += price * float64(item.GetAmount())
			newSubtotal += price * float64(item.GetAmount())
		}
	}
	t.Total = graph.Round(newTotal, cur.Precision, cur.Rounding)
	t.Subtotal = graph.Round(newSubtotal, cur.Precision, cur.Rounding)

	now := time.Now()

	var (
		strNum string
		num    int
	)
	if t.Status == pb.BillingStatus_PAID {
		strNum, num, err = s.GetNewNumber(log, invoicesByPaymentDate, now, invConf.Template, invConf.ResetCounterMode)
		if err != nil {
			log.Error("Failed to get next paid number", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to get next number")
		}
	} else {
		strNum, num, err = s.GetNewNumber(log, unpaidInvoicesByCreatedDate, now, invConf.NewTemplate, "NONE")
		if err != nil {
			log.Error("Failed to get new number for invoice", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to get new number for invoice. "+err.Error())
		}
	}

	if t.Meta == nil {
		t.Meta = make(map[string]*structpb.Value)
	}
	if t.Meta["creator"] == nil {
		t.Meta["creator"] = structpb.NewStringValue(requester)
	}
	if acc.GetPaymentsGateway() == "whmcs" || acc.GetPaymentsGateway() == "" {
		t.Meta["whmcs_sync_required"] = structpb.NewBoolValue(true)
	}

	t.Number = strNum
	if t.Created == 0 {
		t.Created = now.Unix()
	}
	t.Processed = 0
	t.Returned = 0
	r, err := s.invoices.Create(ctx, &graph.Invoice{
		Invoice: t,
		InvoiceNumberMeta: &graph.InvoiceNumberMeta{
			NumericNumber:  num,
			NumberTemplate: invConf.NewTemplate,
		},
	})
	if err != nil {
		log.Error("Failed to create invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to create invoice")
	}

	if err = s.invoicesPublisher(&epb.Event{
		Uuid: r.GetUuid(),
		Key:  billing.InvoiceCreated,
		Data: map[string]*structpb.Value{
			"gw-callback": structpb.NewBoolValue(payments.GetGatewayCallbackValue(ctx, req.Header())),
		},
	}); err != nil {
		log.Error("Failed to publish invoice creation", zap.Error(err))
	}

	nocloud.Log(log, &elpb.Event{
		Uuid:      r.GetUuid(),
		Entity:    "Invoices",
		Action:    "create",
		Scope:     "database",
		Rc:        0,
		Ts:        time.Now().Unix(),
		Snapshot:  &elpb.Snapshot{},
		Requestor: requester,
	})

	if t.Total <= 0 {
		if _, err = s.UpdateInvoiceStatus(ctx, connect.NewRequest(&pb.UpdateInvoiceStatusRequest{
			Uuid:   r.GetUuid(),
			Status: pb.BillingStatus_PAID,
		})); err != nil {
			log.Error("Failed to auto-pay 0 or less total invoice", zap.Error(err))
		}
	}

	resp := connect.NewResponse(r.Invoice)
	return resp, nil
}

func (s *BillingServiceServer) UpdateInvoiceStatus(ctx context.Context, req *connect.Request[pb.UpdateInvoiceStatusRequest]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("UpdateInvoiceStatus")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	t := req.Msg
	log.Debug("UpdateInvoiceStatus request received")

	if t.GetStatus() == pb.BillingStatus_BILLING_STATUS_UNKNOWN {
		t.Status = pb.BillingStatus_DRAFT
	}

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	old, err := s.invoices.Get(ctx, t.GetUuid())
	if err != nil {
		log.Error("Failed to get invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get invoice")
	}
	newInv := &graph.Invoice{
		Invoice:           proto.Clone(old.Invoice).(*pb.Invoice),
		DocumentMeta:      driver.DocumentMeta{Key: old.GetUuid()},
		InvoiceNumberMeta: &graph.InvoiceNumberMeta{},
	}

	newStatus := t.GetStatus()
	oldStatus := old.GetStatus()

	if oldStatus == newStatus {
		return nil, status.Error(codes.InvalidArgument, "Same status")
	}
	if slices.Contains(forbiddenStatusConversions, pair[pb.BillingStatus]{oldStatus, newStatus}) {
		return nil, status.Error(codes.InvalidArgument, "Cannot convert from "+oldStatus.String()+" to "+newStatus.String())
	}

	nowBeforeActions := time.Now().Unix()
	if newStatus == pb.BillingStatus_PAID && s.syncCreatedDate {
		newInv.Created = nowBeforeActions
	}
	newInv.Status = newStatus

	var strNum string
	var num int
	invConf := MakeInvoicesConf(log, &s.settingsClient)

	if newStatus == pb.BillingStatus_PAID {
		goto payment
	} else if newStatus == pb.BillingStatus_RETURNED {
		newInv.Returned = nowBeforeActions
		goto quit
	} else {
		goto quit
	}

payment:
	if req.Msg.GetParams().GetPaymentDate() != 0 {
		newInv.Payment = req.Msg.GetParams().GetPaymentDate()
	} else {
		newInv.Payment = nowBeforeActions
	}

	strNum, num, err = s.GetNewNumber(log, invoicesByPaymentDate, time.Unix(newInv.Payment, 0).In(time.Local), invConf.Template, invConf.ResetCounterMode)
	if err != nil {
		log.Error("Failed to get next number", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get next number")
	}
	newInv.Number = strNum
	newInv.NumericNumber = num
	newInv.NumberTemplate = invConf.Template

quit:
	paidWithBalance, _ := ctx.Value("paid-with-balance").(bool)

	_newInv := proto.Clone(newInv.Invoice).(*pb.Invoice)
	if newInv.Meta == nil {
		newInv.Meta = map[string]*structpb.Value{}
	}
	newInv.Meta["paid_with_balance"] = structpb.NewBoolValue(paidWithBalance)
	newInv.Transactions = nil
	newInv.Instances = nil
	newInv.Items = nil
	newInv, err = s.invoices.Update(ctx, newInv)
	if err != nil {
		log.Error("Failed to update invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update invoice")
	}

	if err = s.invoicesPublisher(&epb.Event{
		Uuid: old.GetUuid(),
		Key:  billing.InvoiceStatusToKey(newStatus),
		Data: map[string]*structpb.Value{
			"paid-with-balance": structpb.NewBoolValue(paidWithBalance),
			"gw-callback":       structpb.NewBoolValue(payments.GetGatewayCallbackValue(ctx, req.Header())),
		},
	}); err != nil {
		log.Error("Failed to publish invoice status update", zap.Error(err))
	}

	diff, _ := jsondiff.Compare(_newInv, old.Invoice)
	nocloud.Log(log, &elpb.Event{
		Uuid:      old.GetUuid(),
		Entity:    "Invoices",
		Action:    strings.ToLower(newStatus.String()),
		Scope:     "database",
		Rc:        0,
		Ts:        time.Now().Unix(),
		Snapshot:  &elpb.Snapshot{Diff: diff.String()},
		Requestor: requester,
	})

	log.Info("Finished invoice update status")
	return connect.NewResponse(_newInv), nil
}

func (s *BillingServiceServer) PayWithBalance(ctx context.Context, r *connect.Request[pb.PayWithBalanceRequest]) (*connect.Response[pb.PayWithBalanceResponse], error) {
	log := s.log.Named("PayWithBalance")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req), zap.String("requester", requester))

	if req.WhmcsId != 0 {
		return s.payWithBalanceWhmcsInvoice(ctx, req.WhmcsId)
	}

	inv, err := s.invoices.Get(ctx, req.GetInvoiceUuid())
	if err != nil {
		log.Warn("Failed to get invoice", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Invoice not found")
	}
	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT)
	if !ok && inv.GetAccount() != requester {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	if slices.Contains([]pb.BillingStatus{pb.BillingStatus_CANCELED, pb.BillingStatus_TERMINATED,
		pb.BillingStatus_BILLING_STATUS_UNKNOWN, pb.BillingStatus_DRAFT}, inv.GetStatus()) {
		return nil, status.Error(codes.InvalidArgument, "Can't pay this invoice. Try again later or contact support")
	}

	if inv.GetType() == pb.ActionType_BALANCE {
		return nil, status.Error(codes.InvalidArgument, "Can't pay top-up balance invoice with balance")
	}

	acc, err := s.accounts.Get(ctx, inv.GetAccount())
	if err != nil {
		log.Warn("Failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Account not found")
	}
	currConf := MakeCurrencyConf(log, &s.settingsClient)
	invConf := MakeInvoicesConf(log, &s.settingsClient)

	balance := acc.GetBalance()
	accCurrency := acc.Currency
	if accCurrency == nil {
		accCurrency = currConf.Currency
	}
	invCurrency := inv.Currency
	if invCurrency == nil {
		invCurrency = currConf.Currency
	}

	if accCurrency != invCurrency {
		balance, err = s.currencies.Convert(ctx, accCurrency, invCurrency, balance)
		if err != nil {
			log.Error("Failed to convert balance", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to convert balance")
		}
	}

	rounded := graph.Round(balance, invCurrency.Precision, invCurrency.Rounding)
	if rounded < inv.GetTotal() {
		return nil, status.Error(codes.FailedPrecondition, "Not enough balance to perform operation")
	}

	if err = checkAdditionalProperties(invConf, *inv, acc); err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to pay invoice. Error: "+err.Error())
	}

	ctx = context.WithValue(ctx, "paid-with-balance", true)
	resp, err := s.UpdateInvoiceStatus(ctxWithRoot(ctx), connect.NewRequest(&pb.UpdateInvoiceStatusRequest{
		Uuid:   inv.GetUuid(),
		Status: pb.BillingStatus_PAID,
		Params: &pb.UpdateInvoiceStatusRequest_Params{
			IsSendEmail: true,
		},
	}))
	if err != nil {
		log.Error("Failed to update invoice status", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to paid with balance. Error: "+err.Error())
	}

	log.Debug("Generating transaction after invoice payment")
	noCancelCtx := context.WithoutCancel(ctx)
	trCtx, err := graph.BeginTransaction(noCancelCtx, s.db, driver.TransactionCollections{
		Exclusive: []string{schema.TRANSACTIONS_COL, schema.RECORDS_COL, schema.ACCOUNTS_COL, schema.INVOICES_COL},
	})
	if err != nil {
		log.Error("Failed to start transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to start transaction. Error: "+err.Error())
	}
	abort := func() {
		if err := graph.AbortTransaction(trCtx, s.db); err != nil {
			log.Error("Failed to abort transaction")
		}
	}
	commit := func() error {
		if err := graph.CommitTransaction(trCtx, s.db); err != nil {
			log.Error("Failed to commit transaction")
			return err
		}
		return nil
	}
	tr, err := s.applyTransaction(ctxWithInternalAccess(trCtx), math.Min(balance, inv.GetTotal()), inv.GetAccount(), invCurrency)
	if err != nil {
		abort()
		log.Error("Failed to create transaction. INVOICE WAS PAID, ACTIONS WERE APPLIED, BUT USER HAVEN'T LOSE BALANCE", zap.Error(err))
		return nil, status.Error(codes.Internal, "Invoice was paid but still encountered an error. Error: "+err.Error())
	}
	if err = commit(); err != nil {
		log.Error("Failed to create transaction. INVOICE WAS PAID, ACTIONS WERE APPLIED, BUT USER HAVEN'T LOSE BALANCE", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to commit transaction. Error: "+err.Error())
	}
	if tr != nil {
		respTrans := resp.Msg.Transactions
		if respTrans == nil {
			respTrans = make([]string, 0)
		}
		respTrans = append(respTrans, tr.GetUuid())
		if err = s.invoices.Patch(noCancelCtx, resp.Msg.GetUuid(), map[string]interface{}{
			"transactions": respTrans,
		}); err != nil {
			log.Error("Failed to patch invoice", zap.Error(err))
			return nil, status.Error(codes.Internal, "Invoice was paid but still encountered an error. Error: "+err.Error())
		}
	}

	return connect.NewResponse(&pb.PayWithBalanceResponse{Success: true}), nil
}

func (s *BillingServiceServer) payWithBalanceWhmcsInvoice(ctx context.Context, invId int64) (*connect.Response[pb.PayWithBalanceResponse], error) {
	log := s.log.Named("payWithBalanceWhmcsInvoice")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	log.Info("Paying WHMCS invoice with balance", zap.Int64("id", invId))
	inv, err := s.whmcsGateway.GetInvoice(ctx, int(invId))
	if err != nil {
		log.Warn("Failed to get invoice", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Invoice not found")
	}

	if slices.Contains([]string{"cancelled", "draft", "paid", "terminated"}, strings.ToLower(inv.Status)) {
		return nil, status.Error(codes.InvalidArgument, "Can't pay this invoice. Try again later or contact support")
	}

	ncInv, err := s.whmcsGateway.GetInvoiceByWhmcsId(int(invId))
	if err == nil {
		log.Info("Found NoCloud invoice with this whmcs_id. Redirecting to pay it on NoCloud", zap.Int64("whmcs_id", invId))
		return s.PayWithBalance(ctx, connect.NewRequest(&pb.PayWithBalanceRequest{InvoiceUuid: ncInv.GetUuid(), WhmcsId: 0}))
	}
	if !errors.Is(whmcs_gateway.ErrNotFound, err) {
		log.Error("Failed to ensure that whmcs invoice exists in NoCloud", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error. Couldn't find your invoice")
	}

	acc, err := s.accounts.Get(ctx, requester)
	if err != nil {
		log.Warn("Failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Account not found")
	}
	clientIdVal, ok := acc.GetData().AsMap()["whmcs_id"].(float64)
	if !ok {
		log.Warn("Failed to get client id", zap.Error(err))
		return nil, status.Error(codes.Internal, "Client not found")
	}
	clientId := int(clientIdVal)
	if inv.UserId != clientId {
		return nil, status.Error(codes.PermissionDenied, "No access to this invoice")
	}
	currConf := MakeCurrencyConf(log, &s.settingsClient)

	balance := acc.GetBalance()
	accCurrency := acc.Currency
	if accCurrency == nil {
		accCurrency = currConf.Currency
	}
	invCurrency := acc.Currency

	if accCurrency.GetId() != invCurrency.GetId() {
		balance, err = s.currencies.Convert(ctx, accCurrency, invCurrency, balance)
		if err != nil {
			log.Error("Failed to convert balance", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to convert balance")
		}
	}

	rounded := graph.Round(balance, invCurrency.Precision, invCurrency.Rounding)
	if rounded < float64(inv.Balance) {
		return nil, status.Error(codes.FailedPrecondition, "Not enough balance to perform operation")
	}

	if err = s.whmcsGateway.PayInvoice(ctx, int(invId), true); err != nil {
		log.Error("Failed to pay invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to perform payment with balance. Error: "+err.Error())
	}

	log.Debug("Generating transaction after whmcs invoice payment")
	noCancelCtx := context.WithoutCancel(ctx)
	trCtx, err := graph.BeginTransaction(noCancelCtx, s.db, driver.TransactionCollections{
		Exclusive: []string{schema.TRANSACTIONS_COL, schema.RECORDS_COL, schema.ACCOUNTS_COL, schema.INVOICES_COL},
	})
	if err != nil {
		log.Error("Failed to start transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to start transaction. Error: "+err.Error())
	}
	abort := func() {
		if err := graph.AbortTransaction(trCtx, s.db); err != nil {
			log.Error("Failed to abort transaction")
		}
	}
	commit := func() error {
		if err := graph.CommitTransaction(trCtx, s.db); err != nil {
			log.Error("Failed to commit transaction")
			return err
		}
		return nil
	}
	_, err = s.applyTransaction(ctxWithInternalAccess(trCtx), math.Min(balance, float64(inv.Balance)), requester, invCurrency)
	if err != nil {
		abort()
		log.Error("Failed to create transaction. INVOICE WAS PAID, ACTIONS WERE APPLIED, BUT USER HAVEN'T LOSE BALANCE", zap.Error(err))
		return nil, status.Error(codes.Internal, "Invoice was paid but still encountered an error. Error: "+err.Error())
	}
	if err = commit(); err != nil {
		log.Error("Failed to create transaction. INVOICE WAS PAID, ACTIONS WERE APPLIED, BUT USER HAVEN'T LOSE BALANCE", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to commit transaction. Error: "+err.Error())
	}

	return connect.NewResponse(&pb.PayWithBalanceResponse{Success: true}), nil
}

func (s *BillingServiceServer) GetInvoicesCount(ctx context.Context, r *connect.Request[pb.GetInvoicesCountRequest]) (*connect.Response[pb.GetInvoicesCountResponse], error) {
	log := s.log.Named("GetInvoicesCount")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req), zap.String("requester", requester))

	acc := requester

	query := `FOR t IN @@invoices`
	vars := map[string]interface{}{
		"@invoices": schema.INVOICES_COL,
	}

	if req.Account != nil {
		acc = *req.Account
		node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
		if !s.ca.HasAccess(ctx, requester, node, access.Level_ADMIN) && requester != req.GetAccount() {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		query += ` FILTER t.account == @acc`
		vars["acc"] = acc
	} else {
		if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
			query += ` FILTER t.account == @acc && t.status != @statusDraft && t.status != @statusTerm`
			vars["acc"] = acc
			vars["statusDraft"] = pb.BillingStatus_DRAFT
			vars["statusTerm"] = pb.BillingStatus_TERMINATED
		}
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "payment" || key == "total" || key == "processed" || key == "created" || key == "returned" || key == "deadline" {
				values := value.GetStructValue().AsMap()
				if val, ok := values["from"]; ok {
					from := val.(float64)
					query += fmt.Sprintf(` FILTER t["%s"] >= %f`, key, from)
				}

				if val, ok := values["to"]; ok {
					to := val.(float64)
					query += fmt.Sprintf(` FILTER t["%s"] <= %f`, key, to)
				}
			} else if key == "number" {
				query += fmt.Sprintf(` FILTER t["%s"] LIKE "%s"`, key, "%"+value.GetStringValue()+"%")
			} else if key == "whmcs_invoice_id" {
				query += fmt.Sprintf(` FILTER t.meta["%s"] LIKE "%s"`, key, "%"+value.GetStringValue()+"%")
			} else if key == "search_param" {
				query += fmt.Sprintf(`
LET acc = DOCUMENT(@@accounts, t.account)
FILTER LOWER(t["number"]) LIKE LOWER("%s") || t._key LIKE "%s" || t.meta["whmcs_invoice_id"] LIKE "%s" || LOWER(acc.title) LIKE LOWER("%s") || LOWER(acc.data.email) LIKE LOWER("%s")`,
					"%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%")
				vars["@accounts"] = schema.ACCOUNTS_COL
			} else if key == "currency" {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER TO_NUMBER(t.currency.id) in @%s`, "currencyIds")
				vars["currencyIds"] = values
				log.Debug("Added currency filter", zap.Any("values", values), zap.String("query", query))
			} else {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER t["%s"] in @%s`, key, key)
				vars[key] = values
			}
		}
	}

	query += ` RETURN t`

	log.Debug("Ready to retrieve invoices", zap.String("query", query), zap.Any("vars", vars))

	queryContext := driver.WithQueryCount(ctx)

	cursor, err := s.db.Query(queryContext, query, vars)
	if err != nil {
		log.Error("Failed to retrieve invoices", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve invoices")
	}
	defer cursor.Close()

	log.Info("invoices count", zap.Int64("count", cursor.Count()))

	resp := connect.NewResponse(&pb.GetInvoicesCountResponse{
		Total: uint64(cursor.Count()),
	})

	return resp, nil
}

func (s *BillingServiceServer) UpdateInvoice(ctx context.Context, r *connect.Request[pb.UpdateInvoiceRequest]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("UpdateInvoice")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg.Invoice
	ignoreNulls := r.Msg.IgnoreNullFields
	log.Debug("Request received", zap.Any("invoice", req), zap.String("requester", requester))

	invConf := MakeInvoicesConf(log, &s.settingsClient)
	includeTaxInPrices := invConf.TaxIncluded

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "Invoice is nil")
	}
	if req.GetStatus() == pb.BillingStatus_BILLING_STATUS_UNKNOWN {
		req.Status = pb.BillingStatus_DRAFT
	}
	if req.GetType() == pb.ActionType_ACTION_TYPE_UNKNOWN {
		req.Type = pb.ActionType_NO_ACTION
	}
	if req.TaxOptions == nil {
		req.TaxOptions = &pb.TaxOptions{}
	}
	req.TaxOptions.TaxIncluded = includeTaxInPrices

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t, err := s.invoices.Get(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to get invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get invoice")
	}
	if invoicesEqual(req, t.Invoice, ignoreNulls) {
		log.Info("Invoice unchanged. Skip", zap.Any("invoice", t.GetUuid()))
		return connect.NewResponse(t.Invoice), nil
	}
	t.Currency = &pb.Currency{Id: t.Currency.GetId()}
	old := proto.Clone(t.Invoice).(*pb.Invoice)

	if req.GetPayment() != 0 && t.GetPayment() != 0 {
		t.Payment = req.GetPayment()
	}
	if req.GetProcessed() != 0 && t.GetProcessed() != 0 {
		t.Processed = req.GetProcessed()
	}
	if req.GetReturned() != 0 && t.GetReturned() != 0 {
		t.Returned = req.GetReturned()
	}
	if req.GetCreated() != 0 {
		t.Created = req.GetCreated()
	}
	t.Deadline = req.GetDeadline()
	t.Uuid = req.GetUuid()
	t.Status = req.GetStatus()
	t.Account = req.GetAccount()
	t.Type = req.GetType()
	t.Properties = req.GetProperties()
	if req.Meta != nil {
		t.Meta = req.GetMeta()
	}
	var diffTax = t.GetTaxOptions().GetTaxRate() != req.GetTaxOptions().GetTaxRate()
	t.TaxOptions = req.GetTaxOptions()

	if t.Account == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing account")
	}

	if req.Instances != nil {
		if !slices.Equal(t.GetInstances(), req.GetInstances()) {
			for _, i := range req.GetInstances() {
				if exists, err := s.instances.Exists(ctx, i); err != nil || !exists {
					log.Error("Given linked instance dont exists", zap.String("instance", i), zap.Error(err))
					return nil, status.Error(codes.InvalidArgument, "Instance you want to link do not exist")
				}
			}
		}
	}

	if req.Items != nil || !ignoreNulls || diffTax {
		var items = t.GetItems()
		if req.Items != nil {
			items = req.GetItems()
		}
		cur, err := s.currencies.Get(ctx, t.Currency.GetId())
		if err != nil {
			log.Error("Failed to get currency", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to get currency")
		}
		var newTotal, newSubtotal float64
		for _, item := range items {
			price := graph.Round(item.GetPrice(), cur.Precision, cur.Rounding)
			item.Price = price
			if item.GetApplyTax() {
				if includeTaxInPrices {
					priceRaw := price / (1 + t.GetTaxOptions().GetTaxRate())
					newTotal += price * float64(item.GetAmount())
					newSubtotal += priceRaw * float64(item.GetAmount())
				} else {
					priceTaxed := price + price*t.GetTaxOptions().GetTaxRate()
					newTotal += priceTaxed * float64(item.GetAmount())
					newSubtotal += price * float64(item.GetAmount())
				}
			} else {
				newTotal += price * float64(item.GetAmount())
				newSubtotal += price * float64(item.GetAmount())
			}
		}
		t.Total = graph.Round(newTotal, cur.Precision, cur.Rounding)
		t.Subtotal = graph.Round(newSubtotal, cur.Precision, cur.Rounding)
	}

	vars := map[string]interface{}{}
	query := fmt.Sprintf("UPDATE @invoice WITH {")
	if req.Transactions != nil || !ignoreNulls {
		query += ` transactions: @transactions,`
		vars["transactions"] = req.GetTransactions()
	}
	if req.Items != nil || !ignoreNulls {
		query += ` items: @items,`
		vars["items"] = req.GetItems()
	}
	if req.Instances != nil || !ignoreNulls {
		query += ` instances: @instances,`
		vars["instances"] = req.GetInstances()
	}
	query += " } IN @@invoices OPTIONS { keepNull: false } "
	vars["invoice"] = t.GetUuid()
	vars["@invoices"] = schema.INVOICES_COL
	// Update + patch transaction
	if ctx, err = graph.BeginTransaction(ctx, s.db, driver.TransactionCollections{
		Write: []string{schema.INVOICES_COL},
	}); err != nil {
		log.Error("Failed to start transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to start transaction")
	}
	abort := func() {
		if err := graph.AbortTransaction(ctx, s.db); err != nil {
			log.Error("Failed to abort transaction")
		}
	}
	upd, err := s.invoices.Update(ctx, t)
	if err != nil {
		abort()
		log.Error("Failed to update invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update invoice")
	}
	if _, err = s.db.Query(ctx, query, vars); err != nil {
		abort()
		log.Error("Failed to patch invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update invoice")
	}
	if upd, err = s.invoices.Get(ctx, t.GetUuid()); err != nil {
		abort()
		log.Error("Failed to get updated invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get updated invoice")
	}
	if err = graph.CommitTransaction(ctx, s.db); err != nil {
		abort()
		log.Error("Failed to commit transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to commit transaction")
	}

	if err = s.invoicesPublisher(&epb.Event{
		Uuid: old.GetUuid(),
		Key:  billing.InvoiceUpdated,
		Data: map[string]*structpb.Value{
			"gw-callback": structpb.NewBoolValue(payments.GetGatewayCallbackValue(ctx, r.Header())),
		},
	}); err != nil {
		log.Error("Failed to publish invoice update", zap.Error(err))
	}

	diff, _ := jsondiff.Compare(upd.Invoice, old)
	nocloud.Log(log, &elpb.Event{
		Uuid:      upd.GetUuid(),
		Entity:    "Invoices",
		Action:    "update",
		Scope:     "database",
		Rc:        0,
		Ts:        time.Now().Unix(),
		Snapshot:  &elpb.Snapshot{Diff: diff.String()},
		Requestor: requester,
	})

	log.Info("Invoice updated", zap.Any("invoice", t.GetUuid()))
	return connect.NewResponse(upd.Invoice), nil
}

func (s *BillingServiceServer) GetInvoice(ctx context.Context, r *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("GetInvoice")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("invoice", req), zap.String("requester", requester))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t, err := s.invoices.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}

	conf := MakeInvoicesConf(s.log, &s.settingsClient)
	if t.Properties == nil {
		t.Properties = &pb.AdditionalProperties{}
	}
	if conf.ForceRequirePhoneVerification {
		t.Properties.PhoneVerificationRequired = true
	}

	return connect.NewResponse(t.Invoice), nil
}

func (s *BillingServiceServer) GetInvoiceSettingsTemplateExample(_ context.Context, _req *connect.Request[pb.GetInvoiceSettingsTemplateExampleRequest]) (*connect.Response[pb.GetInvoiceSettingsTemplateExampleResponse], error) {
	log := s.log.Named("GetInvoiceSettingsTemplateExample")
	req := _req.Msg
	log.Debug("Request received")

	example := s.invoices.ParseNumberIntoTemplate(req.Template, 1, time.Now())
	newExample := s.invoices.ParseNumberIntoTemplate(req.NewTemplate, 1, time.Now())
	var renewalExample string
	if req.IssueRenewalInvoiceAfter > 0 && req.IssueRenewalInvoiceAfter < 1 {
		monthDur := time.Duration(86400*30*(1-req.IssueRenewalInvoiceAfter)) * time.Second
		renewalExample = fmt.Sprintf("**FOR MONTHLY PERIOD** Invoice will be issued before: %s", monthDur.String())
	} else if req.IssueRenewalInvoiceAfter == 1 {
		renewalExample = fmt.Sprintf("Invoice will be issued right after instance expiration")
	} else {
		monthDur := time.Duration(86400*30*0.1) * time.Second
		renewalExample = fmt.Sprintf("Value must be (0:1]. Using default 0.9. **FOR MONTHLY PERIOD** Invoice will be issued before: %s", monthDur.String())
	}
	return connect.NewResponse(&pb.GetInvoiceSettingsTemplateExampleResponse{TemplateExample: example, NewTemplateExample: newExample, IssueRenewalInvoiceAfterExample: renewalExample}), nil
}

func (s *BillingServiceServer) Pay(ctx context.Context, _req *connect.Request[pb.PayRequest]) (*connect.Response[pb.PayResponse], error) {
	log := s.log.Named("Pay")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	req := _req.Msg
	log.Debug("Request received")

	inv, err := s.invoices.Get(ctx, req.InvoiceId)
	if err != nil {
		log.Warn("Error getting invoice", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Internal error or not found")
	}

	if requester != inv.Account {
		ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
		if isRoot := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT); !isRoot {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
	}

	if slices.Contains([]pb.BillingStatus{pb.BillingStatus_CANCELED, pb.BillingStatus_TERMINATED,
		pb.BillingStatus_BILLING_STATUS_UNKNOWN, pb.BillingStatus_DRAFT}, inv.GetStatus()) {
		return nil, status.Error(codes.InvalidArgument, "Can't pay this invoice. Try again later or contact support")
	}

	acc, err := s.accounts.Get(ctx, inv.Account)
	if err != nil {
		log.Error("Error getting account", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Internal error")
	}

	if err = checkAdditionalProperties(MakeInvoicesConf(log, &s.settingsClient), *inv, acc); err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to pay invoice. Error: "+err.Error())
	}

	uri, err := payments.GetPaymentGateway(acc.GetPaymentsGateway()).PaymentURI(ctx, inv.Invoice)
	if err != nil {
		log.Error("Error getting payment uri", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return connect.NewResponse(&pb.PayResponse{PaymentLink: uri}), nil
}

func (s *BillingServiceServer) CreateTopUpBalanceInvoice(ctx context.Context, _req *connect.Request[pb.CreateTopUpBalanceInvoiceRequest]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("CreateTopUpBalanceInvoice")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	req := _req.Msg
	log.Debug("Request received")

	if req.GetSum() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Sum must be greater than 0")
	}

	acc, err := s.accounts.GetAccountOrOwnerAccountIfPresent(ctx, requester)
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get account")
	}
	tax := acc.GetTaxRate()

	ivnToCreate := &pb.Invoice{
		Deadline: time.Now().Add(72 * time.Hour).Unix(),
		Status:   pb.BillingStatus_UNPAID,
		Account:  acc.GetUuid(),
		Total:    req.GetSum(),
		Type:     pb.ActionType_BALANCE,
		Items: []*pb.Item{
			{
				Amount:      1,
				Unit:        "Pcs",
				Price:       req.GetSum(),
				Description: "Пополнение баланса (услуги хостинга, оплата за сервисы)",
				ApplyTax:    true,
			},
		},
		Meta: map[string]*structpb.Value{
			"creator": structpb.NewStringValue(requester),
		},
		TaxOptions: &pb.TaxOptions{
			TaxRate: tax,
		},
	}

	if acc.Currency != nil {
		ivnToCreate.Currency = acc.Currency
	}

	return s.CreateInvoice(ctxWithInternalAccess(ctx), connect.NewRequest(&pb.CreateInvoiceRequest{
		IsSendEmail: true,
		Invoice:     ivnToCreate,
	}))
}

func (s *BillingServiceServer) CreateRenewalInvoice(ctx context.Context, _req *connect.Request[pb.CreateRenewalInvoiceRequest]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("CreateRenewalInvoice")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	requesterId := driver.NewDocumentID(schema.ACCOUNTS_COL, requester)
	req := _req.Msg
	log = log.With(zap.String("instance", req.GetInstance()), zap.String("requester", requester))
	log.Debug("Request received")

	currencyConf := MakeCurrencyConf(log, &s.settingsClient)

	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.INSTANCES_COL, req.GetInstance()), access.Level_ADMIN) {
		log.Warn("Not enough access rights")
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	inst, err := s.instances.GetWithAccess(ctx, requesterId, req.GetInstance())
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get instance")
	}

	res, err := s.instances.GetGroup(ctx, driver.NewDocumentID(schema.INSTANCES_COL, inst.GetUuid()).String())
	if err != nil || res.Group == nil || res.SP == nil {
		log.Error("Error getting instance group or sp", zap.Error(err), zap.Any("group", res.Group), zap.Any("sp", res.SP))
		return nil, status.Error(codes.Internal, "Error getting instance group or sp")
	}

	log = log.With(zap.String("driver", res.SP.GetType()))

	client, ok := s.drivers[res.SP.GetType()]
	if !ok {
		log.Error("Error getting driver. Driver not registered")
		return nil, status.Error(codes.Internal, "Error getting driver. Driver not registered")
	}

	resp, err := client.GetExpiration(ctx, &driverpb.GetExpirationRequest{
		Instance:         inst.Instance,
		ServicesProvider: res.SP,
	})
	if err != nil {
		log.Error("Error getting expiration", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting expiration")
	}
	records := resp.GetRecords()
	log.Info("Got expiration records", zap.Any("records", records))

	periods := make([]int64, 0)
	expirings := make([]int64, 0)
	_processed := 0
	_expiring := 0
	for _, r := range records {
		if r.Product == "" {
			continue
		}
		periods = append(periods, r.Period)
		expirings = append(expirings, r.Expires)
		_expiring++
		_processed++
	}

	if len(periods) == 0 || len(expirings) == 0 {
		log.Error("Error getting periods or expirings. No data")
		return nil, status.Error(codes.Internal, "Error getting periods or expirings. No data")
	}

	if _processed > _expiring {
		log.Warn("WARN. Instance gonna be renewed asynchronously. Product, resources or addons has different periods")
	}

	initCost, _ := s.instances.CalculateInstanceEstimatePrice(inst.Instance, false)
	_, summary, err := s.promocodes.GetDiscountPriceByInstance(inst.Instance, false, true)
	if err != nil {
		log.Error("Error calculating instance estimate price", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error calculating instance estimate price")
	}

	acc, err := s.instances.GetInstanceOwner(ctx, req.GetInstance())
	if err != nil {
		log.Error("Error getting instance owner", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting instance owner")
	}
	acc, err = s.accounts.GetAccountOrOwnerAccountIfPresent(ctx, acc.GetUuid())
	if err != nil {
		log.Error("Error getting instance owner when getting subaccount", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting instance owner")
	}

	if acc.Currency == nil {
		acc.Currency = currencyConf.Currency
	}
	rate, _, err := s.currencies.GetExchangeRate(ctx, currencyConf.Currency, acc.Currency)
	if err != nil {
		log.Error("Error getting exchange rate", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting exchange rate")
	}

	now := time.Now().Unix()
	initCost *= rate
	slices.Sort(periods)
	slices.Sort(expirings)
	period := periods[len(periods)-1]
	expire := expirings[0]
	expireDate := time.Unix(expire, 0)

	forbiddenStatuses := []pb.BillingStatus{pb.BillingStatus_BILLING_STATUS_UNKNOWN, pb.BillingStatus_DRAFT,
		pb.BillingStatus_CANCELED, pb.BillingStatus_TERMINATED, pb.BillingStatus_RETURNED}
	existingInv, err := s.invoices.GetByExpiration(ctx, expire, inst.GetUuid(), forbiddenStatuses)
	if err != nil {
		if !errors.Is(err, graph.ErrNotFound) {
			log.Error("Error getting invoice by expiration", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal error")
		}
	} else {
		response := connect.NewResponse(existingInv.Invoice)
		return response, nil
	}

	var untilDate time.Time
	const monthSecs = 3600 * 24 * 30
	const daySecs = 3600 * 24
	if period == monthSecs {
		if inst.GetMeta() != nil && inst.GetMeta().Started > 0 {
			untilDate = time.Unix(per.GetNextDate(expire, per.BillingMonth, inst.GetMeta().Started), 0)
		} else {
			untilDate = expireDate.AddDate(0, 1, 0)
		}
	} else {
		untilDate = expireDate.Add(time.Duration(period) * time.Second)
	}
	if untilDate.Unix()-expireDate.Unix() > daySecs {
		untilDate = untilDate.AddDate(0, 0, -1)
	}

	fDateNum := func(d int) string {
		if d < 10 {
			return fmt.Sprintf("0%d", d)
		}
		return fmt.Sprintf("%d", d)
	}

	var dueDate = expire
	if dueDate < now {
		dueDate = now + period/2
	}

	bp := inst.GetBillingPlan()
	product, hasProduct := bp.GetProducts()[inst.GetProduct()]
	if !hasProduct {
		log.Warn("Product not found in billing plan", zap.String("product", inst.GetProduct()))
	}
	invoicePrefixVal, _ := bp.GetMeta()["prefix"]
	invoicePrefix := invoicePrefixVal.GetStringValue() + " "
	productTitle := product.GetTitle() + " "
	renewDescription := fmt.Sprintf("%s%s(%s.%s.%d - %s.%s.%d)", invoicePrefix, productTitle,
		fDateNum(expireDate.Day()), fDateNum(int(expireDate.Month())), expireDate.Year(),
		fDateNum(untilDate.Day()), fDateNum(int(untilDate.Month())), untilDate.Year())
	renewDescription = strings.TrimSpace(renewDescription)

	tax := acc.GetTaxRate()
	invCost := initCost

	promoItems := make([]*pb.Item, 0)
	for _, sum := range summary {
		price := -sum.DiscountAmount * rate
		promoItems = append(promoItems, &pb.Item{
			Description: fmt.Sprintf("Скидка %s (промокод %s)", renewDescription, sum.Code),
			Amount:      1,
			Unit:        "Pcs",
			Price:       price,
			ApplyTax:    true,
		})
	}
	items := []*pb.Item{
		{
			Description: renewDescription,
			Amount:      1,
			Unit:        "Pcs",
			Price:       invCost,
			ApplyTax:    true,
		},
	}
	items = append(items, promoItems...)
	billingData := graph.BillingData{
		RenewalData: map[string]graph.RenewalData{
			inst.GetUuid(): {
				ExpirationTs: expire,
			},
		},
	}
	inv := &pb.Invoice{
		Status:    pb.BillingStatus_UNPAID,
		Items:     items,
		Total:     invCost,
		Type:      pb.ActionType_INSTANCE_RENEWAL,
		Instances: []string{inst.GetUuid()},
		Created:   now,
		Deadline:  dueDate, // Until when invoice should be paid
		Account:   acc.GetUuid(),
		Currency:  acc.Currency,
		Meta: map[string]*structpb.Value{
			"creator": structpb.NewStringValue(requester),
		},
		TaxOptions: &pb.TaxOptions{
			TaxRate: tax,
		},
		Properties: &pb.AdditionalProperties{
			PhoneVerificationRequired: bp.GetProperties().GetPhoneVerificationRequired(),
			EmailVerificationRequired: bp.GetProperties().GetEmailVerificationRequired(),
		},
	}
	inv = graph.SetInvoiceBillingData(inv, &billingData)

	if val, ok := ctx.Value("create_as_draft").(bool); ok && val {
		inv.Status = pb.BillingStatus_DRAFT
	}

	return s.CreateInvoice(ctxWithInternalAccess(ctx), connect.NewRequest(&pb.CreateInvoiceRequest{
		IsSendEmail: true,
		Invoice:     inv,
	}))
}

func (s *BillingServiceServer) _HandleGetSingleInvoice(ctx context.Context, acc, uuid string) (*connect.Response[pb.Invoices], error) {
	tr, err := s.invoices.Get(ctx, uuid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Invoice doesn't exist")
	}

	if ok := s.ca.HasAccess(ctx, acc, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT); !ok {
		if ok := s.ca.HasAccess(ctx, acc, driver.NewDocumentID(schema.ACCOUNTS_COL, tr.Account), access.Level_ADMIN); !ok && acc != tr.GetAccount() {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
	}

	conf := MakeInvoicesConf(s.log, &s.settingsClient)
	if tr.Properties == nil {
		tr.Properties = &pb.AdditionalProperties{}
	}
	if conf.ForceRequirePhoneVerification {
		tr.Properties.PhoneVerificationRequired = true
	}

	resp := connect.NewResponse(&pb.Invoices{Pool: []*pb.Invoice{tr.Invoice}})

	return resp, nil
}

func (s *BillingServiceServer) executePostPaidActions(ctx context.Context, log *zap.Logger, inv *graph.Invoice, defCurr *pb.Currency) (*graph.Invoice, error) {

	switch inv.GetType() {
	case pb.ActionType_BALANCE:
		tr, err := s.applyTransaction(ctx, -inv.GetSubtotal(), inv.GetAccount(), inv.GetCurrency())
		if err != nil {
			return inv, fmt.Errorf("failed to apply transaction: %w", err)
		}
		if inv.Transactions == nil {
			inv.Transactions = make([]string, 0)
		}
		if tr != nil {
			inv.Transactions = append(inv.Transactions, tr.GetUuid())
		}
		if err = s.accounts.InvalidateBalanceEvents(ctx, inv.GetAccount()); err != nil {
			s.log.Error("Failed to invalidate balance events for user", zap.Error(err), zap.String("account", inv.GetAccount()))
		}

	case pb.ActionType_INSTANCE_START:
		for _, i := range inv.GetInstances() {
			if i == "" {
				continue
			}
			log := log.With(zap.String("instance", i))

			instOld, err := s.instances.GetWithAccess(ctx, driver.NewDocumentID(schema.ACCOUNTS_COL, schema.ROOT_ACCOUNT_KEY), i)
			if err != nil || instOld.Instance == nil {
				log.Error("Failed to get instance to start", zap.Error(err))
				return nil, fmt.Errorf("failed to get instance to start: %w", err)
			}
			instOld.Uuid = instOld.Key
			// Set auto_start to true. After next driver monitoring instance will be started
			instNew := graph.Instance{
				Instance: proto.Clone(instOld.Instance).(*ipb.Instance),
			}
			cfg := instNew.Config
			if cfg == nil {
				cfg = map[string]*structpb.Value{}
			}
			cfg["auto_start"] = structpb.NewBoolValue(true)
			instNew.Config = cfg
			// Add balance to compensate instance first payment
			acc, err := s.instances.GetInstanceOwner(ctx, i)
			if err != nil {
				return inv, err
			}
			cost, _, err := s.promocodes.GetDiscountPriceByInstance(instOld.Instance, true, true)
			if err != nil {
				return inv, fmt.Errorf("failed to get instance price: %w", err)
			}
			cost, err = s.currencies.Convert(ctx, defCurr, acc.GetCurrency(), cost)
			if err != nil {
				log.Error("Failed to convert cost", zap.Error(err))
				return inv, fmt.Errorf("failed to convert cost: %w", err)
			}
			_, err = s.applyTransaction(ctx, -cost, acc.GetUuid(), acc.GetCurrency())
			if err != nil {
				return inv, fmt.Errorf("failed to apply transaction: %w", err)
			}
			// Update instance in the end due to publish operations inside
			if err := s.instances.Update(ctx, "", instNew.Instance, instOld.Instance); err != nil {
				log.Error("Failed to update instance", zap.Error(err))
				return nil, fmt.Errorf("failed to update instance: %w", err)
			}
		}

	case pb.ActionType_INSTANCE_RENEWAL:
		_z := 0
		errs := &_z
		m := &sync.Mutex{}
		g := &errgroup.Group{}
		for _, i := range inv.GetInstances() {
			id := i
			g.Go(func() error {
				if err := s.instanceCommandsPub(&epb.Event{
					Uuid: id,
					Key:  services_registry.CommandInstanceInvoke,
					Type: "free_renew",
				}); err != nil {
					m.Lock()
					*errs = *errs + 1
					m.Unlock()
					return err
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			if *errs == len(inv.GetInstances()) {
				return inv, fmt.Errorf("failed to publish free_renew command: %w", err)
			}
			log.Error("FATAL: somehow published only some of free_renew commands", zap.Error(err), zap.Int("errs", *errs), zap.Int("instances", len(inv.GetInstances())))
			return inv, ps.NoNackErr(fmt.Errorf("failed to publish free_renew command for some instances, but not for all: %w", err))
		}
	}

	return inv, nil
}

func (s *BillingServiceServer) executePostRefundActions(ctx context.Context, log *zap.Logger, inv *graph.Invoice) (*graph.Invoice, error) {

	// Reverting invoice transactions
	transactions := make([]string, 0)
	if inv.Transactions == nil {
		inv.Transactions = make([]string, 0)
	}
	for _, trId := range inv.GetTransactions() {
		tr, err := s.transactions.Get(ctx, trId)
		if err != nil {
			log.Error("Failed to get transaction", zap.Error(err))
			return nil, fmt.Errorf("failed to get transaction: %w", err)
		}
		if tr, err = s.applyTransaction(ctx, -tr.GetTotal(), tr.GetAccount(), tr.GetCurrency()); err != nil {
			log.Error("Failed to apply transaction", zap.Error(err))
			return nil, fmt.Errorf("failed to apply transaction: %w", err)
		}
		if tr != nil {
			transactions = append(transactions, tr.GetUuid())
		}
	}
	inv.Transactions = append(inv.Transactions, transactions...)

	switch inv.GetType() {
	case pb.ActionType_INSTANCE_START:
		_z := 0
		errs := &_z
		m := &sync.Mutex{}
		g := &errgroup.Group{}
		for _, i := range inv.GetInstances() {
			id := i
			g.Go(func() error {
				if err := s.instanceCommandsPub(&epb.Event{
					Uuid: id,
					Key:  services_registry.CommandInstanceInvoke,
					Type: "suspend",
				}); err != nil {
					m.Lock()
					*errs = *errs + 1
					m.Unlock()
					return err
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			if *errs == len(inv.GetInstances()) {
				return inv, fmt.Errorf("failed to publish suspend command: %w", err)
			}
			log.Error("FATAL: somehow published only some of suspend commands", zap.Error(err), zap.Int("errs", *errs), zap.Int("instances", len(inv.GetInstances())))
			return inv, ps.NoNackErr(fmt.Errorf("failed to publish suspend command for some instances, but not for all: %w", err))
		}

	case pb.ActionType_INSTANCE_RENEWAL:
		_z := 0
		errs := &_z
		m := &sync.Mutex{}
		g := &errgroup.Group{}
		for _, i := range inv.GetInstances() {
			id := i
			g.Go(func() error {
				if err := s.instanceCommandsPub(&epb.Event{
					Uuid: id,
					Key:  services_registry.CommandInstanceInvoke,
					Type: "cancel_renew",
				}); err != nil {
					m.Lock()
					*errs = *errs + 1
					m.Unlock()
					return err
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			if *errs == len(inv.GetInstances()) {
				return inv, fmt.Errorf("failed to publish cancel_renew command: %w", err)
			}
			log.Error("FATAL: somehow published only some of cancel_renew commands", zap.Error(err), zap.Int("errs", *errs), zap.Int("instances", len(inv.GetInstances())))
			return inv, ps.NoNackErr(fmt.Errorf("failed to publish cancel_renew command for some instances, but not for all: %w", err))
		}
	}

	return inv, nil
}
