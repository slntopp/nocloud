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
	"connectrpc.com/connect"
	"context"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud-proto/services"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"strings"
	"time"

	sc "github.com/slntopp/nocloud/pkg/settings/client"

	epb "github.com/slntopp/nocloud-proto/events"
	hpb "github.com/slntopp/nocloud-proto/health"
	ipb "github.com/slntopp/nocloud-proto/instances"
	regpb "github.com/slntopp/nocloud-proto/registry"
	accpb "github.com/slntopp/nocloud-proto/registry/accounts"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	stpb "github.com/slntopp/nocloud-proto/states"
	spb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	settingsClient settingspb.SettingsServiceClient
	accClient      regpb.AccountsServiceClient
	eventsClient   epb.EventsServiceClient
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("SETTINGS_HOST", "settings:8000")
	viper.SetDefault("REGISTRY_HOST", "registry:8000")
	viper.SetDefault("EVENTS_HOST", "eventbus:8000")
	settingsHost := viper.GetString("SETTINGS_HOST")
	registryHost := viper.GetString("REGISTRY_HOST")
	eventsHost := viper.GetString("EVENTS_HOST")

	settingsConn, err := grpc.Dial(settingsHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	settingsClient = settingspb.NewSettingsServiceClient(settingsConn)

	accConn, err := grpc.Dial(registryHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	accClient = regpb.NewAccountsServiceClient(accConn)

	eventsConn, err := grpc.Dial(eventsHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	eventsClient = epb.NewEventsServiceClient(eventsConn)
}

func (s *BillingServiceServer) GenTransactionsRoutineState() []*hpb.RoutineStatus {
	return []*hpb.RoutineStatus{
		s.gen, s.proc,
	}
}

func (s *BillingServiceServer) InvoiceExpiringInstances(ctx context.Context, log *zap.Logger, tick time.Time,
	currencyConf CurrencyConf, roundingConf RoundingConf) {
	log.Info("Issuing invoices for expiring instances", zap.Time("tick", tick))

	list, err := s.services.List(ctx, schema.ROOT_ACCOUNT_KEY, &services.ListRequest{})
	if err != nil {
		log.Error("Error getting services", zap.Error(err))
		return
	}

	log.Debug("Got list of services", zap.Any("count", len(list.Result)))

	// Before 2.4 hours before daily payment
	// Before 3 days before monthly payment and so on
	expiringTime := func(expiringAt, period int64) int64 {
		return expiringAt - (period / 10)
	}

	counter := 0
	// Process every instance
	for _, srv := range list.Result {
		for _, ig := range srv.GetInstancesGroups() {
			for _, i := range ig.GetInstances() {
				log := log.With(zap.String("instance", i.GetUuid()))

				if i.GetStatus() == spb.NoCloudStatus_DEL ||
					i.GetState().GetState() == stpb.NoCloudState_PENDING ||
					i.GetStatus() == spb.NoCloudStatus_SUS {
					log.Info("Instance has been deleted or PENDING or SUSPENDED state. Skipping")
					continue
				}

				cost := 0.0
				plan := i.GetBillingPlan()
				now := time.Now().Unix()

				if i.Data == nil {
					i.Data = map[string]*structpb.Value{}
				}

				expiringResources := false
				expiringProduct := false
				for _, res := range plan.GetResources() {
					log := log.With(zap.String("resource", res.GetKey()))

					if res.GetPeriod() == 0 {
						log.Debug("Resource period is 0. Skipping")
						continue
					}

					lastMonitoringValue, ok := i.Data[res.GetKey()+"_last_monitoring"]
					if !ok {
						log.Debug("Last monitoring not found. Skipping")
						continue
					}
					lm := int64(lastMonitoringValue.GetNumberValue())
					expiringAt := expiringTime(lm, res.GetPeriod())

					// Checking if invoice was issued for this last_monitoring value
					issuedInvTsValue, ok := i.Data[res.GetKey()+"_lm_for_last_issued_invoice"]
					issuedInvTs := int64(issuedInvTsValue.GetNumberValue())
					wasIssued := ok && lm == issuedInvTs
					if wasIssued {
						log.Info("Invoice for current renew cycle was issued before.")
					}

					if now >= expiringAt && !wasIssued {
						log.Debug("Resource is expiring.", zap.Int64("expiring_at", expiringAt), zap.Int64("now", now))
						expiringResources = true
					}

					switch res.GetKey() {
					case "cpu":
						value := i.GetResources()["cpu"].GetNumberValue()
						cost += value * res.GetPrice()
					case "ram":
						value := i.GetResources()["ram"].GetNumberValue() / 1024
						cost += value * res.GetPrice()
					default:
						// Calculate drive size billing
						if strings.Contains(res.GetKey(), "drive") {
							driveType := i.GetResources()["drive_type"].GetStringValue()
							if res.GetKey() != "drive_"+strings.ToLower(driveType) {
								continue
							}
							count := i.GetResources()["drive_size"].GetNumberValue() / 1024
							cost += res.GetPrice() * count
						}
					}

					i.Data[res.GetKey()+"_lm_for_last_issued_invoice"] = structpb.NewNumberValue(float64(lm))
				}

				productBased := true
				product, ok := i.GetBillingPlan().GetProducts()[i.GetProduct()]
				if !ok || product == nil {
					log.Error("Failed to get product(instance's product not found in billing plan or nil)",
						zap.String("product_id", i.GetProduct()))
					productBased = false
				} else {
					if product.GetPeriod() == 0 {
						log.Info("Product period is 0. Skipping billing for this product, but not skipping for resources.", zap.String("product", product.GetTitle()))
						productBased = false
						goto resume
					}
					lastMonitoringValue, ok := i.Data["last_monitoring"]
					if !ok {
						log.Debug("Last monitoring for product not found. Skipping", zap.String("product", product.GetTitle()))
						goto resume
					}
					lm := int64(lastMonitoringValue.GetNumberValue())

					// Checking if invoice was issued for this last_monitoring value
					issuedInvTsValue, ok := i.Data["lm_for_last_issued_invoice"]
					issuedInvTs := int64(issuedInvTsValue.GetNumberValue())
					wasIssued := ok && lm == issuedInvTs
					if wasIssued {
						log.Info("Invoice for current renew cycle was issued before", zap.String("product", product.GetTitle()))
					}

					expiringAt := expiringTime(lm, product.GetPeriod())
					if now >= expiringAt && !wasIssued {
						log.Debug("Product is expiring.", zap.Int64("expiring_at", expiringAt), zap.Int64("now", now), zap.String("product", product.GetTitle()))
						expiringProduct = true
					}
					cost += product.GetPrice()

					i.Data["lm_for_last_issued_invoice"] = structpb.NewNumberValue(float64(lm))
				}

			resume:

				if productBased && !expiringProduct {
					log.Info("Product is not expiring. Skipping invoice creation.", zap.String("product", product.GetTitle()))
					if expiringResources {
						log.Info("Resources are expiring. But skipping invoice creation, because product is not expiring.")
					}
					continue
				}

				if !productBased && !expiringResources {
					log.Info("Resources are not expiring in NonProduct instance. Skipping invoice creation.")
					continue
				}

				log.Debug("Proceeding to invoice creation")

				// Find owner account
				cur, err := s.db.Query(ctx, instanceOwner, map[string]interface{}{
					"instance":    i.GetUuid(),
					"permissions": schema.PERMISSIONS_GRAPH.Name,
					"@instances":  schema.INSTANCES_COL,
					"@accounts":   schema.ACCOUNTS_COL,
				})
				if err != nil {
					log.Error("Error getting instance owner", zap.Error(err))
					continue
				}
				var acc graph.Account
				_, err = cur.ReadDocument(ctx, &acc)
				if err != nil {
					log.Error("Error getting instance owner", zap.Error(err))
					continue
				}
				acc.Uuid = acc.Key
				if acc.GetUuid() == "" {
					log.Error("Instance owner not found")
					continue
				}
				log.Debug("Instance owner found", zap.String("account", acc.GetUuid()))

				if acc.Currency == nil {
					acc.Currency = &currencyConf.Currency
				}
				rate, err := s.currencies.GetExchangeRate(ctx, &currencyConf.Currency, acc.Currency)
				if err != nil {
					log.Error("Error getting exchange rate", zap.Error(err))
					continue
				}

				cost *= rate // Convert from NCU to  account's currency

				log.Debug("Updating instance")
				// Update instance with stamps to track issued invoice
				err = s.instances.Update(context.WithValue(ctx, nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY),
					"", proto.Clone(i).(*ipb.Instance), proto.Clone(i).(*ipb.Instance))
				if err != nil {
					log.Error("Failed to update instance. Not issuing invoice.", zap.Error(err))
					continue
				}

				inv := &pb.Invoice{
					Status: pb.BillingStatus_UNPAID,
					Items: []*pb.Item{
						{
							Title:    fmt.Sprintf("Instance '%s' renewal", i.GetTitle()),
							Amount:   1,
							Unit:     "Instance",
							Price:    cost,
							Instance: i.GetUuid(),
						},
					},
					Total:    cost,
					Type:     pb.ActionType_INSTANCE_RENEWAL,
					Created:  now,
					Deadline: time.Unix(now, 0).Add(24 * time.Hour).Unix(), // Until when invoice should be paid
					Account:  acc.GetUuid(),
					Currency: acc.Currency,
					Meta: map[string]*structpb.Value{
						"auto_created": structpb.NewBoolValue(true),
						"creator":      structpb.NewStringValue("nocloud.billing.IssueInvoicesRoutine"),
					},
				}

				ctx = context.WithValue(ctx, nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY)
				_, err = s.CreateInvoice(ctx, connect.NewRequest(inv))
				if err != nil {
					log.Error("Failed to create invoice", zap.Error(err))
					continue
				}

				log.Debug("Invoice created", zap.String("invoice", inv.Uuid))
				counter++
			}
		}
	}

	log.Info("Routine finished", zap.Int("invoices_created", counter))
}

func (s *BillingServiceServer) IssueInvoicesRoutine(ctx context.Context) {
	log := s.log.Named("IssueInvoicesRoutine")

start:

	routineConf := MakeRoutineConf(ctx, log)
	roundingConf := MakeRoundingConf(ctx, log)
	currencyConf := MakeCurrencyConf(ctx, log)

	upd := make(chan bool, 1)
	go sc.Subscribe([]string{currencyKey}, upd)

	log.Info("Got Configuration", zap.Any("currency", currencyConf), zap.Any("routine", routineConf), zap.Any("rounding", roundingConf))

	ticker := time.NewTicker(time.Second * time.Duration(routineConf.Frequency))
	tick := time.Now()
	for {
		log.Info("Entering new Iteration", zap.Time("ts", tick))
		s.InvoiceExpiringInstances(ctx, log, tick, currencyConf, roundingConf)

		s.proc.LastExecution = tick.Format("2006-01-02T15:04:05Z07:00")
		select {
		case tick = <-ticker.C:
			continue
		case <-upd:
			log.Info("New Configuration Received, restarting Routine")
			goto start
		}
	}
}

func (s *BillingServiceServer) GenTransactions(ctx context.Context, log *zap.Logger, tick time.Time,
	currencyConf CurrencyConf, roundingConf RoundingConf) {
	log.Info("Starting Generating Transactions Sub-Routine", zap.Time("tick", tick))
	s.gen.Status.Status = hpb.Status_RUNNING
	s.gen.Status.Error = nil

	_, err := s.db.Query(ctx, generateTransactions, map[string]interface{}{
		"@transactions": schema.TRANSACTIONS_COL,
		"@instances":    schema.INSTANCES_COL,
		"@services":     schema.SERVICES_COL,
		"@records":      schema.RECORDS_COL,
		"@accounts":     schema.ACCOUNTS_COL,
		"permissions":   schema.PERMISSIONS_GRAPH.Name,
		"now":           tick.Unix(),
		"graph":         schema.BILLING_GRAPH.Name,
		"currencies":    schema.CUR_COL,
		"currency":      currencyConf.Currency,
	})
	if err != nil {
		log.Error("Error Generating Transactions", zap.Error(err))
		s.gen.Status.Status = hpb.Status_HASERRS
		err_s := err.Error()
		s.gen.Status.Error = &err_s
	}
	s.gen.LastExecution = tick.Format("2006-01-02T15:04:05Z07:00")

	log.Info("Starting Processing Transactions Sub-Routine", zap.Time("tick", tick))
	s.proc.Status.Status = hpb.Status_RUNNING
	s.gen.Status.Error = nil

	_, err = s.db.Query(ctx, processTransactions, map[string]interface{}{
		"@transactions": schema.TRANSACTIONS_COL,
		"@accounts":     schema.ACCOUNTS_COL,
		"@records":      schema.RECORDS_COL,
		"accounts":      schema.ACCOUNTS_COL,
		"now":           tick.Unix(),
		"graph":         schema.BILLING_GRAPH.Name,
		"currencies":    schema.CUR_COL,
		"currency":      currencyConf.Currency,
	})
	if err != nil {
		log.Error("Error Processing Transactions", zap.Error(err))
		s.proc.Status.Status = hpb.Status_HASERRS
		err_s := err.Error()
		s.proc.Status.Error = &err_s
	}
}

func (s *BillingServiceServer) SuspendAccountsRoutineState() *hpb.RoutineStatus {
	return s.sus
}

func (s *BillingServiceServer) SuspendAccountsRoutine(ctx context.Context) {
	log := s.log.Named("AccountSuspendRoutine")

start:
	suspConf := MakeSuspendConf(ctx, log)
	routineConf := MakeRoutineConf(ctx, log)

	upd := make(chan bool, 1)
	go sc.Subscribe([]string{monFreqKey}, upd)

	log.Info("Got Configuration", zap.Any("suspend", suspConf), zap.Any("routine", routineConf))

	ticker := time.NewTicker(time.Second * time.Duration(routineConf.Frequency))
	tick := time.Now()

	for {
		s.sus.Status.Status = hpb.Status_RUNNING
		s.sus.LastExecution = tick.Format("2006-01-02T15:04:05Z07:00")
		s.sus.Status.Error = nil

		cursor, err := s.db.Query(ctx, accToSuspend, map[string]interface{}{
			"conf": suspConf,
		})

		if err != nil {
			log.Error("Error Quering Accounts to Suspend", zap.Error(err))
			s.sus.Status.Status = hpb.Status_HASERRS
			err_str := fmt.Sprintf("Error Quering Accounts to Suspend: %s", err.Error())
			s.sus.Status.Error = &err_str

			time.Sleep(time.Second)
			continue
		}

		for cursor.HasMore() {
			acc := &accpb.Account{}
			meta, err := cursor.ReadDocument(ctx, &acc)
			log.Info("Acc id", zap.Any("id", meta.ID))
			if err != nil {
				log.Error("Error Reading Account", zap.Error(err), zap.Any("meta", meta))
				continue
			}
			if _, err := accClient.Suspend(ctx, &accpb.SuspendRequest{Uuid: acc.GetUuid()}); err != nil {
				log.Error("Error Suspending Account", zap.Error(err))
			}
		}

		cursor2, err := s.db.Query(ctx, accToUnsuspend, map[string]interface{}{
			"conf": suspConf,
		})
		if err != nil {
			log.Error("Error Quering Accounts to Unsuspend", zap.Error(err))
			s.sus.Status.Status = hpb.Status_HASERRS
			err_str := fmt.Sprintf("Error Quering Accounts to Unsuspend: %s", err.Error())
			s.sus.Status.Error = &err_str

			time.Sleep(time.Second)
			continue
		}

		for cursor2.HasMore() {
			acc := &accpb.Account{}
			meta, err := cursor2.ReadDocument(ctx, &acc)
			log.Info("Acc id", zap.Any("id", meta.ID))
			if err != nil {
				log.Error("Error Reading Account", zap.Error(err), zap.Any("meta", meta))
				continue
			}
			if _, err := accClient.Unsuspend(ctx, &accpb.UnsuspendRequest{Uuid: acc.GetUuid()}); err != nil {
				log.Error("Error Unsuspending Account", zap.Error(err))
			}
		}

		select {
		case tick = <-ticker.C:
			continue
		case <-upd:
			log.Info("New Configuration Received, restarting Routine")
			goto start
		}
	}

}

func (s *BillingServiceServer) GenTransactionsRoutine(ctx context.Context) {
	log := s.log.Named("GenerateTransactionsRoutine")

start:

	routineConf := MakeRoutineConf(ctx, log)
	roundingConf := MakeRoundingConf(ctx, log)
	currencyConf := MakeCurrencyConf(ctx, log)

	upd := make(chan bool, 1)
	go sc.Subscribe([]string{monFreqKey, currencyKey}, upd)

	log.Info("Got Configuration", zap.Any("currency", currencyConf), zap.Any("routine", routineConf), zap.Any("rounding", roundingConf))

	ticker := time.NewTicker(time.Second * time.Duration(routineConf.Frequency))
	tick := time.Now()
	for {
		log.Info("Entering new Iteration", zap.Time("ts", tick))
		s.GenTransactions(ctx, log, tick, currencyConf, roundingConf)

		s.proc.LastExecution = tick.Format("2006-01-02T15:04:05Z07:00")
		select {
		case tick = <-ticker.C:
			continue
		case <-upd:
			log.Info("New Configuration Received, restarting Routine")
			goto start
		}
	}
}

const accToUnsuspend = `
let conf = @conf

let candidates = (
	for acc in Accounts
		filter acc.suspended
		filter conf.auto_resume
		return acc
)

let local = (
    for acc in candidates
        filter acc.suspend['limit'] && (acc.balance >= acc.suspend['limit'])
        return acc
)
    
let global = (
    for acc in candidates
        filter acc.balance >= conf['limit']
        filter acc.balance >= acc.suspend['limit']
        return acc
)

FOR acc IN union_distinct(local, global)
    RETURN MERGE(acc, {uuid:acc._key})
`

const accToSuspend = `
LET conf = @conf

LET now = DATE_NOW()
LET day = DATE_DAYOFWEEK(now)
LET hour = DATE_HOUR(now)
LET now_matching = !conf.schedule[day].off && hour >= conf.schedule[day].from && hour <= conf.schedule[day].to

LET candidates = (
    FOR acc IN Accounts
        FILTER acc.balance != null
		FILTER !acc.suspended
        FILTER !acc.suspend_conf.immune
        return acc
)

LET global = (
    FOR acc IN candidates
        FILTER now_matching
        FILTER conf.is_enabled
        FILTER acc.balance < conf['limit']
        FILTER (acc.balance - acc.suspend_conf['limit']) < 0
        RETURN acc
)

LET extra = (
    FOR acc IN candidates
        FILTER conf.is_extra_enabled
        FILTER acc.balance < conf.extra_limit
        RETURN acc
)

LET local = (
    FOR acc IN candidates
		FILTER now_matching
        FILTER acc.balance < acc.suspend_conf['limit']
        RETURN acc
)

FOR acc IN UNION_DISTINCT(global, local, extra)
	RETURN MERGE(acc, {uuid: acc._key})
`

const generateTransactions = `
FOR service IN @@services // Iterate over Services
	LET instances = (
        FOR i IN 2 OUTBOUND service
        GRAPH @permissions
        FILTER IS_SAME_COLLECTION(@@instances, i)
            RETURN i._key 
	)

    LET account = LAST( // Find Service owner Account
		FOR node, edge, path IN 2
		INBOUND service
		GRAPH @permissions
		FILTER path.edges[*].role == ["owner","owner"]
		FILTER IS_SAME_COLLECTION(node, @@accounts)
			RETURN node
    )

	// Prefer user currency to default if present
	LET currency = account.currency != null ? account.currency : @currency

    LET records = ( // Collect all unprocessed records
        FOR record IN @@records
        FILTER record.exec <= @now
        FILTER !record.processed
        FILTER record.instance IN instances

		LET inst = DOCUMENT(CONCAT(@instances, "/", record.instance))
		LET bp = DOCUMENT(CONCAT(@billing_plans, "/", inst.billing_plan.uuid))
		LET resources = bp.resources == null ? [] : bp.resources
		LET item = record.product == null ? LAST(FOR res in resources FILTER res.key == record.resource return res) : bp.products[record.product]

		LET rate = PRODUCT(
			FOR vertex, edge IN OUTBOUND
			SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(record.currency.id)))
			TO DOCUMENT(CONCAT(@currencies, "/", currency.id))
			GRAPH @graph
			FILTER edge
				RETURN edge.rate
		)
        LET cost = record.total * rate * item.price
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
        service: service._key,
        records: records[*]._key,
        total: SUM(records[*].cost), // Calculate Total
		meta: {type: "transaction"},
    } IN @@transactions RETURN NEW
`

const processTransactions = `
FOR t IN @@transactions // Iterate over Transactions
FILTER t.exec != null
FILTER t.exec <= @now
FILTER !t.processed
    LET account = DOCUMENT(CONCAT(@accounts, "/", t.account))
	// Prefer user currency to default if present
	FILTER account != null
	LET currency = account.currency != null ? account.currency : @currency
	LET rate = PRODUCT(
		FOR vertex, edge IN OUTBOUND
		SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(t.currency.id)))
		TO DOCUMENT(CONCAT(@currencies, "/", currency.id))
		GRAPH @graph
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
