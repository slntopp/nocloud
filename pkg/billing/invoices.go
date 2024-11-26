package billing

import (
	"context"
	"fmt"
	"github.com/slntopp/nocloud/pkg/nocloud/payments"
	"math"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	elpb "github.com/slntopp/nocloud-proto/events_logging"
	instancespb "github.com/slntopp/nocloud-proto/instances"
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

const invoicesAccessKey = nocloud.ContextKey("invoices-internal-access")

func ctxWithRoot(ctx context.Context) context.Context {
	return context.WithValue(ctx, nocloud.NoCloudAccount, schema.ROOT_NAMESPACE_KEY)
}

func ctxWithInvoicesAccess(ctx context.Context) context.Context {
	return context.WithValue(ctx, invoicesAccessKey, true)
}

func compareFloat(a, b float64, precisionDigits int) bool {
	return math.Abs(a-b) < math.Pow10(-precisionDigits)
}

func round(f float64, precisionDigits int) float64 {
	return math.Round(f*math.Pow10(precisionDigits)) / math.Pow10(precisionDigits)
}

type pair[T any] struct {
	f T
	s T
}

var forbiddenStatusConversions = []pair[pb.BillingStatus]{}

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

const instanceInstanceGroup = `
LET ig = LAST( // Find Instance instance group
    FOR node, edge, path IN 1
    INBOUND DOCUMENT(@@instances, @instance)
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner"]
    FILTER IS_SAME_COLLECTION(node, @@igs)
        RETURN node
    )
RETURN ig`

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
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	acc := requestor

	query := `FOR t IN @@invoices`
	vars := map[string]interface{}{
		"@invoices":   schema.INVOICES_COL,
		"@currencies": schema.CUR_COL,
	}

	if req.GetUuid() != "" {
		return s._HandleGetSingleInvoice(ctx, acc, req.GetUuid())
	}

	var isAdmin bool
	if s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
		isAdmin = true
	}

	if req.Account != nil {
		acc = *req.Account
		node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
		if !s.ca.HasAccess(ctx, requestor, node, access.Level_ADMIN) && requestor != req.GetAccount() {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		query += ` FILTER t.account == @acc`
		vars["acc"] = acc
	} else {
		if !s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
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
	query += ` RETURN merge(t, {uuid: t._key, currency: DOCUMENT(@@currencies, TO_STRING(t.currency.id))})`

	log.Debug("Ready to retrieve invoices", zap.String("query", query), zap.Any("vars", vars))

	cursor, err := s.db.Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to retrieve invoices", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve invoices")
	}
	defer cursor.Close()

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
		invoices = append(invoices, invoice)
	}

	log.Debug("Invoices retrieved", zap.Any("invoices", invoices))

	for _, inv := range invoices {
		if isAdmin {
			continue
		}
		if !strings.Contains(inv.GetCurrency().GetTitle(), "EUR") {
			continue
		}
		inv.Currency.Title = "EUR"
	}

	resp := connect.NewResponse(&pb.Invoices{Pool: invoices})
	return resp, nil
}

func (s *BillingServiceServer) CreateInvoice(ctx context.Context, req *connect.Request[pb.CreateInvoiceRequest]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("CreateInvoice")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	t := req.Msg.Invoice
	log.Debug("Request received", zap.Any("invoice", t), zap.String("requester", requester))
	invConf := MakeInvoicesConf(ctx, log, &s.settingsClient)
	defCurr := MakeCurrencyConf(ctx, log, &s.settingsClient).Currency

	if t.GetStatus() == pb.BillingStatus_BILLING_STATUS_UNKNOWN {
		t.Status = pb.BillingStatus_DRAFT
	}
	if t.GetType() == pb.ActionType_ACTION_TYPE_UNKNOWN {
		t.Type = pb.ActionType_NO_ACTION
	}

	rootNs := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	rootAccess := s.ca.HasAccess(ctx, requester, rootNs, access.Level_ROOT)
	if has, ok := ctx.Value(invoicesAccessKey).(bool); !rootAccess && (t.Account != requester || !ok || !has) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	if t.GetStatus() != pb.BillingStatus_DRAFT && t.GetStatus() != pb.BillingStatus_UNPAID {
		return nil, status.Error(codes.InvalidArgument, "Status can be only DRAFT and UNPAID on creation")
	}
	if t.GetTotal() < 0 {
		return nil, status.Error(codes.InvalidArgument, "Negative total")
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
	if t.Deadline != 0 && t.Deadline < time.Now().Unix() {
		return nil, status.Error(codes.InvalidArgument, "Deadline in the past")
	}

	// Rounding invoice items
	cur, err := s.currencies.Get(ctx, t.Currency.GetId())
	if err != nil {
		log.Error("Failed to get currency", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get currency")
	}
	precision := int(cur.Precision)
	var newTotal float64
	for _, item := range t.GetItems() {
		price := round(item.GetPrice(), precision)
		item.Price = price
		newTotal += price * float64(item.GetAmount())
	}
	t.Total = round(newTotal, precision)

	now := time.Now()

	strNum, num, err := s.GetNewNumber(log, unpaidInvoicesByCreatedDate, now, invConf.NewTemplate, "NONE")
	if err != nil {
		log.Error("Failed to get new number for invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get new number for invoice. "+err.Error())
	}

	acc, err := s.accounts.GetAccountOrOwnerAccountIfPresent(ctx, t.Account)
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get account")
	}
	tax := acc.GetTaxRate()

	if t.Meta == nil {
		t.Meta = make(map[string]*structpb.Value)
	}
	if t.Meta[graph.InvoiceTaxMetaKey] == nil {
		t.Meta[graph.InvoiceTaxMetaKey] = structpb.NewNumberValue(tax)
	}

	//trCtx, err := s.invoices.BeginTransaction(ctx)
	//if err != nil {
	//	log.Error("Failed to begin transaction", zap.Error(err))
	//	return nil, status.Error(codes.Internal, "Failed to begin transaction")
	//}

	t.Number = strNum
	t.Created = now.Unix()
	t.Payment = 0
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
		_ = s.invoices.AbortTransaction(ctx)
		log.Error("Failed to create invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to create invoice")
	}

	if !payments.GetGatewayCallbackValue(ctx, req.Header()) {
		if err := payments.GetPaymentGateway(acc.GetPaymentsGateway()).CreateInvoice(ctx, r.Invoice); err != nil {
			//_ = s.invoices.AbortTransaction(trCtx)
			log.Error("Failed to create invoice through gateway", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to create invoice through gateway")
		}
	}

	//_ = s.invoices.CommitTransaction(trCtx)

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

	inv, err := s.invoices.Get(ctx, r.GetUuid())
	if err != nil {
		log.Error("Failed to get invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get invoice")
	}

	resp := connect.NewResponse(inv.Invoice)
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
	var nowAfterActions int64
	newInv.Status = newStatus

	acc, err := s.accounts.Get(ctx, newInv.GetAccount())
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get account")
	}

	var strNum string
	var num int
	invConf := MakeInvoicesConf(ctx, log, &s.settingsClient)
	currConf := MakeCurrencyConf(ctx, log, &s.settingsClient)

	if newStatus == pb.BillingStatus_PAID {
		goto payment
	} else if newStatus == pb.BillingStatus_RETURNED {
		goto returning
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

	if newInv, err = s.executePostPaidActions(ctx, log, newInv, currConf.Currency); err != nil {
		log.Error("FATAL: Failed to execute postpaid actions after invoice was paid")
	}
	goto quit

returning:
	newInv.Returned = nowBeforeActions
	if newInv, err = s.executePostRefundActions(ctx, log, newInv); err != nil {
		log.Error("FATAL: Failed to execute post refund actions after invoice was returned")
	}

quit:
	nowAfterActions = time.Now().Unix()
	newInv.Processed = nowAfterActions

	_, err = s.invoices.Replace(ctx, newInv)
	if err != nil {
		log.Error("Failed to update invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to patch status. Actions should be applied, but invoice wasn't updated")
	}

	nocloud.Log(log, &elpb.Event{
		Uuid:      old.GetUuid(),
		Entity:    "Invoices",
		Action:    strings.ToLower(newStatus.String()),
		Scope:     "database",
		Rc:        0,
		Ts:        time.Now().Unix(),
		Snapshot:  &elpb.Snapshot{},
		Requestor: requester,
	})

	if !payments.GetGatewayCallbackValue(ctx, req.Header()) {
		if err := payments.GetPaymentGateway(acc.GetPaymentsGateway()).UpdateInvoice(ctx, newInv.Invoice, old.Invoice); err != nil {
			log.Error("Failed to update invoice through gateway", zap.Error(err))
		}
	}

	log.Info("Finished invoice update status")
	return connect.NewResponse(newInv.Invoice), nil
}

func (s *BillingServiceServer) PayWithBalance(ctx context.Context, r *connect.Request[pb.PayWithBalanceRequest]) (*connect.Response[pb.PayWithBalanceResponse], error) {
	log := s.log.Named("PayWithBalance")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requester))

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

	if inv.GetType() == pb.ActionType_BALANCE {
		return nil, status.Error(codes.InvalidArgument, "Can't pay top-up balance invoice with balance")
	}

	acc, err := s.accounts.Get(ctx, inv.GetAccount())
	if err != nil {
		log.Warn("Failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Account not found")
	}
	currConf := MakeCurrencyConf(ctx, log, &s.settingsClient)

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

	if balance < inv.GetTotal() {
		return nil, status.Error(codes.FailedPrecondition, "Not enough balance to perform operation")
	}

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

	tr, err := s.applyTransaction(ctx, inv.GetTotal(), inv.GetAccount(), invCurrency)
	if err != nil {
		log.Error("Failed to create transaction. INVOICE WAS PAID, ACTIONS WERE APPLIED, BUT USER HAVEN'T LOSE BALANCE", zap.Error(err))
		return nil, status.Error(codes.Internal, "Invoice was paid but still encountered an error. Error: "+err.Error())
	}
	newInv := resp.Msg
	if newInv.Transactions == nil {
		newInv.Transactions = make([]string, 0)
	}
	newInv.Transactions = append(newInv.Transactions, tr.GetUuid())
	if _, err = s.invoices.Update(ctx, &graph.Invoice{Invoice: newInv, DocumentMeta: driver.DocumentMeta{Key: newInv.GetUuid()}}); err != nil {
		log.Error("Failed to update invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Invoice was paid but still encountered an error. Error: "+err.Error())
	}

	return connect.NewResponse(&pb.PayWithBalanceResponse{Success: true}), nil
}

func (s *BillingServiceServer) payWithBalanceWhmcsInvoice(ctx context.Context, invId int64) (*connect.Response[pb.PayWithBalanceResponse], error) {
	log := s.log.Named("payWithBalanceWhmcsInvoice")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	inv, err := s.whmcsGateway.GetInvoice(ctx, int(invId))
	if err != nil {
		log.Warn("Failed to get invoice", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Invoice not found")
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
	currConf := MakeCurrencyConf(ctx, log, &s.settingsClient)

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

	if balance < float64(inv.Balance) {
		return nil, status.Error(codes.FailedPrecondition, "Not enough balance to perform operation")
	}

	if err = s.whmcsGateway.PayInvoice(ctx, int(invId)); err != nil {
		log.Error("Failed to pay invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to perform payment with balance. Error: "+err.Error())
	}

	_, err = s.applyTransaction(ctx, float64(inv.Balance), requester, invCurrency)
	if err != nil {
		log.Error("Failed to create transaction. INVOICE WAS PAID, ACTIONS WERE APPLIED, BUT USER HAVEN'T LOSE BALANCE", zap.Error(err))
		return nil, status.Error(codes.Internal, "Invoice was paid but still encountered an error. Error: "+err.Error())
	}

	return connect.NewResponse(&pb.PayWithBalanceResponse{Success: true}), nil
}

func (s *BillingServiceServer) GetInvoicesCount(ctx context.Context, r *connect.Request[pb.GetInvoicesCountRequest]) (*connect.Response[pb.GetInvoicesCountResponse], error) {
	log := s.log.Named("GetInvoicesCount")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	acc := requestor

	query := `FOR t IN @@invoices`
	vars := map[string]interface{}{
		"@invoices": schema.INVOICES_COL,
	}

	if req.Account != nil {
		acc = *req.Account
		node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
		if !s.ca.HasAccess(ctx, requestor, node, access.Level_ADMIN) && requestor != req.GetAccount() {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		query += ` FILTER t.account == @acc`
		vars["acc"] = acc
	} else {
		if !s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
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
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg.Invoice
	log.Debug("Request received", zap.Any("invoice", req), zap.String("requestor", requestor))

	if req.GetStatus() == pb.BillingStatus_BILLING_STATUS_UNKNOWN {
		req.Status = pb.BillingStatus_DRAFT
	}
	if req.GetType() == pb.ActionType_ACTION_TYPE_UNKNOWN {
		req.Type = pb.ActionType_NO_ACTION
	}

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t, err := s.invoices.Get(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to get invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get invoice")
	}
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
	if req.GetDeadline() != 0 && t.GetDeadline() != 0 {
		t.Deadline = req.GetDeadline()
	}
	if req.GetCreated() != 0 {
		t.Created = req.GetCreated()
	}

	t.Uuid = req.GetUuid()
	t.Meta = req.GetMeta()
	t.Status = req.GetStatus()
	t.Account = req.GetAccount()
	t.Total = req.GetTotal()
	t.Type = req.GetType()
	t.Items = req.GetItems()
	if t.Items == nil {
		t.Items = make([]*pb.Item, 0)
	}
	if req.Currency != nil {
		t.Currency = req.GetCurrency()
	}

	if t.Account == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing account")
	}

	// Rounding invoice items
	cur, err := s.currencies.Get(ctx, t.Currency.GetId())
	if err != nil {
		log.Error("Failed to get currency", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get currency")
	}
	precision := int(cur.Precision)
	var newTotal float64
	for _, item := range t.GetItems() {
		price := round(item.GetPrice(), precision)
		item.Price = price
		newTotal += price * float64(item.GetAmount())
	}
	t.Total = round(newTotal, precision)

	upd, err := s.invoices.Replace(ctx, t)
	if err != nil {
		log.Error("Failed to update invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update invoice")
	}

	acc, err := s.accounts.GetAccountOrOwnerAccountIfPresent(ctx, t.Account)
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get account")
	}

	if !payments.GetGatewayCallbackValue(ctx, r.Header()) {
		if err := payments.GetPaymentGateway(acc.GetPaymentsGateway()).UpdateInvoice(ctx, upd.Invoice, old); err != nil {
			log.Error("Failed to update invoice through gateway", zap.Error(err))
		}
	}

	return connect.NewResponse(t.Invoice), nil
}

func (s *BillingServiceServer) GetInvoice(ctx context.Context, r *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("GetInvoice")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("invoice", req), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t, err := s.invoices.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(t.Invoice), nil
}

func (s *BillingServiceServer) GetInvoiceSettingsTemplateExample(ctx context.Context, _req *connect.Request[pb.GetInvoiceSettingsTemplateExampleRequest]) (*connect.Response[pb.GetInvoiceSettingsTemplateExampleResponse], error) {
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

	acc, err := s.accounts.Get(ctx, inv.Account)
	if err != nil {
		log.Error("Error getting account", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Internal error")
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
	sum := req.GetSum() + req.GetSum()*tax

	ivnToCreate := &pb.Invoice{
		Deadline: time.Now().Add(72 * time.Hour).Unix(),
		Status:   pb.BillingStatus_UNPAID,
		Account:  acc.GetUuid(),
		Total:    sum,
		Type:     pb.ActionType_BALANCE,
		Items: []*pb.Item{
			{
				Amount:      1,
				Unit:        "Pcs",
				Price:       sum,
				Description: "Пополнение баланса (услуги хостинга, оплата за сервисы)",
			},
		},
		Meta: map[string]*structpb.Value{
			"creator":               structpb.NewStringValue(requester),
			graph.InvoiceTaxMetaKey: structpb.NewNumberValue(tax),
		},
	}

	if acc.Currency != nil {
		ivnToCreate.Currency = acc.Currency
	}

	return s.CreateInvoice(ctxWithInvoicesAccess(ctx), connect.NewRequest(&pb.CreateInvoiceRequest{
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

	currencyConf := MakeCurrencyConf(ctx, log, &s.settingsClient)

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
		log := log.With(zap.Any("record", r))
		if r.Product == "" {
			log.Info("Ignoring non product record")
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
	cost, err := s.promocodes.GetDiscountPriceByInstance(inst.Instance, false)
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
	cost *= rate // Convert from NCU to  account's currency
	slices.Sort(periods)
	slices.Sort(expirings)
	period := periods[len(periods)-1]
	expire := expirings[0]
	expireDate := time.Unix(expire, 0)

	var untilDate time.Time
	const monthSecs = 3600 * 24 * 30
	const daySecs = 3600 * 24
	if period == monthSecs {
		untilDate = expireDate.AddDate(0, 1, 0)
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
	cost = cost + cost*tax

	inv := &pb.Invoice{
		Status: pb.BillingStatus_UNPAID,
		Items: []*pb.Item{
			{
				Description: renewDescription,
				Amount:      1,
				Unit:        "Pcs",
				Price:       cost,
				Instance:    inst.GetUuid(),
			},
		},
		Total:    cost,
		Type:     pb.ActionType_INSTANCE_RENEWAL,
		Created:  now,
		Deadline: dueDate, // Until when invoice should be paid
		Account:  acc.GetUuid(),
		Currency: acc.Currency,
		Meta: map[string]*structpb.Value{
			"creator":               structpb.NewStringValue(requester),
			"no_discount_price":     structpb.NewStringValue(fmt.Sprintf("%.2f %s", initCost, currencyConf.Currency.GetTitle())),
			graph.InvoiceTaxMetaKey: structpb.NewNumberValue(tax),
		},
	}

	if val, ok := ctx.Value("create_as_draft").(bool); ok && val {
		inv.Status = pb.BillingStatus_DRAFT
	}

	return s.CreateInvoice(ctxWithInvoicesAccess(ctx), connect.NewRequest(&pb.CreateInvoiceRequest{
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

	resp := connect.NewResponse(&pb.Invoices{Pool: []*pb.Invoice{tr.Invoice}})

	return resp, nil
}

func (s *BillingServiceServer) executePostPaidActions(ctx context.Context, log *zap.Logger, inv *graph.Invoice, defCurr *pb.Currency) (*graph.Invoice, error) {
	tax := inv.GetMeta()[graph.InvoiceTaxMetaKey].GetNumberValue()

	c, err := s.currencies.Get(ctx, inv.GetCurrency().GetId())
	if err != nil {
		return inv, err
	}

	switch inv.GetType() {
	case pb.ActionType_BALANCE:
		total := inv.GetTotal() / (tax + 1)
		total = round(total, int(c.Precision))
		tr, err := s.applyTransaction(ctx, total, inv.GetAccount(), inv.GetCurrency())
		if err != nil {
			return inv, fmt.Errorf("failed to apply transaction: %w", err)
		}
		if inv.Transactions == nil {
			inv.Transactions = make([]string, 0)
		}
		inv.Transactions = append(inv.Transactions, tr.GetUuid())

	case pb.ActionType_INSTANCE_START:
		for _, item := range inv.GetItems() {
			i := item.GetInstance()
			log = log.With(zap.String("instance", i))
			instOld, err := s.instances.GetWithAccess(ctx, driver.NewDocumentID(schema.ACCOUNTS_COL, schema.ROOT_ACCOUNT_KEY), i)
			if err != nil {
				log.Error("Failed to get instance to start", zap.Error(err))
				continue
			}
			instOld.Uuid = instOld.Key
			// Set auto_start to true. After next driver monitoring instance will be started
			instNew := graph.Instance{
				Instance: proto.Clone(instOld.Instance).(*ipb.Instance),
			}
			cfg := instNew.Config
			cfg["auto_start"] = structpb.NewBoolValue(true)
			instNew.Config = cfg

			if err := s.instances.Update(ctx, "", instNew.Instance, instOld.Instance); err != nil {
				log.Error("Failed to update instance", zap.Error(err))
				continue
			}
			instance := instNew

			// Add balance to compensate instance first payment
			acc, err := s.instances.GetInstanceOwner(ctx, instance.GetUuid())
			if err != nil {
				return inv, err
			}
			cost, err := s.promocodes.GetDiscountPriceByInstance(instance.Instance, true, true)
			if err != nil {
				return inv, fmt.Errorf("failed to get instance price: %w", err)
			}
			cost, err = s.currencies.Convert(ctx, defCurr, acc.GetCurrency(), cost)
			if err != nil {
				log.Error("Failed to convert cost", zap.Error(err))
				return inv, fmt.Errorf("failed to convert cost: %w", err)
			}
			_, err = s.applyTransaction(ctx, cost, acc.GetUuid(), acc.GetCurrency())
			if err != nil {
				return inv, fmt.Errorf("failed to apply transaction: %w", err)
			}
		}

	case pb.ActionType_INSTANCE_RENEWAL:
		for _, item := range inv.GetItems() {
			i := item.GetInstance()
			log = log.With(zap.String("instance", i))
			if i == "" {
				log.Debug("Instance item is empty")
				continue
			}
			invokeReq := connect.NewRequest(&instancespb.InvokeRequest{
				Uuid:   i,
				Method: "free_renew",
			})
			invokeReq.Header().Set("Authorization", "Bearer "+s.rootToken)
			if _, err := s.instancesClient.Invoke(ctx, invokeReq); err != nil {
				log.Error("Failed to renew instance", zap.Error(err))
				continue
			}
		}
	}

	return inv, nil
}

func (s *BillingServiceServer) executePostRefundActions(ctx context.Context, log *zap.Logger, inv *graph.Invoice) (*graph.Invoice, error) {

	transactions := make([]string, 0)
	if inv.Transactions == nil {
		inv.Transactions = make([]string, 0)
	}
	for _, trId := range inv.GetTransactions() {
		tr, err := s.transactions.Get(ctx, trId)
		if err != nil {
			log.Error("Failed to get transaction", zap.Error(err))
			continue
		}
		if tr, err = s.applyTransaction(ctx, -tr.GetTotal(), tr.GetAccount(), tr.GetCurrency()); err != nil {
			log.Error("Failed to apply transaction", zap.Error(err))
			continue
		}
		transactions = append(transactions, tr.GetUuid())
	}
	inv.Transactions = append(inv.Transactions, transactions...)

	switch inv.GetType() {
	case pb.ActionType_INSTANCE_START:
		// TODO: maybe start returning should be done without suspending
		for _, item := range inv.GetItems() {
			id := item.GetInstance()
			invokeReq := connect.NewRequest(&instancespb.InvokeRequest{
				Uuid:   id,
				Method: "suspend",
			})
			invokeReq.Header().Set("Authorization", "Bearer "+s.rootToken)
			if _, err := s.instancesClient.Invoke(ctx, invokeReq); err != nil {
				log.Error("Failed to suspend instance", zap.Error(err))
				continue
			}
		}

	case pb.ActionType_INSTANCE_RENEWAL:
		for _, item := range inv.GetItems() {
			i := item.GetInstance()
			log = log.With(zap.String("instance", i))
			if i == "" {
				log.Debug("Instance item is empty")
				continue
			}
			invokeReq := connect.NewRequest(&instancespb.InvokeRequest{
				Uuid:   i,
				Method: "cancel_renew",
			})
			invokeReq.Header().Set("Authorization", "Bearer "+s.rootToken)
			if _, err := s.instancesClient.Invoke(ctx, invokeReq); err != nil {
				log.Error("Failed to cancel renew instance", zap.Error(err))
				continue
			}
		}
	}

	return inv, nil
}
