package billing

import (
	"connectrpc.com/connect"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	epb "github.com/slntopp/nocloud-proto/events"
	healthpb "github.com/slntopp/nocloud-proto/health"
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
	{pb.BillingStatus_PAID, pb.BillingStatus_UNPAID},
	{pb.BillingStatus_PAID, pb.BillingStatus_DRAFT},
	{pb.BillingStatus_DRAFT, pb.BillingStatus_TERMINATED},
	{pb.BillingStatus_UNPAID, pb.BillingStatus_TERMINATED},
}

func (s *BillingServiceServer) Consume(ctx context.Context) {
	log := s.log.Named("ExpiringInstancesConsumer")
init:
	log.Info("Trying to register instances expiring consumer")

	ch, err := s.rbmq.Channel()
	if err != nil {
		log.Error("Failed to open a channel", zap.Error(err))
		time.Sleep(time.Second)
		goto init
	}

	queue, _ := ch.QueueDeclare(
		"instance_expiring",
		true, false, false, true, nil,
	)

	records, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Error("Failed to register a consumer", zap.Error(err))
		time.Sleep(time.Second)
		goto init
	}

	s.ConsumerStatus.Status.Status = healthpb.Status_RUNNING
	currencyConf := MakeCurrencyConf(ctx, log)

	log.Info("Instances expiring consumer registered. Reading messages")
	for msg := range records {
		log.Debug("Received a message")
		var recs []*pb.Record
		err = json.Unmarshal(msg.Body, &recs)
		if err != nil {
			log.Error("Failed to unmarshal record", zap.Error(err))
			if err = msg.Ack(false); err != nil {
				log.Warn("Failed to Acknowledge the delivery while unmarshal message", zap.Error(err))
			}
			continue
		}
		log.Debug("Message unmarshalled", zap.Any("records", &recs))
		err := s.processExpiringRecords(ctx, recs, currencyConf)
		if err != nil {
			log.Error("Failed to process record", zap.Error(err))
		}
		if err = msg.Ack(false); err != nil {
			log.Warn("Failed to Acknowledge the delivery while unmarshal message", zap.Error(err))
		}
		continue
	}
}

const instanceOwner = `
LET account = LAST( // Find Instance owner Account
    FOR node, edge, path IN 4
    INBOUND DOCUMENT(@@instances, @instance)
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@accounts)
        RETURN node
    )`

func (s *BillingServiceServer) processExpiringRecords(ctx context.Context, recs []*pb.Record, currency CurrencyConf) error {

	var i *graph.Instance
	var plan *graph.BillingPlan
	var sum float64
	for _, rec := range recs {
		var err error
		i, err = s.instances.Get(ctx, rec.GetInstance())
		if err != nil {
			return err
		}
		plan, err = s.plans.Get(ctx, i.GetBillingPlan())
		if err != nil {
			return err
		}
		if product, ok := plan.GetProducts()[rec.Product]; ok {
			sum += product.Price * rec.Total
		}
		// Scan each resource to find presented in current record. TODO: optimize
		for _, res := range plan.GetResources() {
			if res.Key == rec.Resource {
				sum += res.Price * rec.Total
			}
		}
	}

	if plan == nil {
		return errors.New("got nil plan")
	}

	if i == nil {
		return errors.New("got nil instance")
	}

	if sum == 0 {
		return errors.New("payment sum is zero")
	}

	// Find owner account
	cur, err := s.db.Query(ctx, instanceOwner, map[string]interface{}{
		"instance":    i.GetUuid(),
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"@instances":  schema.INSTANCES_COL,
		"@accounts":   schema.ACCOUNTS_COL,
	})
	if err != nil {
		return err
	}
	var acc graph.Account
	_, err = cur.ReadDocument(ctx, &acc)
	if err != nil {
		return err
	}

	if acc.Currency == nil {
		acc.Currency = currency.Currency
	}
	rate, err := s.currencies.GetExchangeRate(ctx, acc.Currency, currency.Currency)
	if err != nil {
		return err
	}

	inv := &pb.Invoice{
		Exec:    time.Now().Add(time.Duration(plan.GetProducts()[i.GetProduct()].GetPeriod()) * time.Second).Unix(),
		Status:  pb.BillingStatus_UNPAID,
		Total:   sum * rate,
		Created: time.Now().Unix(),
		Type:    pb.ActionType_INSTANCE_RENEWAL,
		Items: []*pb.Item{
			{Title: i.Title + " renewal", Amount: int64(sum * rate), Instance: i.GetUuid()},
		},
		Account:  acc.GetUuid(),
		Currency: acc.Currency,
	}

	_, err = s.CreateInvoice(ctx, connect.NewRequest(inv))
	return err
}

func (s *BillingServiceServer) GetInvoices(ctx context.Context, r *connect.Request[pb.GetInvoicesRequest]) (*connect.Response[pb.Invoices], error) {
	log := s.log.Named("GetTransactions")
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

	if req.Service != nil {
		service := *req.Service
		node := driver.NewDocumentID(schema.SERVICES_COL, service)
		if !graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN) {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		if req.Account == nil {
			query += ` FILTER t.service == @service`
		} else {
			query += ` && t.service == @service`
		}
		vars["service"] = service
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "exec" || key == "total" || key == "proc" || key == "created" {
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

	log.Debug("Ready to retrieve transactions", zap.String("query", query), zap.Any("vars", vars))

	cursor, err := s.db.Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to retrieve transactions", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve transactions")
	}
	defer cursor.Close()

	var transactions []*pb.Invoice
	for {
		transaction := &pb.Invoice{}
		meta, err := cursor.ReadDocument(ctx, transaction)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			log.Error("Failed to retrieve transactions", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to retrieve transactions")
		}
		transaction.Uuid = meta.Key
		transactions = append(transactions, transaction)
	}

	log.Debug("Transactions retrieved", zap.Any("transactions", transactions))
	resp := connect.NewResponse(&pb.Invoices{Pool: transactions})
	return resp, nil
}

func (s *BillingServiceServer) CreateInvoice(ctx context.Context, req *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("CreateTransaction")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	t := req.Msg
	log.Debug("Request received", zap.Any("transaction", t), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	if t.GetTotal() == 0 {
		return nil, status.Error(codes.InvalidArgument, "Zero total")
	}
	if t.Account == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing account")
	}
	if t.GetType() == pb.ActionType_BALANCE && t.GetTotal() >= 0 {
		return nil, status.Error(codes.InvalidArgument, "Positive total in balance invoice")
	}

	newTr, err := s.CreateTransaction(ctx, connect.NewRequest(&pb.Transaction{
		Priority: pb.Priority_NORMAL,
		Account:  t.GetAccount(),
		Total:    t.GetTotal(),
		Exec:     0,
	}))
	if err != nil {
		log.Error("Failed to create transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to create transaction")
	}

	t.Transactions = []string{newTr.Msg.Uuid}
	t.Created = time.Now().Unix()
	r, err := s.invoices.Create(ctx, t)
	if err != nil {
		log.Error("Failed to create invoice", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to create invoice")
	}

	eventsClient.Publish(ctx, &epb.Event{
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

	// Cannot rollback from cancelled and terminated
	if oldStatus == pb.BillingStatus_CANCELED || oldStatus == pb.BillingStatus_TERMINATED {
		return nil, status.Error(codes.InvalidArgument, "Cannot rollback from cancelled and terminated statuses")
	}
	if slices.Contains(forbiddenStatusConversions, pair[pb.BillingStatus]{oldStatus, newStatus}) {
		return nil, status.Error(codes.InvalidArgument, "Cannot convert from "+oldStatus.String()+" to "+newStatus.String())
	}

	// Update invoice's transactions to perform payment
	if newStatus == pb.BillingStatus_PAID {
		for _, trId := range old.GetTransactions() {
			tr, err := s.transactions.Get(ctx, trId)
			if err != nil {
				log.Error("Failed to get transaction", zap.Error(err))
				return nil, status.Error(codes.Internal, "Failed to get transaction")
			}
			tr.Priority = pb.Priority_URGENT
			_, err = s.UpdateTransaction(ctx, connect.NewRequest(tr))
			if err != nil {
				log.Error("Failed to update transaction", zap.Error(err))
				return nil, status.Error(codes.Internal, "Failed to update transaction")
			}
		}
	}

	now := time.Now().Unix()
	patch := map[string]interface{}{
		"status":    newStatus,
		"proc":      now,
		"processed": true,
	}
	err = s.invoices.Patch(ctx, t.GetUuid(), patch)
	if err != nil {
		log.Error("Failed to update status", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update status")
	}
	old.Status = newStatus
	old.Proc = now
	old.Processed = true

	_, err = s.accounts.Get(ctx, old.GetAccount())
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get account")
	}

	var resp *connect.Response[pb.Invoice]

	if newStatus == pb.BillingStatus_PAID {
		goto payment
	} else if newStatus == pb.BillingStatus_TERMINATED {
		goto termination
	} else {
		resp = connect.NewResponse(old)
		return resp, nil
	}

payment:

	switch old.GetType() {
	case pb.ActionType_BALANCE:
		// Updated transactions raised account's balance

	case pb.ActionType_INSTANCE_CREATION:
		for _, item := range old.GetItems() {
			i := item.GetInstance()
			instOld, err := s.instances.Get(ctx, i)
			if err != nil {
				log.Warn("Failed to get instance to start", zap.Error(err))
				continue
			}
			instNew := graph.Instance{
				Instance:     proto.Clone(instOld.Instance).(*ipb.Instance),
				DocumentMeta: instOld.DocumentMeta,
			}
			cfg := instNew.Config
			cfg["auto_start"] = structpb.NewBoolValue(true)
			instNew.Config = cfg
			if err := s.instances.Update(ctx, "", instNew.Instance, instOld.Instance); err != nil {
				log.Warn("Failed to update instance auto_start", zap.Error(err))
				continue
			}
			log.Info("Updated auto_start for instances after invoice was paid")
		}

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
				return nil, err
			}
		}
		log.Info("Renewed instance after invoice was paid")
	}
	resp = connect.NewResponse(old)
	return resp, nil

termination:
	// TODO: undo action after invoice termination

	resp = connect.NewResponse(old)
	return resp, nil
}

func (s *BillingServiceServer) GetInvoicesCount(ctx context.Context, r *connect.Request[pb.GetInvoicesCountRequest]) (*connect.Response[pb.GetInvoicesCountResponse], error) {
	log := s.log.Named("GetTransactionsCount")
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
			if key == "exec" || key == "total" || key == "proc" || key == "created" {
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

	if req.Service != nil {
		service := *req.Service
		node := driver.NewDocumentID(schema.SERVICES_COL, service)
		if !graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN) {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		if req.Account == nil {
			query += ` FILTER t.service == @service`
		} else {
			query += ` && t.service == @service`
		}
		vars["service"] = service
	}

	query += ` RETURN t`

	log.Debug("Ready to retrieve transactions", zap.String("query", query), zap.Any("vars", vars))

	queryContext := driver.WithQueryCount(ctx)

	cursor, err := s.db.Query(queryContext, query, vars)
	if err != nil {
		log.Error("Failed to retrieve transactions", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve transactions")
	}
	defer cursor.Close()

	log.Info("transactions count", zap.Int64("count", cursor.Count()))

	resp := connect.NewResponse(&pb.GetInvoicesCountResponse{
		Total: uint64(cursor.Count()),
	})

	return resp, nil
}

func (s *BillingServiceServer) UpdateInvoice(ctx context.Context, r *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("UpdateTransaction")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("transaction", req), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t, err := s.invoices.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}

	if err != nil {
		log.Error("Failed to get transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get transaction")
	}

	exec := t.GetExec()
	if exec != 0 {
		log.Error("Transaction has exec timestamp")
		return nil, status.Error(codes.Internal, "Transaction has exec timestamp")
	}
	if req.GetExec() != 0 {
		t.Exec = req.GetExec()
	}
	t.Uuid = req.GetUuid()
	t.Meta = req.GetMeta()
	t.Status = req.GetStatus()

	_, err = s.invoices.Update(ctx, t)
	if err != nil {
		log.Error("Failed to update transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update transaction")
	}

	return connect.NewResponse(t), nil
}

func (s *BillingServiceServer) GetInvoice(ctx context.Context, r *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("UpdateTransaction")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("transaction", req), zap.String("requestor", requestor))

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
		return nil, status.Error(codes.NotFound, "Transaction doesn't exist")
	}

	ok := graph.HasAccess(ctx, s.db, acc, driver.NewDocumentID(schema.ACCOUNTS_COL, tr.Account), access.Level_ADMIN)

	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enoguh Access Rights")
	}

	resp := connect.NewResponse(&pb.Invoices{Pool: []*pb.Invoice{tr}})

	return resp, nil
}
