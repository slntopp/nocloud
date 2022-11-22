/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

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
	"time"

	hpb "github.com/slntopp/nocloud-proto/health"
	regpb "github.com/slntopp/nocloud-proto/registry"
	accpb "github.com/slntopp/nocloud-proto/registry/accounts"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	sc "github.com/slntopp/nocloud/pkg/settings/client"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	settingsClient settingspb.SettingsServiceClient
	accClient      regpb.AccountsServiceClient
)

// Settings Key storing routine conf
const (
	monFreqKey string = "billing-gen-transactions-routine"
	suspKey    string = "global-suspend-conf"
)

type RoutineConf struct {
	Frequency int `json:"freq"` // Frequency in Seconds
}

type SuspendSchedule struct {
	Day  int  `json:"day"`
	Off  bool `json:"off"`
	From int  `json:"from"`
	To   int  `json:"to"`
}

type SuspendConf struct {
	AutoResume     bool              `json:"auto_resume"`
	IsEnabled      bool              `json:"is_enabled"`
	Limit          float64           `json:"limit"`
	Schedule       []SuspendSchedule `json:"schedule"`
	IsExtraEnabled bool              `json:"is_extra_enabled"`
	ExtraLimit     float64           `json:"extra_limit"`
}

var (
	routineSetting = &sc.Setting[RoutineConf]{
		Value: RoutineConf{
			Frequency: 60,
		},
		Description: "Transactions Generating and Processing Routine Configuration",
		Public:      false,
	}
	suspendedSetting = &sc.Setting[SuspendConf]{
		Value: SuspendConf{
			AutoResume: true,
			IsEnabled:  true,
			Limit:      10,
			Schedule: []SuspendSchedule{
				{
					Day: 0,
					Off: true,
				},
				{
					Day:  1,
					From: 10,
					To:   22,
				},
				{
					Day:  2,
					From: 10,
					To:   22,
				},
				{
					Day:  3,
					From: 10,
					To:   22,
				},
				{
					Day:  4,
					From: 10,
					To:   22,
				},
				{
					Day:  5,
					From: 10,
					To:   22,
				},
				{
					Day: 6,
					Off: true,
				},
			},
			// IsExtraEnabled: true,
			// ExtraLimit: -100,
		},
	}
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("SETTINGS_HOST", "settings:8000")
	viper.SetDefault("REGISTRY_HOST", "registry:8000")
	settingsHost := viper.GetString("SETTINGS_HOST")
	registryHost := viper.GetString("REGISTRY_HOST")

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
}

func MakeRoutineConf(ctx context.Context, log *zap.Logger) (conf RoutineConf) {
	sc.Setup(log, ctx, &settingsClient)

	if err := sc.Fetch(monFreqKey, &conf, routineSetting); err != nil {
		conf = routineSetting.Value
	}

	return conf
}

func MakeSuspendConf(ctx context.Context, log *zap.Logger) (conf SuspendConf) {
	sc.Setup(log, ctx, &settingsClient)

	if err := sc.Fetch(suspKey, &conf, suspendedSetting); err != nil {
		conf = suspendedSetting.Value
	}

	return conf
}

func (s *BillingServiceServer) GenTransactionsRoutineState() []*hpb.RoutineStatus {
	return []*hpb.RoutineStatus{
		s.gen, s.proc,
	}
}

func (s *BillingServiceServer) GenTransactionsRoutine(ctx context.Context) {
	log := s.log.Named("Routine")

	routineConf := MakeRoutineConf(ctx, log)
	log.Info("Got Configuration", zap.Any("routine", routineConf))

	ticker := time.NewTicker(time.Second * time.Duration(routineConf.Frequency))
	tick := time.Now()
	for {
		suspConf := MakeSuspendConf(ctx, log)
		log.Info("Got Configuration", zap.Any("suspend", suspConf))

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
			"accounts":      schema.ACCOUNTS_COL,
			"now":           tick.Unix(),
		})
		if err != nil {
			log.Error("Error Processing Transactions", zap.Error(err))
			s.proc.Status.Status = hpb.Status_HASERRS
			err_s := err.Error()
			s.proc.Status.Error = &err_s
		}

		cursor, err := s.db.Query(ctx, accToSuspend, map[string]interface{}{
			"conf": suspConf,
		})
		if err != nil {
			log.Error("Error Quering Accounts to Suspend", zap.Error(err))
		}

		for cursor.HasMore() {
			acc := &accpb.Account{}
			meta, err := cursor.ReadDocument(ctx, &acc)
			if err != nil {
				log.Error("Error Reading Account", zap.Error(err), zap.Any("meta", meta))
				continue
			}
			// log.Debug("acc", zap.String("uuid", acc.GetUuid()), zap.Any("value", acc))
			if _, err := accClient.Suspend(ctx, &accpb.SuspendRequest{Uuid: acc.GetUuid()}); err != nil {
				log.Error("Error Suspending Account", zap.Error(err))
			}
		}

		cursor2, err := s.db.Query(ctx, accToUnsuspend, map[string]interface{}{
			"conf": suspConf,
		})
		if err != nil {
			log.Error("Error Quering Accounts to Unsuspend", zap.Error(err))
		}

		for cursor2.HasMore() {
			acc := &accpb.Account{}
			meta, err := cursor2.ReadDocument(ctx, &acc)
			if err != nil {
				log.Error("Error Reading Account", zap.Error(err), zap.Any("meta", meta))
				continue
			}
			// log.Debug("acc", zap.String("uuid", acc.GetUuid()), zap.Any("value", acc))
			if _, err := accClient.Unsuspend(ctx, &accpb.UnsuspendRequest{Uuid: acc.GetUuid()}); err != nil {
				log.Error("Error Unsuspending Account", zap.Error(err))
			}
		}

		s.proc.LastExecution = tick.Format("2006-01-02T15:04:05Z07:00")
		tick = <-ticker.C
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
            RETURN i._key )

    LET account = LAST( // Find Service owner Account
    FOR node, edge, path IN 2
    INBOUND service
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@accounts)
        RETURN node
    )
    
    LET records = ( // Collect all unprocessed records
        FOR record IN @@records
        FILTER record.exec <= @now
        FILTER !record.processed
        FILTER record.instance IN instances
            UPDATE record._key WITH { processed: true } IN @@records RETURN NEW
    )
    
    FILTER LENGTH(records) > 0 // Skip if no Records (no empty Transaction)
    INSERT {
        exec: @now, // Timestamp in seconds
        processed: false,
        account: account._key,
        service: service._key,
        records: records[*]._key,
        total: SUM(records[*].total) // Calculate Total
    } IN @@transactions RETURN NEW
`

const processTransactions = `
FOR t IN @@transactions // Iterate over Transactions
FILTER t.exec <= @now
FILTER !t.processed
    LET account = DOCUMENT(CONCAT(@accounts, "/", t.account))
    UPDATE account WITH { balance: account.balance - t.total } IN @@accounts
    UPDATE t WITH { processed: true, proc: @now } IN @@transactions
`
