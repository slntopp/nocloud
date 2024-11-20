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
	"github.com/arangodb/go-driver"
	ipb "github.com/slntopp/nocloud-proto/instances"
	"google.golang.org/protobuf/proto"
	"time"

	"connectrpc.com/connect"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud-proto/services"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"google.golang.org/protobuf/types/known/structpb"

	sc "github.com/slntopp/nocloud/pkg/settings/client"

	dpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	hpb "github.com/slntopp/nocloud-proto/health"
	accpb "github.com/slntopp/nocloud-proto/registry/accounts"
	stpb "github.com/slntopp/nocloud-proto/states"
	spb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

func (s *BillingServiceServer) RoutinesState() []*hpb.RoutineStatus {
	return []*hpb.RoutineStatus{
		s.gen, s.proc, s.sus, s.inv,
	}
}

func (s *BillingServiceServer) InvoiceExpiringInstances(ctx context.Context, log *zap.Logger, tick time.Time,
	currencyConf CurrencyConf, roundingConf RoundingConf, iPub Pub) {
	log.Info("Trying to issue invoices for expiring instances", zap.Time("tick", tick))

	list, err := s.services.List(ctx, schema.ROOT_ACCOUNT_KEY, &services.ListRequest{})
	if err != nil {
		log.Error("Error getting services", zap.Error(err))
		return
	}

	log.Debug("Got list of services", zap.Any("count", len(list.Result)))
	invConf := MakeInvoicesConf(ctx, log, &s.settingsClient)

	var expiringPercentage = 0.9
	if invConf.IssueRenewalInvoiceAfter > 0 && invConf.IssueRenewalInvoiceAfter <= 1 {
		expiringPercentage = invConf.IssueRenewalInvoiceAfter
	} else {
		log.Warn("Wrong expiring percentage in config. Using default value", zap.Float64("percentage", expiringPercentage))
	}
	expiringTime := func(expiringAt, period int64) int64 {
		return expiringAt - int64(float64(period)*(1.0-expiringPercentage))
	}

	counter := 0
	// Process every instance
	for _, srv := range list.Result {
		for _, ig := range srv.GetInstancesGroups() {
			for _, i := range ig.GetInstances() {
				log := log.With(zap.String("instance", i.GetUuid()))

				if i.GetStatus() == spb.NoCloudStatus_DEL ||
					i.GetState().GetState() == stpb.NoCloudState_PENDING ||
					i.GetState().GetState() == stpb.NoCloudState_DELETED ||
					i.GetStatus() == spb.NoCloudStatus_INIT {
					continue
				}

				now := time.Now().Unix()

				if i.Config == nil {
					i.Config = map[string]*structpb.Value{}
				}

				lastRenewInvoice, ok := i.Config["renew_invoice"]
				if ok {
					inv, err := s.invoices.Get(ctx, lastRenewInvoice.GetStringValue())
					if err != nil {
						log.Error("Error getting invoice. Skipping invoice creation", zap.Error(err))
						continue
					}
					if inv.GetDeadline() > now && inv.GetStatus() == pb.BillingStatus_UNPAID {
						log.Info("Last renew invoice for this instance didn't expired yet. Skipping invoice creation", zap.String("invoice", inv.GetUuid()))
						continue
					}
				} else {
					log.Info("No last renew invoice for this instance. Continue")
				}

				res, err := s.instances.GetGroup(ctx, driver.NewDocumentID(schema.INSTANCES_COL, i.GetUuid()).String())
				if err != nil || res.Group == nil || res.SP == nil {
					log.Error("Error getting instance group or sp", zap.Error(err), zap.Any("group", res.Group), zap.Any("sp", res.SP))
					continue
				}

				log = log.With(zap.String("driver", res.SP.GetType()))

				client, ok := s.drivers[res.SP.GetType()]
				if !ok {
					log.Error("Error getting driver. Driver not registered")
					continue
				}

				resp, err := client.GetExpiration(ctx, &dpb.GetExpirationRequest{
					Instance:         i,
					ServicesProvider: res.SP,
				})
				if err != nil {
					log.Error("Error getting expiration", zap.Error(err))
					continue
				}
				records := resp.GetRecords()
				log.Info("Got expiration records", zap.Any("records", records))

				// If got expiring records calculate renewal for all instance billing resources
				var expiring = false
				for _, r := range records {
					log := log.With(zap.Any("record", r))
					if expiringTime(r.Expires, r.Period) > now {
						log.Info("Not expiring yet")
						continue
					}
					if r.Product == "" {
						log.Info("Ignoring non product record")
						continue
					}
					expiring = true
				}

				if !expiring {
					log.Info("Instance is not expiring. Skipping invoice creation.")
					continue
				}

				ctx = context.WithValue(ctx, nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY)
				ctx = context.WithValue(ctx, "create_as_draft", true)
				invResp, err := s.CreateRenewalInvoice(ctx, connect.NewRequest(&pb.CreateRenewalInvoiceRequest{
					Instance: i.GetUuid(),
				}))
				if err != nil {
					log.Error("Failed to create invoice", zap.Error(err))
					continue
				}

				log.Debug("Invoice created", zap.String("invoice", invResp.Msg.GetUuid()))

				log.Debug("Updating instance")
				i.Config["renew_invoice"] = structpb.NewStringValue(invResp.Msg.GetUuid())
				if err = s.instances.Update(ctx, "", proto.Clone(i).(*ipb.Instance), proto.Clone(i).(*ipb.Instance)); err != nil {
					log.Error("Failed to update instance", zap.Error(err))
					continue
				}

				log.Debug("Finished")
				counter++
			}
		}
	}

	log.Info("Routine finished", zap.Int("invoices_created", counter))
}

const updateDataQuery = `
UPDATE DOCUMENT(@@collection, @key) WITH { data: @data } IN @@collection
`

func (s *BillingServiceServer) IssueInvoicesRoutine(ctx context.Context) {
	log := s.log.Named("IssueInvoicesRoutine")

	ch, err := s.rbmq.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel", zap.Error(err))
	}
	defer ch.Close()

start:

	routineConf := MakeRoutineConf(ctx, log, &s.settingsClient)
	roundingConf := MakeRoundingConf(ctx, log, &s.settingsClient)
	currencyConf := MakeCurrencyConf(ctx, log, &s.settingsClient)
	iPub := NewInstanceDataPublisher(ch)

	upd := make(chan bool, 1)
	go sc.Subscribe([]string{currencyKey}, upd)

	log.Info("Got Configuration", zap.Any("currency", currencyConf), zap.Any("routine", routineConf), zap.Any("rounding", roundingConf))

	ticker := time.NewTicker(time.Second * time.Duration(routineConf.Frequency))
	tick := time.Now()
	for {
		s.inv.Status.Status = hpb.Status_RUNNING
		s.inv.Status.Error = nil

		log.Info("Entering new Iteration", zap.Time("ts", tick))
		s.InvoiceExpiringInstances(ctx, log, tick, currencyConf, roundingConf, iPub)

		s.inv.LastExecution = tick.Format("2006-01-02T15:04:05Z07:00")
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
		"@transactions":  schema.TRANSACTIONS_COL,
		"@instances":     schema.INSTANCES_COL,
		"@services":      schema.SERVICES_COL,
		"@records":       schema.RECORDS_COL,
		"@accounts":      schema.ACCOUNTS_COL,
		"@addons":        schema.ADDONS_COL,
		"@billing_plans": schema.BILLING_PLANS_COL,
		"permissions":    schema.PERMISSIONS_GRAPH.Name,
		"now":            tick.Unix(),
		"graph":          schema.BILLING_GRAPH.Name,
		"currencies":     schema.CUR_COL,
		"currency":       currencyConf.Currency,
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
	suspConf := MakeSuspendConf(ctx, log, &s.settingsClient)
	routineConf := MakeRoutineConf(ctx, log, &s.settingsClient)

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
			if _, err := s.accClient.Suspend(ctx, &accpb.SuspendRequest{Uuid: acc.GetUuid()}); err != nil {
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
			if _, err := s.accClient.Unsuspend(ctx, &accpb.UnsuspendRequest{Uuid: acc.GetUuid()}); err != nil {
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

	routineConf := MakeRoutineConf(ctx, log, &s.settingsClient)
	roundingConf := MakeRoundingConf(ctx, log, &s.settingsClient)
	currencyConf := MakeCurrencyConf(ctx, log, &s.settingsClient)

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
        filter LENGTH(acc.account_owner) == 0
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

LET subs = (
    FOR acc IN UNION_DISTINCT(global, local)
        FILTER IS_ARRAY(acc.subaccounts)
        FOR sub IN acc.subaccounts
           RETURN DOCUMENT(Accounts, sub)
)

FOR acc IN union_distinct(local, global, subs)
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
        FILTER LENGTH(acc.account_owner) == 0
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

LET subs = (
    FOR acc IN UNION_DISTINCT(global, local, extra)
        FILTER IS_ARRAY(acc.subaccounts)
        FOR sub IN acc.subaccounts
           RETURN DOCUMENT(Accounts, sub)
)

FOR acc IN UNION_DISTINCT(global, local, extra, subs)
    RETURN MERGE(acc, {uuid:acc._key})
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
			RETURN LENGTH(node.account_owner) > 0 ? DOCUMENT(@@accounts, node.account_owner) : node
    )

	// Prefer user currency to default if present
	LET currency = account.currency != null ? account.currency : @currency

    LET records = ( // Collect all unprocessed records
        FOR record IN @@records
        FILTER record.exec <= @now
        FILTER !record.processed
        FILTER record.instance IN instances

        LET instance = DOCUMENT(@@instances, record.instance)
        LET bp = DOCUMENT(@@billing_plans, instance.billing_plan.uuid)
        LET resources = bp.resources == null ? [] : bp.resources
        LET addon = DOCUMENT(@@addons, record.addon)
        LET product_period = bp.products[instance.product].period
        LET item_price = record.product == null ? (record.resource == null ? addon.periods[product_period] : LAST(FOR res in resources FILTER res.key == record.resource return res).price) : bp.products[record.product].price
        LET final_price = record.cost > 0 ? record.cost : item_price

		LET rate = PRODUCT(
			FOR vertex, edge IN OUTBOUND
			SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(record.currency.id)))
			TO DOCUMENT(CONCAT(@currencies, "/", currency.id))
			GRAPH @graph
			FILTER edge
				RETURN edge.rate
		)
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
