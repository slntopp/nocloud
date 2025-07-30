/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package billing

import (
	"context"
	"fmt"
	spb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	ps "github.com/slntopp/nocloud/pkg/pubsub"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	healthpb "github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

var (
	RabbitMQConn string
)

type RecordsServiceServer struct {
	log        *zap.Logger
	rbmq       rabbitmq.Connection
	records    graph.RecordsController
	plans      graph.BillingPlansController
	instances  graph.InstancesController
	addons     graph.AddonsController
	promocodes graph.PromocodesController
	ca         graph.CommonActionsController

	settingsClient spb.SettingsServiceClient

	db driver.Database

	ConsumerStatus *healthpb.RoutineStatus
}

func NewRecordsServiceServer(logger *zap.Logger, conn rabbitmq.Connection, db driver.Database, settingsClient spb.SettingsServiceClient,
	records graph.RecordsController, plans graph.BillingPlansController, instances graph.InstancesController, addons graph.AddonsController,
	promocodes graph.PromocodesController, ca graph.CommonActionsController) *RecordsServiceServer {
	log := logger.Named("RecordsService")

	return &RecordsServiceServer{
		log:        log,
		rbmq:       conn,
		records:    records,
		plans:      plans,
		instances:  instances,
		addons:     addons,
		promocodes: promocodes,
		ca:         ca,

		settingsClient: settingsClient,

		db: db,
		ConsumerStatus: &healthpb.RoutineStatus{
			Routine: "Records Consumer",
			Status: &healthpb.ServingStatus{
				Service: "Billing Machine",
				Status:  healthpb.Status_STOPPED,
			},
		},
	}
}

func (s *RecordsServiceServer) Consume(_ctx context.Context, pubsub *ps.PubSub[*pb.Record], wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.WithoutCancel(_ctx)
	log := s.log.Named("RecordsConsumer")

start:
	opt := ps.ConsumeOptions{
		Durable:    true,
		NoWait:     false,
		Exclusive:  false,
		WithRetry:  true,
		DelayMilli: 300 * 1000, // Every 5 minute
		MaxRetries: 36,         // 3 hours in general
	}
	records, err := pubsub.Consume("records", "", "records", opt)
	if err != nil {
		log.Fatal("Failed to consume records", zap.Error(err))
		return
	}

	s.ConsumerStatus.Status.Status = healthpb.Status_RUNNING
	currencyConf := MakeCurrencyConf(log, &s.settingsClient)

	for {
		select {
		case <-_ctx.Done():
			log.Info("Context is done. Quitting")
			return
		case msg, ok := <-records:
			if !ok {
				log.Warn("Messages channel is closed")
				goto start
			}

			log.Debug("Received a message")
			var record pb.Record
			err = proto.Unmarshal(msg.Body, &record)
			log.Debug("Message unmarshalled", zap.Any("record", &record))
			if err != nil {
				log.Error("Failed to unmarshal record. Skip", zap.Error(err))
				if err = msg.Ack(false); err != nil {
					log.Error("Failed to Acknowledge the delivery while unmarshal message", zap.Error(err))
				}
				continue
			}
			err := s.ProcessRecord(ctx, &record, currencyConf, time.Now().Unix())
			if err != nil {
				ps.HandleAckNack(log, msg, err)
				continue
			}
			if err = msg.Ack(false); err != nil {
				log.Error("Failed to Acknowledge the delivery while unmarshal message", zap.Error(err))
			}
			continue
		}
	}

}

func (s *RecordsServiceServer) ProcessRecord(ctx context.Context, record *pb.Record, currencyConf CurrencyConf, now int64) error {
	log := s.log.Named("Process record")
	log = log.With(zap.String("instance", record.GetInstance()),
		zap.String("product", record.GetProduct()),
		zap.String("resource", record.GetResource()),
		zap.String("addon", record.GetAddon()))

	if record.Total == 0 {
		log.Warn("Got zero record, skipping", zap.Any("record", &record))
		return nil
	}

	isSkip, err := s.checkRecord(ctx, record, log)
	if isSkip && err == nil {
		log.Info("Skip prepaid record for suspended account", zap.Any("record", &record))
		return nil
	}

	if record.GetMeta() == nil {
		record.Meta = map[string]*structpb.Value{}
	}

	// Find out record's item price (product, resource or addon)
	var itemPrice float64
	inst, err := s.instances.GetWithAccess(ctx, driver.NewDocumentID(schema.ACCOUNTS_COL, schema.ROOT_ACCOUNT_KEY), record.Instance)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return err
	}
	bp, err := s.plans.Get(ctx, &pb.Plan{Uuid: inst.BillingPlan.Uuid})
	if err != nil {
		log.Error("Failed to get plan", zap.Error(err))
		return err
	}

	_, ok := bp.Products[record.Product]
	if _, okWithAddon := bp.Products[inst.GetProduct()]; (!ok && record.Product != "") || (!okWithAddon && record.Addon != "") {
		log.Error("Invalid record. Addon record or product record, but no product in billing plan", zap.Any("record", &record))
		return nil
	}

	var (
		resType, resource string
	)
	if record.Product != "" {
		itemPrice = bp.Products[record.Product].Price
		resType = "product"
		resource = record.Product
	} else if record.Resource != "" {
		for _, res := range bp.Resources {
			if res.Key == record.Resource {
				itemPrice = res.Price
				resType = "resource"
				resource = record.Resource
				break
			}
		}
	} else if record.Addon != "" {
		if inst.Product == nil || *inst.Product == "" {
			log.Error("Invalid record. Record is addon record but no product for instance", zap.Any("record", &record))
			return nil
		}
		product := bp.Products[*inst.Product]
		addon, err := s.addons.Get(ctx, record.Addon)
		if err != nil {
			log.Error("Failed to get addon", zap.Error(err))
			return err
		}
		itemPrice = addon.Periods[product.Period]
		resType = "addon"
		resource = record.Addon
	} else {
		log.Error("Invalid record. Record has no item", zap.Any("record", &record))
		return nil
	}
	log.Debug("Got item price", zap.Any("itemPrice", itemPrice))
	record.Meta["no_discount_price"] = structpb.NewStringValue(fmt.Sprintf("%.2f %s", itemPrice, currencyConf.Currency.GetTitle()))
	if itemPrice, _, err = s.promocodes.GetDiscountPriceByResource(inst.Instance, currencyConf.Currency, itemPrice,
		currencyConf.Currency, resType, resource); err != nil {
		log.Error("Failed to get discount price", zap.Error(err))
		return err
	}
	log.Debug("Item price after discount", zap.Any("itemPrice", itemPrice))

	record.Meta["transactionType"] = structpb.NewStringValue("system")
	record.Cost = itemPrice

	recordId := s.records.Create(ctx, record)
	if recordId == "" {
		log.Warn("No record id. Skipping")
		return nil
	}

	log.Debug("Record created", zap.String("record_id", recordId.Key()))
	s.ConsumerStatus.LastExecution = time.Now().Format("2006-01-02T15:04:05Z07:00")
	if record.Priority != pb.Priority_NORMAL {
		started := time.Now()
		trCtx, err := graph.BeginTransaction(ctx, s.db, driver.TransactionCollections{
			Exclusive: []string{schema.TRANSACTIONS_COL, schema.RECORDS_COL, schema.ACCOUNTS_COL, schema.INVOICES_COL},
		})
		if err != nil {
			log.Error("Failed to start transaction", zap.Error(err))
			return err
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

		cur, err := s.db.Query(trCtx, generateUrgentTransactionsByRecord, map[string]interface{}{
			"@transactions":  schema.TRANSACTIONS_COL,
			"@instances":     schema.INSTANCES_COL,
			"@billing_plans": schema.BILLING_PLANS_COL,
			"@services":      schema.SERVICES_COL,
			"@records":       schema.RECORDS_COL,
			"@accounts":      schema.ACCOUNTS_COL,
			"@addons":        schema.ADDONS_COL,
			"permissions":    schema.PERMISSIONS_GRAPH.Name,
			"priority":       record.Priority,
			"now":            now,
			"graph":          schema.BILLING_GRAPH.Name,
			"currencies":     schema.CUR_COL,
			"currency":       currencyConf.Currency,
		})
		if err != nil {
			abort()
			log.Error("Error Generating Transactions", zap.Error(err))
			return err
		}
		sinceGeneration := time.Since(started)

		var tr pb.Transaction
		if cur.HasMore() {
			doc, err := cur.ReadDocument(ctx, &tr)
			if err != nil {
				return err
			}
			_, err = s.db.Query(trCtx, processUrgentTransactions, map[string]interface{}{
				"tr":            doc.ID,
				"@transactions": schema.TRANSACTIONS_COL,
				"@accounts":     schema.ACCOUNTS_COL,
				"accounts":      schema.ACCOUNTS_COL,
				"@records":      schema.RECORDS_COL,
				"now":           now,
				"graph":         schema.BILLING_GRAPH.Name,
				"currencies":    schema.CUR_COL,
				"currency":      currencyConf.Currency,
			})
			if err != nil {
				abort()
				log.Error("Error Process Transactions", zap.Error(err))
				return err
			}
		}

		if err = commit(); err != nil {
			return err
		}
		log.Debug("Processing complete", zap.Float64("generating_duration", sinceGeneration.Seconds()),
			zap.Float64("sum_duration", time.Since(started).Seconds()), zap.String("record_id", recordId.Key()))
	}

	log.Info("Record processed", zap.String("record_id", recordId.Key()))
	return nil
}

func (s *RecordsServiceServer) checkRecord(ctx context.Context, rec *pb.Record, log *zap.Logger) (bool, error) {
	if rec.Product == "" {
		return false, nil
	}

	inst := driver.NewDocumentID(schema.INSTANCES_COL, rec.Instance)

	accCursor, err := s.db.Query(ctx, getAccountStatus, map[string]interface{}{
		"inst":        inst,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"@services":   schema.SERVICES_COL,
		"@accounts":   schema.ACCOUNTS_COL,
	})
	if err != nil {
		log.Error("Cannot get acc status", zap.Error(err))
		return false, err
	}

	isSuspended := false

	if accCursor.HasMore() {
		_, err := accCursor.ReadDocument(ctx, &isSuspended)
		if err != nil {
			log.Error("Cannot get acc status", zap.Error(err))
			return false, err
		}
	}

	instCursor, err := s.db.Query(ctx, getPlanFromRecord, map[string]interface{}{
		"inst":    inst,
		"product": rec.Product,
	})
	if err != nil {
		log.Error("Cannot get instance plan", zap.Error(err))
		return false, err
	}

	kind := pb.Kind_POSTPAID

	if instCursor.HasMore() {
		_, err := instCursor.ReadDocument(ctx, &kind)
		if err != nil {
			log.Error("Cannot get instance plan", zap.Error(err))
			return false, err
		}
	}

	if kind == pb.Kind_PREPAID && isSuspended {
		return true, nil
	}

	return false, nil
}

const getAccountStatus = `
LET doc = DOCUMENT(@inst)

LET srv = LAST(
FOR node, edge, path IN 2
    INBOUND doc
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@services)
        RETURN node
    )

LET account = LAST(
    FOR node, edge, path IN 2
    INBOUND srv
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@accounts)
        RETURN LENGTH(node.account_owner) > 0 ? DOCUMENT(@@accounts, node.account_owner) : node
    )
    
RETURN account.suspended
`

const getPlanFromRecord = `
LET doc = DOCUMENT(@inst)
LET d = doc.billing_plan.products[@product]
RETURN d.kind
`

const generateUrgentTransactionsByRecord = `
        FOR record IN @@records
        FILTER !record.processed && record.instance && record.instance != "" && record.priority == @priority

        LET instance = DOCUMENT(@@instances, record.instance)
        FILTER instance
        LET account = LAST( // Find Instance owner Account
    		FOR node, edge, path IN 4
    		INBOUND instance
    		GRAPH @permissions
    		FILTER path.edges[*].role == ["owner","owner","owner","owner"]
    		FILTER IS_SAME_COLLECTION(node, @@accounts)
            RETURN LENGTH(node.account_owner) > 0 ? DOCUMENT(@@accounts, node.account_owner) : node
        )
        LET service = LAST( // Find Instance service
    		FOR node, edge, path IN 2
    		INBOUND instance
    		GRAPH @permissions
    		FILTER path.edges[*].role == ["owner","owner"]
    		FILTER IS_SAME_COLLECTION(node, @@services)
            RETURN node
        )
        LET currency = account.currency != null ? account.currency : @currency

		LET rate = PRODUCT(
			FOR vertex, edge IN OUTBOUND SHORTEST_PATH
			// Cast to NCU if currency is not specified
			DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(record.currency.id))) TO
			DOCUMENT(CONCAT(@currencies, "/", currency.id)) GRAPH @graph
			FILTER edge
				RETURN edge.rate
		)

        LET bp = DOCUMENT(@@billing_plans, instance.billing_plan.uuid)
        LET resources = bp.resources == null ? [] : bp.resources
        LET addon = DOCUMENT(@@addons, record.addon)
        LET product_period = bp.products[instance.product].period
        LET item_price = record.product == null ? (record.resource == null ? addon.periods[product_period] : LAST(FOR res in resources FILTER res.key == record.resource return res).price) : bp.products[record.product].price
        LET final_price = record.cost > 0 ? record.cost : item_price // If already defined when was created, then use this value. If not, then use calculated value

        LET cost = record.total * rate * final_price
        UPDATE record._key WITH { 
           processed: true, 
           cost: cost,
           currency: currency,
           service: service._key,
           account: account._key
        } IN @@records

        INSERT {
            exec: @now, // Timestamp in seconds
	    	created: @now,
            processed: false,
	    	currency: currency,
            account: account._key,
	    	priority: @priority,
            service: service._key,
            records: [record._key],
            total: cost,
	    	meta: {type: "transaction"},
        } IN @@transactions RETURN NEW
`

const generateUrgentTransactions = `
FOR service IN @@services // Iterate over Services
	LET instances = (
        FOR i IN 2 OUTBOUND service
        GRAPH @permissions
        FILTER IS_SAME_COLLECTION(@@instances, i)
            RETURN i._key )

    LET account = LAST( // Find Service owner Account
    FOR node, edge, path IN 2
    INBOUND service
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@accounts)
        RETURN LENGTH(node.account_owner) > 0 ? DOCUMENT(@@accounts, node.account_owner) : node
    )

	LET currency = account.currency != null ? account.currency : @currency
    
    LET records = ( // Collect all unprocessed records
        FOR record IN @@records
        FILTER !record.processed && record.instance IN instances && record.priority == @priority

		LET rate = PRODUCT(
			FOR vertex, edge IN OUTBOUND SHORTEST_PATH
			// Cast to NCU if currency is not specified
			DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(record.currency.id))) TO
			DOCUMENT(CONCAT(@currencies, "/", currency.id)) GRAPH @graph
			FILTER edge
				RETURN edge.rate
		)

        LET instance = DOCUMENT(@@instances, record.instance)
        LET bp = DOCUMENT(@@billing_plans, instance.billing_plan.uuid)
        LET resources = bp.resources == null ? [] : bp.resources
        LET addon = DOCUMENT(@@addons, record.addon)
        LET product_period = bp.products[instance.product].period
        LET item_price = record.product == null ? (record.resource == null ? addon.periods[product_period] : LAST(FOR res in resources FILTER res.key == record.resource return res).price) : bp.products[record.product].price
        LET final_price = record.cost > 0 ? record.cost : item_price // If already defined when was created, then use this value. If not, then use calculated value

        LET cost = record.total * rate * final_price
            UPDATE record._key WITH { 
				processed: true, 
				cost: cost,
				currency: currency,
				service: service._key,
				account: account._key
			} IN @@records RETURN NEW
    )
    
    FILTER LENGTH(records) > 0 // Skip if no Records (no empty Transaction)
    INSERT {
        exec: @now, // Timestamp in seconds
		created: @now,
        processed: false,
		currency: currency,
        account: account._key,
		priority: @priority,
        service: service._key,
        records: records[*]._key,
        total: SUM(records[*].cost),
		meta: {type: "transaction"},
    } IN @@transactions RETURN NEW
`

const processUrgentTransactions = `
LET t = DOCUMENT(@tr) 

LET account = DOCUMENT(CONCAT(@accounts, "/", t.account))

LET currency = account.currency != null ? account.currency : @currency
LET rate = PRODUCT(
	FOR vertex, edge IN OUTBOUND SHORTEST_PATH
	DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(t.currency.id))) TO
	DOCUMENT(CONCAT(@currencies, "/", currency.id)) GRAPH @graph
	FILTER edge
	RETURN edge.rate
)
LET total = t.total * rate

FOR r in t.records
UPDATE r WITH {meta: {transaction: t._key, payment_date: @now}} in @@records

UPDATE account WITH { balance: account.balance - t.total * rate} IN @@accounts
UPDATE t WITH { 
	processed: true, 
	proc: @now,
	total: total,
	currency: currency
} IN @@transactions
`

func (s *BillingServiceServer) GetRecords(ctx context.Context, r *connect.Request[pb.Transaction]) (*connect.Response[pb.Records], error) {
	log := s.log.Named("GetRecords")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	if req.Uuid == "" {
		log.Error("Request has no UUID", zap.String("requestor", requestor))
		return nil, status.Error(codes.InvalidArgument, "Request has no UUID")
	}

	tr, err := s.transactions.Get(ctx, req.Uuid)
	if err != nil {
		log.Error("Failed to get transaction", zap.String("requestor", requestor), zap.String("uuid", req.Uuid))
		return nil, status.Error(codes.NotFound, "Transaction not found")
	}
	log.Debug("Transaction found", zap.String("requestor", requestor), zap.Any("transaction", tr))

	ok := s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.ACCOUNTS_COL, tr.Account), access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Permission denied")
	}

	pool, err := s.records.Get(ctx, req.Uuid)
	if err != nil {
		log.Error("Failed to get records", zap.String("requestor", requestor), zap.String("uuid", req.Uuid))
		return nil, status.Error(codes.Internal, "Failed to get Records")
	}

	log.Debug("Records found", zap.String("transaction", tr.Uuid), zap.Any("records", pool))

	resp := connect.NewResponse(&pb.Records{Pool: pool})

	return resp, nil
}

func (s *BillingServiceServer) GetInstancesReports(ctx context.Context, r *connect.Request[pb.GetInstancesReportRequest]) (*connect.Response[pb.GetInstancesReportResponse], error) {
	log := s.log.Named("GetRecords")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	ok := s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Permission denied")
	}

	reports, err := s.records.GetInstancesReports(ctx, req)
	if err != nil {
		log.Error("Failed to get reports", zap.Error(err))
	}

	resp := connect.NewResponse(&pb.GetInstancesReportResponse{Reports: reports})

	return resp, nil
}

func (s *BillingServiceServer) GetInstancesReportsCount(ctx context.Context, r *connect.Request[pb.GetInstancesReportsCountRequest]) (*connect.Response[pb.GetReportsCountResponse], error) {
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	ok := s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Permission denied")
	}

	res, err := s.records.GetInstancesReportsCount(ctx)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&pb.GetReportsCountResponse{Total: res})

	return resp, nil
}

func (s *BillingServiceServer) GetRecordsReports(ctx context.Context, r *connect.Request[pb.GetRecordsReportsRequest]) (*connect.Response[pb.GetRecordsReportsResponse], error) {
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg
	if req.Account != nil || req.Service != nil {
		if req.Account != nil {
			acc := *req.Account
			node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
			if !s.ca.HasAccess(ctx, requestor, node, access.Level_ADMIN) {
				return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
			}
		}

		if req.Service != nil {
			srv := *req.Account
			node := driver.NewDocumentID(schema.ACCOUNTS_COL, srv)
			if !s.ca.HasAccess(ctx, requestor, node, access.Level_ADMIN) {
				return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
			}
		}
	} else {
		ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
		if ok := s.ca.HasAccess(ctx, requestor, ns, access.Level_ROOT); !ok {
			return nil, status.Error(codes.PermissionDenied, "Permission denied")
		}
	}

	return s.records.GetRecordsReports(ctx, req)
}

func (s *BillingServiceServer) GetRecordsReportsCount(ctx context.Context, r *connect.Request[pb.GetRecordsReportsCountRequest]) (*connect.Response[pb.GetReportsCountResponse], error) {
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	if req.Account != nil || req.Service != nil {
		if req.Account != nil {
			acc := *req.Account
			node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
			if !s.ca.HasAccess(ctx, requestor, node, access.Level_ADMIN) {
				return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
			}
		}

		if req.Service != nil {
			srv := *req.Account
			node := driver.NewDocumentID(schema.ACCOUNTS_COL, srv)
			if !s.ca.HasAccess(ctx, requestor, node, access.Level_ADMIN) {
				return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
			}
		}
	} else {
		ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
		if ok := s.ca.HasAccess(ctx, requestor, ns, access.Level_ROOT); !ok {
			return nil, status.Error(codes.PermissionDenied, "Permission denied")
		}
	}

	res, err := s.records.GetRecordsReportsCount(ctx, req)

	if err != nil {
		return nil, err
	}

	types, err := s.records.GetUnique(ctx)

	if err != nil {
		return nil, err
	}

	s.log.Debug("Types", zap.Any("types", types))

	value, err := structpb.NewValue(types)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&pb.GetReportsCountResponse{Total: res, Unique: value})

	return resp, nil
}
