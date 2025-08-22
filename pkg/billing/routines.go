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
	"github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	"github.com/slntopp/nocloud-proto/events"
	elpb "github.com/slntopp/nocloud-proto/events_logging"
	instancespb "github.com/slntopp/nocloud-proto/instances"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/structpb"
	"sync"
	"time"

	sc "github.com/slntopp/nocloud/pkg/settings/client"

	hpb "github.com/slntopp/nocloud-proto/health"
	accpb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("LOW_FREQ_ROUTINES_FREQ_SECONDS", 7200)
	lowFrequentRoutinesFreqSeconds = viper.GetInt("LOW_FREQ_ROUTINES_FREQ_SECONDS")
}

var lowFrequentRoutinesFreqSeconds = 7200

func (s *BillingServiceServer) RoutinesState() []*hpb.RoutineStatus {
	return []*hpb.RoutineStatus{
		s.gen, s.proc, s.sus, s.cron,
	}
}

const updateDataQuery = `
UPDATE DOCUMENT(@@collection, @key) WITH { data: @data } IN @@collection
`

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

func (s *BillingServiceServer) SuspendAccountsRoutine(_ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.WithoutCancel(_ctx)

	log := s.log.Named("AccountSuspendRoutine")

start:
	suspConf := MakeSuspendConf(log, &s.settingsClient)
	routineConf := MakeRoutineConf(log, &s.settingsClient)

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
		case <-_ctx.Done():
			log.Info("Context is done. Quitting")
			return
		case tick = <-ticker.C:
			continue
		case <-upd:
			log.Info("New Configuration Received, restarting Routine")
			goto start
		}
	}

}

func (s *BillingServiceServer) GenTransactionsRoutine(_ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.WithoutCancel(_ctx)

	log := s.log.Named("GenerateTransactionsRoutine")

start:

	routineConf := MakeRoutineConf(log, &s.settingsClient)
	roundingConf := MakeRoundingConf(log, &s.settingsClient)
	currencyConf := MakeCurrencyConf(log, &s.settingsClient)

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
		case <-_ctx.Done():
			log.Info("Context is done. Quitting")
			return
		case tick = <-ticker.C:
			continue
		case <-upd:
			log.Info("New Configuration Received, restarting Routine")
			goto start
		}
	}
}

func (s *BillingServiceServer) AutoPayInvoicesRoutine(_ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.WithoutCancel(_ctx)
	log := s.log.Named("AutoPayInvoicesRoutine")

	const maxRetries = 3
	var retries = 0
start:
	roundingConf := MakeRoundingConf(log, &s.settingsClient)
	currencyConf := MakeCurrencyConf(log, &s.settingsClient)

	upd := make(chan bool, 1)

	log.Info("Got Configuration", zap.Any("currency", currencyConf), zap.Any("rounding", roundingConf))

	ticker := time.NewTicker(time.Second * time.Duration(lowFrequentRoutinesFreqSeconds))
	tick := time.Now()

	for {
		log.Info("Entering new Iteration", zap.Time("ts", tick))
		err := s.AutoPayInvoices(ctx, log, tick, currencyConf, roundingConf)
		if err != nil && retries < maxRetries {
			log.Info("Retrying...")
			time.Sleep(time.Second * 5)
			retries++
			goto start
		}

		retries = 0
		select {
		case <-_ctx.Done():
			log.Info("Context is done. Quitting")
			return
		case tick = <-ticker.C:
			continue
		case <-upd:
			log.Info("New Configuration Received, restarting Routine")
			goto start
		}
	}
}

const lastAutoPaymentAttemptKey = "last_auto_payment_attempt"
const autoPaymentAttemptsKey = "auto_payment_attempts_count"

func (s *BillingServiceServer) getInstanceExpiration(ctx context.Context, inst *instancespb.ResponseInstance) (int64, int64, error) {
	client, ok := s.drivers[inst.Type]
	if !ok {
		return 0, 0, fmt.Errorf("driver not found for type: %s", inst.Type)
	}
	resp, err := client.GetExpiration(ctx, &vanilla.GetExpirationRequest{
		Instance:         inst.Instance,
		ServicesProvider: &sppb.ServicesProvider{Uuid: inst.Sp},
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get expiration records: %w", err)
	}
	expires := int64(0)
	period := int64(0)
	for _, rec := range resp.Records {
		if rec.Product != "" {
			expires = rec.Expires
			period = rec.Period
			break
		}
	}
	if expires == 0 {
		return 0, 0, fmt.Errorf("expiration is 0")
	}
	return expires, period, nil
}

func (s *BillingServiceServer) AutoPayInvoices(ctx context.Context, log *zap.Logger, _ time.Time,
	_ CurrencyConf, _ RoundingConf) error {
	log.Info("Starting Auto Pay Invoices Routine")
	rootToken, _ := ctx.Value(nocloud.NoCloudToken).(string)

	// Fetch unpaid renewal invoices
	unpaidFilter, _ := structpb.NewList([]any{int(pb.BillingStatus_UNPAID)})
	actionFilter, _ := structpb.NewList([]any{int(pb.ActionType_INSTANCE_RENEWAL)})
	invReq := connect.NewRequest(&pb.GetInvoicesRequest{
		Filters: map[string]*structpb.Value{
			"status": structpb.NewListValue(unpaidFilter),
			"type":   structpb.NewListValue(actionFilter),
		},
	})
	invResp, err := s.GetInvoices(ctx, invReq)
	if err != nil {
		log.Error("Failed to list invoices", zap.Error(err))
		return err
	}
	log.Debug("Fetched invoices", zap.Any("len", len(invResp.Msg.Pool)))
	var (
		passedInvoices     = map[string]*pb.Invoice{}
		instanceToInvoice  = map[string][]string{}
		instancesToFetch   = make([]any, 0)
		instancesSearchMap = map[string]*instancespb.ResponseInstance{}
	)
	unmarkAllByInstance := func(inst string) {
		for _, inv := range instanceToInvoice[inst] {
			delete(passedInvoices, inv)
		}
		delete(instanceToInvoice, inst)
	}
	for _, inv := range invResp.Msg.Pool {
		if inv.Meta == nil ||
			inv.Meta["auto_created"] == nil ||
			!inv.Meta["auto_created"].GetBoolValue() {
			continue
		}
		if int(inv.Meta[autoPaymentAttemptsKey].GetNumberValue()) >= 1 {
			continue
		}
		passedInvoices[inv.GetUuid()] = inv
		for _, inst := range inv.GetInstances() {
			if instanceToInvoice[inst] == nil {
				instanceToInvoice[inst] = make([]string, 0)
			}
			instanceToInvoice[inst] = append(instanceToInvoice[inst], inv.GetUuid())
			instancesToFetch = append(instancesToFetch, inst)
		}
	}

	if len(instancesToFetch) == 0 {
		log.Debug("No instances to fetch. Skipping...")
		return nil
	}
	// Fetch required instances
	var (
		page  = uint64(1)
		limit = uint64(100_000)
	)
	instancesFilter, _ := structpb.NewList(instancesToFetch)
	instancesStatusFilter, _ := structpb.NewList([]any{statuses.NoCloudStatus_UP, statuses.NoCloudStatus_INIT,
		statuses.NoCloudStatus_SUS, statuses.NoCloudStatus_UNSPECIFIED})
	instReq := connect.NewRequest(&instancespb.ListInstancesRequest{
		Page:  &page,
		Limit: &limit,
		Filters: map[string]*structpb.Value{
			"uuids":  structpb.NewListValue(instancesFilter),
			"status": structpb.NewListValue(instancesStatusFilter),
		},
	})
	instReq.Header().Set("Authorization", "Bearer "+rootToken)
	instResp, err := s.instancesClient.List(ctx, instReq)
	if err != nil {
		log.Error("Failed to list instances", zap.Error(err))
		return err
	}
	log.Debug("Obtained instances", zap.Any("len", len(instResp.Msg.Pool)))

	// Filter instances and invoices that are not fitted for auto payment
	// 1. Filter if at least 1 instance is not on auto-renew
	var passedInstances = make([]*instancespb.ResponseInstance, 0)
	for _, instValue := range instResp.Msg.Pool {
		inst := instValue.Instance
		if inst.Meta == nil || !inst.Meta.GetAutoRenew() {
			unmarkAllByInstance(inst.GetUuid())
			continue
		}
		passedInstances = append(passedInstances, instValue)
		instancesSearchMap[inst.GetUuid()] = instValue
	}
	// 2. Fetch and compare expiration data with expiration data in current invoices
	log.Debug("Auto renew candidates: instances", zap.Any("len", len(passedInstances)))
	log.Debug("Auto renew candidates: invoices", zap.Any("len", len(passedInvoices)), zap.Any("uuids", maps.Keys(passedInvoices)))
	now := time.Now().Unix()
	isTimeToRenew := func(expiration, period, now int64) bool {
		const day = int64(24 * time.Hour / time.Second)
		const threeDays = 3 * day
		if period <= 0 {
			return now >= expiration
		}
		if period < threeDays {
			start := expiration - period
			elapsed := now - start
			return elapsed*100 >= period*80
		}
		if now >= expiration {
			return true
		}
		expDayStart := expiration - (expiration % day)
		nowDayStart := now - (now % day)
		if nowDayStart != expDayStart {
			return false
		}
		hourUTC := (now % day) / int64(time.Hour/time.Second)
		return hourUTC >= 6 && hourUTC < 21
	}
	for _, instValue := range passedInstances {
		inst := instValue.Instance
		expiration, period, err := s.getInstanceExpiration(ctx, instValue)
		if err != nil {
			log.Warn("Failed to get instance expiration", zap.Error(err), zap.String("instance", inst.GetUuid()))
			unmarkAllByInstance(inst.GetUuid())
			continue
		}
		log.Debug("Got expiration for instance", zap.String("instance", inst.GetUuid()), zap.Int64("expiration", expiration), zap.Int64("period", period))
		if period == 0 || !isTimeToRenew(expiration, period, now) {
			unmarkAllByInstance(inst.GetUuid())
			continue
		}
		for _, invID := range instanceToInvoice[inst.GetUuid()] {
			invoice, ok := passedInvoices[invID]
			if !ok {
				continue
			}
			inv := graph.Invoice{Invoice: invoice}
			billingData := inv.BillingData()
			if billingData == nil || billingData.RenewalData == nil ||
				billingData.RenewalData[inst.GetUuid()].ExpirationTs != expiration {
				delete(passedInvoices, invoice.GetUuid())
			}
		}
	}

	log.Debug("Auto renew: picked invoices", zap.Any("len", len(passedInvoices)), zap.Any("uuids", maps.Keys(passedInvoices)))
	// Now passedInvoices contains invoices, that are ready for auto-payment
	now = time.Now().Unix()
	for _, inv := range passedInvoices {
		log.Debug("Trying to auto-pay invoice", zap.Any("uuid", inv.GetUuid()))
		if inv.Total <= 0 {
			continue
		}
		if _, err = s.PayWithBalance(ctxWithRoot(ctx), connect.NewRequest(&pb.PayWithBalanceRequest{
			InvoiceUuid: inv.GetUuid(),
		})); err != nil {
			log.Warn("Failed to auto-pay INSTANCE_RENEW invoice from user balance", zap.Error(err), zap.String("invoice", inv.GetUuid()))
			// Send auto payment failure events
			for _, id := range inv.GetInstances() {
				d := instancesSearchMap[id]
				if d == nil {
					continue
				}
				inst := d.Instance
				if inst == nil {
					continue
				}
				// Add extra fields
				eventData := map[string]*structpb.Value{}
				eventData["instance"] = structpb.NewStringValue(inst.GetTitle())
				eventData["instance_uuid"] = structpb.NewStringValue(inst.GetUuid())
				bp := inst.GetBillingPlan()
				if bp != nil && bp.GetProducts() != nil {
					bpProducts := bp.GetProducts()
					instProduct, _ := bpProducts[inst.GetProduct()]
					if instProduct != nil {
						eventData["product"] = structpb.NewStringValue(instProduct.GetTitle())
					}
				}
				if inst.Data != nil {
					eventData["next_payment_date"] = structpb.NewNumberValue(inst.Data["next_payment_date"].GetNumberValue())
				}
				// Send
				if _, err = s.eventsClient.Publish(ctx, &events.Event{
					Type: "email",
					Uuid: d.Account,
					Key:  "auto_payment_failure",
					Data: eventData,
					Ts:   now,
				}); err != nil {
					log.Error("Failed to send auto_payment_failure event", zap.Error(err))
				}
			}
		} else {
			nocloud.Log(log, &elpb.Event{
				Uuid:      inv.GetUuid(),
				Entity:    "Invoices",
				Action:    "auto_payment",
				Scope:     "database",
				Rc:        0,
				Ts:        time.Now().Unix(),
				Snapshot:  &elpb.Snapshot{Diff: ""},
				Requestor: schema.ROOT_ACCOUNT_KEY,
			})
		}
		inv.Meta[lastAutoPaymentAttemptKey] = structpb.NewNumberValue(float64(now))
		inv.Meta[autoPaymentAttemptsKey] = structpb.NewNumberValue(inv.Meta[autoPaymentAttemptsKey].GetNumberValue() + 1)
		if err = s.invoices.Patch(ctx, inv.GetUuid(), map[string]interface{}{
			"meta": inv.Meta,
		}); err != nil {
			log.Error("Failed to patch invoice after payment attempt", zap.Error(err))
		}
	}

	log.Info("Finished Auto Pay Invoices Routine")
	return nil
}

func (s *BillingServiceServer) SendLowCreditsNotificationsRoutine(_ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.WithoutCancel(_ctx)
	log := s.log.Named("SendLowCreditsNotificationsRoutine")

	const maxRetries = 3
	var retries = 0
start:
	roundingConf := MakeRoundingConf(log, &s.settingsClient)
	currencyConf := MakeCurrencyConf(log, &s.settingsClient)

	upd := make(chan bool, 1)

	log.Info("Got Configuration", zap.Any("currency", currencyConf), zap.Any("rounding", roundingConf))

	ticker := time.NewTicker(time.Second * time.Duration(lowFrequentRoutinesFreqSeconds))
	tick := time.Now()

	for {
		log.Info("Entering new Iteration", zap.Time("ts", tick))
		err := s.SendLowCreditsNotifications(ctx, log, tick, currencyConf, roundingConf)
		if err != nil && retries < maxRetries {
			log.Info("Retrying...")
			time.Sleep(time.Second * 5)
			retries++
			goto start
		}

		retries = 0
		select {
		case <-_ctx.Done():
			log.Info("Context is done. Quitting")
			return
		case tick = <-ticker.C:
			continue
		case <-upd:
			log.Info("New Configuration Received, restarting Routine")
			goto start
		}
	}
}

func (s *BillingServiceServer) SendLowCreditsNotifications(ctx context.Context, log *zap.Logger, _ time.Time,
	_ CurrencyConf, _ RoundingConf) error {
	log.Info("Starting Send Low Credits Notify Routine")
	rootToken, _ := ctx.Value(nocloud.NoCloudToken).(string)

	// Fetch all auto_renew plans
	plansReq := connect.NewRequest(&pb.ListRequest{
		Anonymously: false,
		ShowDeleted: true,
		Filters: map[string]*structpb.Value{
			"auto_payment": structpb.NewBoolValue(true),
		},
	})
	plansResp, err := s.ListPlans(ctx, plansReq)
	if err != nil {
		log.Error("Failed to list plans", zap.Error(err))
		return err
	}
	log.Debug("Obtained plans", zap.Any("len", len(plansResp.Msg.Pool)))
	if len(plansResp.Msg.Pool) == 0 {
		log.Debug("No auto_payment plans. Quitting...")
		return nil
	}

	var (
		instancesPlansFilter = make([]any, 0)
	)
	for _, p := range plansResp.Msg.Pool {
		instancesPlansFilter = append(instancesPlansFilter, p.GetUuid())
	}

	// Fetch required instances
	var (
		page  = uint64(1)
		limit = uint64(100_000)
	)
	instancesFilter, _ := structpb.NewList(instancesPlansFilter)
	instancesStatusFilter, _ := structpb.NewList([]any{statuses.NoCloudStatus_UP, statuses.NoCloudStatus_INIT,
		statuses.NoCloudStatus_SUS, statuses.NoCloudStatus_UNSPECIFIED})
	instReq := connect.NewRequest(&instancespb.ListInstancesRequest{
		Page:  &page,
		Limit: &limit,
		Filters: map[string]*structpb.Value{
			"billing_plan": structpb.NewListValue(instancesFilter),
			"status":       structpb.NewListValue(instancesStatusFilter),
			"started":      structpb.NewBoolValue(true),
		},
	})
	instReq.Header().Set("Authorization", "Bearer "+rootToken)
	instResp, err := s.instancesClient.List(ctx, instReq)
	if err != nil {
		log.Error("Failed to list instances", zap.Error(err))
		return err
	}
	log.Debug("Obtained instances", zap.Any("len", len(instResp.Msg.Pool)))
	if len(instResp.Msg.Pool) == 0 {
		log.Debug("No active instances, connected with auto_payment plans. Quitting...")
		return nil
	}

	var (
		accountsCandidates = make([]any, 0)
	)
	for _, instValue := range instResp.Msg.Pool {
		accountsCandidates = append(accountsCandidates, instValue.Account)
	}

	accountsFilter, _ := structpb.NewList(accountsCandidates)
	accounts, _, _, err := s.accounts.ListImproved(ctx, schema.ROOT_ACCOUNT_KEY, 100, 0, 0, "", "", map[string]*structpb.Value{
		"uuid": structpb.NewListValue(accountsFilter),
	})
	if err != nil {
		log.Error("Failed to list accounts", zap.Error(err))
		return err
	}
	log.Debug("Obtained accounts", zap.Any("len", len(accounts)))

	currencies, err := s.currencies.GetCurrencies(ctx, true)
	if err != nil {
		log.Error("Failed to get currencies", zap.Error(err))
		return err
	}
	var defaultCurrency *pb.Currency
	for _, c := range currencies {
		if c.Default {
			defaultCurrency = c
			break
		}
	}
	if defaultCurrency == nil {
		log.Error("Default currency not found")
		return fmt.Errorf("default currency not found")
	}
	allRates, err := s.currencies.GetExchangeRates(ctx)
	if err != nil {
		log.Error("Failed to get exchange rates", zap.Error(err))
		return err
	}

	// Now accounts contains valid candidates for low credit events
	for _, account := range accounts {
		balance := float64(0)
		if account.Balance != nil {
			balance = *account.Balance
		}
		balanceInDefault, ok := convertWithRate(allRates, account.Currency, &pb.Currency{Id: 0}, balance)
		if !ok {
			log.Error("Failed to convert client balance to default currency", zap.Any("client_currency", account.Currency))
			return fmt.Errorf("cannot convert client's balance")
		}
		s.processLowCreditEvent(ctx, log, &account, balanceInDefault, 5, defaultCurrency.Code, 1)
		s.processLowCreditEvent(ctx, log, &account, balanceInDefault, 2, defaultCurrency.Code, 2)
	}

	log.Info("Finished Send Low Credits Notify Routine")
	return nil
}

func (s *BillingServiceServer) processLowCreditEvent(ctx context.Context, log *zap.Logger,
	account *graph.Account, convertedBalance float64,
	limit float64, defaultCurrencyCode string, eventNum int) {
	ensure(&account.Balance)
	ensure(&account.Meta)
	ensure(&account.Meta.Notifications)
	ensure(&account.Meta.Notifications.FirstBalanceNotify)
	ensure(&account.Meta.Notifications.SecondBalanceNotify)
	ensure(&account.Meta.Notifications.SecondBalanceNotify.Base)
	ensure(&account.Meta.Notifications.FirstBalanceNotify.Base)
	ensure(&account.Meta.Notifications.FirstBalanceNotify.Invalidated)
	ensure(&account.Meta.Notifications.SecondBalanceNotify.Invalidated)
	ensure(&account.Meta.Notifications.FirstBalanceNotify.Base.Disabled)
	ensure(&account.Meta.Notifications.SecondBalanceNotify.Base.Disabled)
	if convertedBalance == 0 {
		return
	}
	var (
		eventKey     = "low_credits_first"
		notification = account.Meta.Notifications.FirstBalanceNotify
		threshold    = limit
	)
	if eventNum == 2 {
		eventKey = "low_credits_second"
		notification = account.Meta.Notifications.SecondBalanceNotify
	}
	if notification.Threshold != nil {
		threshold = *notification.Threshold
	}
	if *notification.Invalidated || convertedBalance >= threshold || *notification.Base.Disabled {
		return
	}

	var code = defaultCurrencyCode
	if account.Currency != nil {
		code = account.Currency.Code
	}
	eventData := map[string]*structpb.Value{
		"balance_amount": structpb.NewNumberValue(*account.Balance),
		"currency_code":  structpb.NewStringValue(code),
	}
	if _, err := s.eventsClient.Publish(ctx, &events.Event{
		Type: "email",
		Uuid: account.Uuid,
		Key:  eventKey,
		Data: eventData,
		Ts:   time.Now().Unix(),
	}); err != nil {
		log.Error(fmt.Sprintf("Failed to send %s event", eventKey), zap.Error(err))
		return
	}
	log.Info(fmt.Sprintf("Event %s was successfully sent for %s account", eventKey, account.GetUuid()))
	notification.Invalidated = ptr(true)
	notification.Base.LastNotification = time.Now().Unix()
	notification.Base.SentNotifications += 1
	if err := s.accounts.Update(ctx, *account, map[string]interface{}{
		"meta": account.Meta,
	}); err != nil {
		log.Error(fmt.Sprintf("Failed to invalidate account's event: %s", account.GetUuid()), zap.Error(err))
	}
}

func ensure[T any](p **T) {
	if *p == nil {
		*p = new(T)
	}
}

func convertWithRate(rates []*pb.GetExchangeRateResponse, from *pb.Currency, to *pb.Currency, amount float64) (float64, bool) {
	var fromId, toId int32
	if from != nil {
		fromId = from.Id
	}
	if to != nil {
		toId = to.Id
	}
	if fromId == toId {
		return amount, true
	}
	for _, r := range rates {
		if r.To == nil {
			r.To = &pb.Currency{}
		}
		if r.From == nil {
			r.From = &pb.Currency{}
		}
		if r.To.Id == toId && r.From.Id == fromId {
			return amount * r.Rate, true
		}
	}
	return amount, false
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
        FILTER TO_NUMBER(acc.balance) < TO_NUMBER(conf['limit'])
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
        FILTER acc.suspend_conf['limit'] && acc.balance < acc.suspend_conf['limit']
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
        FOR record IN @@records
        FILTER !record.processed && record.instance && record.instance != "" && record.exec <= @now

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
            service: service._key,
            records: [record._key],
            total: cost,
	    	meta: {type: "transaction"},
        } IN @@transactions RETURN NEW
`

const processTransactions = `
FOR t IN @@transactions // Iterate over Transactions
FILTER t.exec != null
FILTER t.exec <= @now
FILTER !t.processed
    LET account = t.account != "" ? DOCUMENT(CONCAT(@accounts, "/", t.account)) : null
	// Prefer user currency to default if present
	FILTER account != null
	LET currency = account.currency != null ? account.currency : @currency
	LET rate = PRODUCT(
		FOR vertex, edge IN OUTBOUND
		SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(t.currency.id)))
		TO DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(currency.id)))
		GRAPH @graph
		FILTER edge
			RETURN edge.rate
	)
	LET total = t.total * rate

	FOR r in t.records
        FILTER r != "" && DOCUMENT(@@records, r) != null
		UPDATE r WITH {meta: {transaction: t._key, payment_date: @now}} in @@records

    UPDATE account WITH { balance: account.balance - t.total * rate} IN @@accounts
    UPDATE t WITH { 
		processed: true, 
		proc: @now,
		total: total,
		currency: currency
	} IN @@transactions
`
