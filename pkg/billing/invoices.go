package billing

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	epb "github.com/slntopp/nocloud-proto/events"
	ipb "github.com/slntopp/nocloud-proto/instances"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"slices"
	"time"
)

type pair[T any] struct {
	f T
	s T
}

var forbiddenStatusConversions = []pair[pb.BillingStatus]{
	// From DRAFT
	{pb.BillingStatus_DRAFT, pb.BillingStatus_RETURNED},
	// From UNPAID
	{pb.BillingStatus_UNPAID, pb.BillingStatus_RETURNED},
	// From PAID
	{pb.BillingStatus_PAID, pb.BillingStatus_DRAFT},
	{pb.BillingStatus_PAID, pb.BillingStatus_UNPAID},
	{pb.BillingStatus_PAID, pb.BillingStatus_CANCELED},
	// From CANCELLED  (All forbidden)
	// From RETURNED   (All forbidden)
	// From TERMINATED (All forbidden)
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

func (s *BillingServiceServer) GetInvoices(ctx context.Context, r *connect.Request[pb.GetInvoicesRequest]) (*connect.Response[pb.Invoices], error) {
	log := s.log.Named("GetInvoice")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	acc := requestor

	query := `FOR t IN @@invoices`
	vars := map[string]interface{}{
		"@invoices": schema.INVOICES_COL,
	}

	if req.GetUuid() != "" {
		return s._HandleGetSingleInvoice(ctx, acc, req.GetUuid())
	}

	if req.Account != nil {
		acc = *req.Account
		node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
		if !graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN) {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		query += ` FILTER t.account == @acc`
		vars["acc"] = acc
	} else {
		if acc != schema.ROOT_ACCOUNT_KEY {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "exec" || key == "total" || key == "processed" || key == "created" {
				values := value.GetStructValue().AsMap()
				if val, ok := values["from"]; ok {
					from := val.(float64)
					query += fmt.Sprintf(` FILTER record["%s"] >= %f`, key, from)
				}

				if val, ok := values["to"]; ok {
					to := val.(float64)
					query += fmt.Sprintf(` FILTER record["%s"] <= %f`, key, to)
				}
			} else {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER record["%s"] in @%s`, key, key)
				vars[key] = values
			}
		}
	}

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT t.%s %s`
		field, sort := req.GetField(), req.GetSort()

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
	query += ` RETURN merge(t, {uuid: t._key})`

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
	resp := connect.NewResponse(&pb.Invoices{Pool: invoices})
	return resp, nil
}

func (s *BillingServiceServer) CreateInvoice(ctx context.Context, req *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("CreateInvoice")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	t := req.Msg
	log.Debug("Request received", zap.Any("invoice", t), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	if t.GetStatus() != pb.BillingStatus_DRAFT && t.GetStatus() != pb.BillingStatus_UNPAID {
		return nil, status.Error(codes.InvalidArgument, "Status can be only DRAFT and UNPAID on creation")
	}
	if t.GetTotal() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Zero or negative total")
	}
	if t.Account == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing account")
	}
	if t.Transactions == nil {
		t.Transactions = []string{}
	}
	if t.Deadline != 0 && t.Deadline < time.Now().Unix() {
		return nil, status.Error(codes.InvalidArgument, "Deadline in the past")
	}
	if len(req.Msg.GetItems()) > 0 {
		sum := float32(0)
		for _, item := range req.Msg.GetItems() {
			sum += item.GetAmount()
			if item.Instance == "" {
				return nil, status.Error(codes.InvalidArgument, "Missing instance in item")
			}
		}
		if float64(sum) != t.GetTotal() {
			return nil, status.Error(codes.InvalidArgument, "Sum of existing items not equals to total")
		}
	}

	// Create transaction if it's balance deposit
	if t.GetType() == pb.ActionType_BALANCE {
		var transactionTotal = t.GetTotal()
		transactionTotal *= -1

		// Convert invoice's currency to default currency(according to how creating transaction works)
		defCurr := MakeCurrencyConf(ctx, log).Currency
		rate, err := s.currencies.GetExchangeRate(ctx, t.GetCurrency(), defCurr)
		if err != nil {
			log.Error("Failed to get exchange rate", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to get exchange rate")
		}

		newTr, err := s.CreateTransaction(ctx, connect.NewRequest(&pb.Transaction{
			Priority: pb.Priority_NORMAL,
			Account:  t.GetAccount(),
			Currency: defCurr,
			Total:    transactionTotal * rate,
			Exec:     0,
		}))
		if err != nil {
			log.Error("Failed to create transaction", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to create transaction for invoice")
		}
		t.Transactions = []string{newTr.Msg.Uuid}
	}

	t.Created = time.Now().Unix()
	t.Exec = 0
	t.Processed = 0
	r, err := s.invoices.Create(ctx, t)
	if err != nil {
		log.Error("Failed to create invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to create invoice")
	}

	_, _ = eventsClient.Publish(ctx, &epb.Event{
		Type: "email",
		Uuid: t.GetAccount(),
		Key:  "invoice_created",
	})

	resp := connect.NewResponse(r)
	return resp, nil
}

func (s *BillingServiceServer) UpdateInvoiceStatus(ctx context.Context, req *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("UpdateStatus")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	t := req.Msg
	log.Debug("UpdateStatus request received")

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	old, err := s.invoices.Get(ctx, t.GetUuid())
	if err != nil {
		log.Error("Failed to get invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get invoice")
	}
	newStatus := t.GetStatus()
	oldStatus := old.GetStatus()

	if oldStatus == newStatus {
		return nil, status.Error(codes.InvalidArgument, "Same status")
	}
	// Cannot rollback from cancelled, terminated or returned statuses
	if oldStatus == pb.BillingStatus_CANCELED ||
		oldStatus == pb.BillingStatus_TERMINATED ||
		oldStatus == pb.BillingStatus_RETURNED {
		return nil, status.Error(codes.InvalidArgument, "Cannot rollback from cancelled, terminated or returned statuses")
	}
	if slices.Contains(forbiddenStatusConversions, pair[pb.BillingStatus]{oldStatus, newStatus}) {
		return nil, status.Error(codes.InvalidArgument, "Cannot convert from "+oldStatus.String()+" to "+newStatus.String())
	}

	nowBeforeActions := time.Now().Unix()
	patch := map[string]interface{}{
		"status": newStatus,
	}
	old.Status = newStatus

	_, err = s.accounts.Get(ctx, old.GetAccount())
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get account")
	}

	transactions := old.GetTransactions()
	var resp *connect.Response[pb.Invoice]

	if newStatus == pb.BillingStatus_PAID {
		goto payment
	} else if newStatus == pb.BillingStatus_RETURNED {
		goto returning
	} else {
		goto quit
	}

payment:
	log.Info("Starting invoice payment", zap.String("invoice", old.GetUuid()))
	patch["exec"] = nowBeforeActions
	old.Exec = nowBeforeActions

	log.Debug("Updating transactions to perform payment.")
	for _, trId := range old.GetTransactions() {
		tr, err := s.transactions.Get(ctx, trId)
		if err != nil {
			log.Error("Failed to get transaction", zap.Error(err))
			continue
		}
		tr.Exec = nowBeforeActions // Setting transaction process time. Should trigger transaction process
		_, err = s.UpdateTransaction(ctx, connect.NewRequest(tr))
		if err != nil {
			log.Error("Failed to update transaction", zap.Error(err))
			continue
		}
	}
	log.Debug("Transactions were updated and processed.")

	// BALANCE action was processed after transaction was processed
	// NO_ACTION action don't need any processing
	switch old.GetType() {
	case pb.ActionType_INSTANCE_START:
		log.Debug("Paid action: instance start")
		for _, item := range old.GetItems() {
			i := item.GetInstance()
			instOld, err := s.instances.Get(ctx, i)
			if err != nil {
				log.Warn("Failed to get instance to start", zap.Error(err), zap.String("instance", i))
				continue
			}
			// Set auto_start to true. After next driver monitoring instance will be started
			instNew := graph.Instance{
				Instance: proto.Clone(instOld.Instance).(*ipb.Instance),
			}
			cfg := instNew.Config
			cfg["auto_start"] = structpb.NewBoolValue(true)
			instNew.Config = cfg
			if err := s.instances.Update(ctx, "", instNew.Instance, instOld.Instance); err != nil {
				log.Warn("Failed to update instance", zap.Error(err), zap.String("instance", i))
				continue
			}
			log.Debug("Successfully updated auto_start for instance", zap.String("instance", i))
		}
		log.Info("Updated auto_start for instances after invoice was paid")

	case pb.ActionType_INSTANCE_RENEWAL:
		log.Debug("Paid action: instance renewal")
		for _, item := range old.GetItems() {
			i := item.GetInstance()
			instOld, err := s.instances.Get(ctx, i)
			if err != nil {
				log.Error("Failed to get instance to renew", zap.Error(err))
				continue
			}
			instNew := proto.Clone(instOld.Instance).(*ipb.Instance)
			plan, err := s.plans.Get(ctx, &pb.Plan{Uuid: instOld.GetBillingPlan().GetUuid()})
			if err != nil {
				log.Error("Failed to get plan", zap.Error(err))
				continue
			}

			// Update every *_last_monitoring data to put aside instance billing
			// TODO: should work correctly but need to make 100% sure
			var last, next int64
			for _, resource := range plan.GetResources() {
				_, ok := instOld.Data[resource.Key+"_last_monitoring"]
				if ok {
					last = int64(instOld.Data[resource.Key+"_last_monitoring"].GetNumberValue())
					instNew.Data[resource.Key+"_last_monitoring"] = structpb.NewNumberValue(float64(last + resource.GetPeriod()))
				}
				_, ok = instOld.Data[resource.Key+"_next_payment_date"]
				if ok {
					next = int64(instOld.Data[resource.Key+"_next_payment_date"].GetNumberValue())
					instNew.Data[resource.Key+"_next_payment_date"] = structpb.NewNumberValue(float64(next + resource.GetPeriod()))
				}
			}
			product := plan.GetProducts()[instOld.GetProduct()]
			_, ok := instOld.Data["last_monitoring"]
			if ok {
				last = int64(instOld.Data["last_monitoring"].GetNumberValue())
				instNew.Data["last_monitoring"] = structpb.NewNumberValue(float64(last + product.GetPeriod()))
			}
			_, ok = instOld.Data["next_payment_date"]
			if ok {
				next = int64(instOld.Data["next_payment_date"].GetNumberValue())
				instNew.Data["next_payment_date"] = structpb.NewNumberValue(float64(next + product.GetPeriod()))
			}

			if err := s.instances.Update(ctx, "", instNew, instOld.Instance); err != nil {
				log.Error("Failed to update instance", zap.Error(err))
				continue
			}
			log.Debug("Renewed instance", zap.String("instance", i))
		}
		log.Info("Renewed instances after invoice was paid")
	}

	patch["processed"] = time.Now().Unix()
	old.Processed = int64(patch["processed"].(float64))
	goto quit

returning:
	log.Info("Starting invoice returning", zap.String("invoice", old.GetUuid()))
	// Create same amount of transactions but rewert their total
	// Make them urgent and set exec time to apply them to account's balance immediately
	log.Debug("Creating rewert transactions")
	for _, trId := range old.GetTransactions() {
		tr, err := s.transactions.Get(ctx, trId)
		if err != nil {
			log.Error("Failed to get transaction", zap.Error(err))
			continue
		}
		tr.Uuid = ""
		tr.Priority = pb.Priority_URGENT
		tr.Exec = nowBeforeActions
		tr.Records = nil
		tr.Created = nowBeforeActions
		tr.Total = tr.Total * -1
		t, err := s.CreateTransaction(ctx, connect.NewRequest(tr))
		if err != nil {
			log.Error("Failed to create rewert transaction", zap.Error(err))
			continue
		}
		if t.Msg.GetUuid() == "" {
			log.Error("Created transaction uuid is empty")
			continue
		}
		transactions = append(transactions, t.Msg.GetUuid())
	}
	log.Debug("Patching invoice with rewert transactions")
	if err = s.invoices.Patch(ctx, old.GetUuid(), map[string]interface{}{"transactions": transactions}); err != nil {
		log.Error("Failed to patch invoice with rewert transactions", zap.Error(err))
	}
	log.Debug("Ended revert transactions creation")
	log.Debug("Starting action termination")
	// BALANCE was reverted on revert transactions
	// NO_ACTION action don't need any reverting actions
	switch old.GetType() {
	case pb.ActionType_INSTANCE_START:
		// TODO: research how to properly stop or pause already running instance
		log.Debug("NOT IMPLEMENTED: Returning of started instance")

	case pb.ActionType_INSTANCE_RENEWAL:
		for _, item := range old.GetItems() {
			i := item.GetInstance()
			instOld, err := s.instances.Get(ctx, i)
			if err != nil {
				log.Error("Failed to get instance to renew", zap.Error(err))
				continue
			}
			instNew := proto.Clone(instOld.Instance).(*ipb.Instance)
			plan, err := s.plans.Get(ctx, &pb.Plan{Uuid: instOld.GetBillingPlan().GetUuid()})
			if err != nil {
				log.Error("Failed to get plan", zap.Error(err))
				continue
			}

			// Update every *_last_monitoring data to cancel renewal
			// TODO: make sure this works correctly
			var last, next int64
			for _, resource := range plan.GetResources() {
				_, ok := instOld.Data[resource.Key+"_last_monitoring"]
				if ok {
					last = int64(instOld.Data[resource.Key+"_last_monitoring"].GetNumberValue())
					instNew.Data[resource.Key+"_last_monitoring"] = structpb.NewNumberValue(float64(last - resource.GetPeriod()))
				}
				_, ok = instOld.Data[resource.Key+"_next_payment_date"]
				if ok {
					next = int64(instOld.Data[resource.Key+"_next_payment_date"].GetNumberValue())
					instNew.Data[resource.Key+"_next_payment_date"] = structpb.NewNumberValue(float64(next - resource.GetPeriod()))
				}
			}
			product := plan.GetProducts()[instOld.GetProduct()]
			_, ok := instOld.Data["last_monitoring"]
			if ok {
				last = int64(instOld.Data["last_monitoring"].GetNumberValue())
				instNew.Data["last_monitoring"] = structpb.NewNumberValue(float64(last - product.GetPeriod()))
			}
			_, ok = instOld.Data["next_payment_date"]
			if ok {
				next = int64(instOld.Data["next_payment_date"].GetNumberValue())
				instNew.Data["next_payment_date"] = structpb.NewNumberValue(float64(next - product.GetPeriod()))
			}

			if err := s.instances.Update(ctx, "", instNew, instOld.Instance); err != nil {
				log.Error("Failed to update instance", zap.Error(err))
				continue
			}

			log.Debug("Reverted instance renewal", zap.String("instance", i))
		}
	}
	log.Debug("Finished invoice returning")
	goto quit

quit:
	err = s.invoices.Patch(ctx, t.GetUuid(), patch)
	if err != nil {
		log.Error("Failed to update status", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to patch status. Actions should be applied, but invoice wasn't updated")
	}

	resp = connect.NewResponse(old)
	return resp, nil
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

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "exec" || key == "total" || key == "processed" || key == "created" {
				values := value.GetStructValue().AsMap()
				if val, ok := values["from"]; ok {
					from := val.(float64)
					query += fmt.Sprintf(` FILTER record["%s"] >= %f`, key, from)
				}

				if val, ok := values["to"]; ok {
					to := val.(float64)
					query += fmt.Sprintf(` FILTER record["%s"] <= %f`, key, to)
				}
			} else {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER record["%s"] in @%s`, key, key)
				vars[key] = values
			}
		}
	}

	if req.Account != nil {
		acc = *req.Account
		node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
		if !graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN) {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		query += ` FILTER t.account == @acc`
		vars["acc"] = acc
	} else {
		if acc != schema.ROOT_ACCOUNT_KEY {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
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

func (s *BillingServiceServer) UpdateInvoice(ctx context.Context, r *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("UpdateInvoice")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("invoice", req), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t, err := s.invoices.Get(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to get invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get invoice")
	}

	exec := t.GetExec()
	if exec != 0 {
		log.Error("Invoice has exec timestamp")
		return nil, status.Error(codes.Internal, "Invoice has exec timestamp")
	}
	if req.GetExec() != 0 {
		t.Exec = req.GetExec()
	}
	t.Uuid = req.GetUuid()
	t.Meta = req.GetMeta()
	t.Status = req.GetStatus()
	t.Processed = req.GetProcessed()
	t.Account = req.GetAccount()
	t.Deadline = req.GetDeadline()
	t.Total = req.GetTotal()
	t.Currency = req.GetCurrency()
	t.Type = req.GetType()
	t.Items = req.GetItems()
	if req.Transactions != nil {
		t.Transactions = req.Transactions
	}

	if t.Type == pb.ActionType_BALANCE {
		var transactionTotal = t.GetTotal()
		transactionTotal *= -1

		// Convert invoice's currency to default currency(according to how creating transaction works)
		defCurr := MakeCurrencyConf(ctx, log).Currency
		rate, err := s.currencies.GetExchangeRate(ctx, t.GetCurrency(), defCurr)
		if err != nil {
			log.Error("Failed to get exchange rate", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to get exchange rate")
		}

		newTr, err := s.CreateTransaction(ctx, connect.NewRequest(&pb.Transaction{
			Priority: pb.Priority_NORMAL,
			Account:  t.GetAccount(),
			Currency: defCurr,
			Total:    transactionTotal * rate,
			Exec:     0,
		}))
		if err != nil {
			log.Error("Failed to create transaction", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to create transaction for invoice")
		}
		t.Transactions = []string{newTr.Msg.Uuid}
	}

	if len(t.GetItems()) > 0 {
		sum := float32(0)
		for _, item := range t.Items {
			sum += item.GetAmount()
			if item.Instance == "" {
				return nil, status.Error(codes.InvalidArgument, "Missing instance in item")
			}
		}
		if float64(sum) != t.Total {
			return nil, status.Error(codes.InvalidArgument, "Sum of items does not match total")
		}
	}

	if t.Account == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing account")
	}

	_, err = s.invoices.Update(ctx, t)
	if err != nil {
		log.Error("Failed to update invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update invoice")
	}

	return connect.NewResponse(t), nil
}

func (s *BillingServiceServer) GetInvoice(ctx context.Context, r *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("GetInvoice")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("invoice", req), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t, err := s.invoices.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(t), nil
}

func (s *BillingServiceServer) _HandleGetSingleInvoice(ctx context.Context, acc, uuid string) (*connect.Response[pb.Invoices], error) {
	tr, err := s.invoices.Get(ctx, uuid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Invoice doesn't exist")
	}

	ok := graph.HasAccess(ctx, s.db, acc, driver.NewDocumentID(schema.ACCOUNTS_COL, tr.Account), access.Level_ADMIN)

	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enoguh Access Rights")
	}

	resp := connect.NewResponse(&pb.Invoices{Pool: []*pb.Invoice{tr}})

	return resp, nil
}

//func (s *BillingServiceServer) Consume(ctx context.Context) {
//	log := s.log.Named("ExpiringInstancesConsumer")
//init:
//	log.Info("Trying to register instances expiring consumer")
//
//	ch, err := s.rbmq.Channel()
//	if err != nil {
//		log.Error("Failed to open a channel", zap.Error(err))
//		time.Sleep(time.Second)
//		goto init
//	}
//
//	queue, _ := ch.QueueDeclare(
//		"instance_expiring",
//		true, false, false, true, nil,
//	)
//
//	records, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
//	if err != nil {
//		log.Error("Failed to register a consumer", zap.Error(err))
//		time.Sleep(time.Second)
//		goto init
//	}
//
//	s.ConsumerStatus.Status.Status = healthpb.Status_RUNNING
//	currencyConf := MakeCurrencyConf(ctx, log)
//
//	log.Info("Instances expiring consumer registered. Reading messages")
//	for msg := range records {
//		log.Debug("Received a message")
//		var recs []*pb.Record
//		err = json.Unmarshal(msg.Body, &recs)
//		if err != nil {
//			log.Error("Failed to unmarshal record", zap.Error(err))
//			if err = msg.Ack(false); err != nil {
//				log.Warn("Failed to Acknowledge the delivery while unmarshal message", zap.Error(err))
//			}
//			continue
//		}
//		log.Debug("Message unmarshalled", zap.Any("records", &recs))
//		err := s.processExpiringRecords(ctx, recs, currencyConf)
//		if err != nil {
//			log.Error("Failed to process record", zap.Error(err))
//		}
//		if err = msg.Ack(false); err != nil {
//			log.Warn("Failed to Acknowledge the delivery while unmarshal message", zap.Error(err))
//		}
//		continue
//	}
//}

//func (s *BillingServiceServer) processExpiringRecords(ctx context.Context, recs []*pb.Record, currency CurrencyConf) error {
//
//	var i *graph.Instance
//	var plan *graph.BillingPlan
//	var sum float64
//	for _, rec := range recs {
//		var err error
//		i, err = s.instances.Get(ctx, rec.GetInstance())
//		if err != nil {
//			return err
//		}
//		plan, err = s.plans.Get(ctx, i.GetBillingPlan())
//		if err != nil {
//			return err
//		}
//		if product, ok := plan.GetProducts()[rec.Product]; ok {
//			sum += product.Price * rec.Total
//		}
//		// Scan each resource to find presented in current record. TODO: optimize
//		for _, res := range plan.GetResources() {
//			if res.Key == rec.Resource {
//				sum += res.Price * rec.Total
//			}
//		}
//	}
//
//	if plan == nil {
//		return errors.New("got nil plan")
//	}

//if i == nil {
//	return errors.New("got nil instance")
//}
//
//if sum == 0 {
//	return errors.New("payment sum is zero")
//}
//
//// Make sure we're not gonna send invoice twice for the same notification
//// If it past less time than payment_period / 10 then it's considered as previous renew notification
//// payment_period / 10 --- same in ione driver
//now := time.Now().Unix()
//lastInvoiceData, ok := i.Data["last_renew_invoice"]
//if ok {
//	period := plan.GetProducts()[i.GetProduct()].GetPeriod()
//	lastInvoice := int64(lastInvoiceData.GetNumberValue())
//	if now-lastInvoice <= period/10 {
//		s.log.Info("INFO: Skipping renew invoice issuing.", zap.Int64("diff from last notify", time.Now().Unix()-lastInvoice))
//		return nil
//	}
//}
//
//// Find owner account
//cur, err := s.db.Query(ctx, instanceOwner, map[string]interface{}{
//	"instance":    i.GetUuid(),
//	"permissions": schema.PERMISSIONS_GRAPH.Name,
//	"@instances":  schema.INSTANCES_COL,
//	"@accounts":   schema.ACCOUNTS_COL,
//})
//if err != nil {
//	return err
//}
//var acc graph.Account
//_, err = cur.ReadDocument(ctx, &acc)
//if err != nil {
//	return err
//}

//	if acc.Currency == nil {
//		acc.Currency = currency.Currency
//	}
//	rate, err := s.currencies.GetExchangeRate(ctx, currency.Currency, acc.Currency)
//	if err != nil {
//		return err
//	}
//
//	newInst := proto.Clone(i.Instance).(*ipb.Instance)
//	newInst.Data["last_renew_invoice"] = structpb.NewNumberValue(float64(now))
//	if err := s.instances.Update(ctx, "", newInst, i.Instance); err != nil {
//		s.log.Error("Failed to update instance last_renew_invoice. Skipping invoice creation", zap.Error(err))
//		return err
//	}
//
//	inv := &pb.Invoice{
//		Exec:    time.Now().Add(time.Duration(plan.GetProducts()[i.GetProduct()].GetPeriod()) * time.Second).Unix(),
//		Status:  pb.BillingStatus_UNPAID,
//		Total:   sum * rate,
//		Created: now,
//		Type:    pb.ActionType_INSTANCE_RENEWAL,
//		Items: []*pb.Item{
//			{Title: i.Title + " renewal", Amount: int64(sum * rate), Instance: i.GetUuid()},
//		},
//		Account:  acc.GetUuid(),
//		Currency: acc.Currency,
//	}
//
//	_, err = s.CreateInvoice(context.WithValue(ctx, nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY), connect.NewRequest(inv))
//	return err
//}
