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
	"time"

	sc "github.com/slntopp/nocloud/pkg/settings/client"

	epb "github.com/slntopp/nocloud-proto/events"
	hpb "github.com/slntopp/nocloud-proto/health"
	regpb "github.com/slntopp/nocloud-proto/registry"
	accpb "github.com/slntopp/nocloud-proto/registry/accounts"
	settingspb "github.com/slntopp/nocloud-proto/settings"
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
        filter acc.suspend['limit'] && (acc.balance > acc.suspend['limit'])
        return acc
)
    
let global = (
    for acc in candidates
        filter acc.balance > conf['limit']
        filter acc.balance > acc.suspend['limit']
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
        FILTER acc.balance <= conf['limit']
        FILTER (acc.balance - acc.suspend_conf['limit']) <= 0
        RETURN acc
)

LET extra = (
    FOR acc IN candidates
        FILTER conf.is_extra_enabled
        FILTER acc.balance <= conf.extra_limit
        RETURN acc
)

LET local = (
    FOR acc IN candidates
		FILTER now_matching
        FILTER acc.balance <= acc.suspend_conf['limit']
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
		LET rate = PRODUCT(
			FOR vertex, edge IN OUTBOUND
			SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(record.currency)))
			TO DOCUMENT(CONCAT(@currencies, "/", currency))
			GRAPH @graph
			FILTER edge
				RETURN edge.rate
		)
        LET total = record.total * rate
            UPDATE record._key WITH { 
				processed: true, 
				total: total,
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
        total: SUM(records[*].total), // Calculate Total
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
		SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(t.currency)))
		TO DOCUMENT(CONCAT(@currencies, "/", currency))
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
