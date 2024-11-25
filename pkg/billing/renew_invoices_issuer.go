package billing

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	dpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	ipb "github.com/slntopp/nocloud-proto/instances"
	"github.com/slntopp/nocloud-proto/services"
	"github.com/slntopp/nocloud-proto/states"
	"github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"slices"
	"strings"
	"time"
)

type accountPool struct {
	Account        graph.Account
	Instances      []*ipb.Instance
	Invoices       []*graph.Invoice
	InstExpRecords map[string][]*dpb.ExpirationRecord
}

func (s *BillingServiceServer) InvoiceExpiringInstancesCronJob(ctx context.Context, log *zap.Logger) {
	log = log.Named("InvoicesIssuer")
	log.Info("Starting Invoice Expiring Instances Cron Job")

	var (
		errs     int
		warns    int
		gotPanic bool
		started  bool
		preErrs  int
		preWarns int
		created  int
	)

	accountPools := map[string]*accountPool{}
	invoices := make([]*graph.Invoice, 0)
	const days15 = int64(3600 * 24 * 15)
	const days10 = int64(3600 * 24 * 10)
	invConf := MakeInvoicesConf(ctx, log, &s.settingsClient)
	currConf := MakeCurrencyConf(ctx, log, &s.settingsClient)
	var expiringPercentage = 0.9
	if invConf.IssueRenewalInvoiceAfter > 0 && invConf.IssueRenewalInvoiceAfter <= 1 {
		expiringPercentage = invConf.IssueRenewalInvoiceAfter
	} else {
		log.Warn("Wrong expiring percentage in config. Using default value", zap.Float64("percentage", expiringPercentage))
	}
	isExpiring := func(now, expiringAt, period int64) bool {
		if period > days15 {
			return (expiringAt - days10) < now
		}
		return (expiringAt - int64(float64(period)*(1.0-expiringPercentage))) < now
	}

	list, err := s.services.List(ctx, schema.ROOT_ACCOUNT_KEY, &services.ListRequest{})
	if err != nil {
		log.Error("Error getting services", zap.Error(err))
		preErrs++
		goto end
	}
	log.Debug("Got list of services", zap.Any("count", len(list.Result)))

	for _, srv := range list.Result {
		for _, ig := range srv.GetInstancesGroups() {
			for _, inst := range ig.GetInstances() {
				log := log.With(zap.String("instance", inst.GetUuid()))
				if inst.GetProduct() == "" {
					continue
				}
				acc, err := s.instances.GetInstanceOwner(ctx, inst.GetUuid())
				if err != nil {
					log.Error("Error getting instance owner", zap.Error(err))
					preErrs++
					continue
				}
				instId := driver.NewDocumentID(schema.INSTANCES_COL, inst.GetUuid()).String()
				spResp, err := s.instances.GetGroup(ctx, instId)
				if err != nil {
					log.Error("Error getting instance group and sp", zap.Error(err))
					preErrs++
					continue
				}
				if spResp.SP == nil || spResp.Group == nil {
					log.Error("Error getting instance group and sp. Nil values", zap.Error(err))
					preErrs++
					continue
				}
				if accountPools[acc.GetUuid()] == nil {
					accountPools[acc.GetUuid()] = &accountPool{
						Account:        acc,
						Instances:      []*ipb.Instance{},
						Invoices:       []*graph.Invoice{},
						InstExpRecords: map[string][]*dpb.ExpirationRecord{},
					}
				}
				accountPools[acc.GetUuid()].Instances = append(accountPools[acc.GetUuid()].Instances, inst)
				client, ok := s.drivers[ig.GetType()]
				if !ok {
					log.Error("Driver not found", zap.String("type", ig.GetType()))
					preErrs++
					continue
				}
				expResp, err := client.GetExpiration(ctx, &dpb.GetExpirationRequest{Instance: inst, ServicesProvider: spResp.SP})
				if err != nil {
					log.Error("Error getting instance expirations", zap.Error(err))
					preErrs++
					continue
				}
				accountPools[acc.GetUuid()].InstExpRecords[inst.GetUuid()] = expResp.GetRecords()
			}
		}
	}

	invoices, err = s.invoices.List(ctx, "")
	if err != nil {
		log.Error("Error listing invoices", zap.Error(err))
		preErrs++
		goto end
	}
	for _, inv := range invoices {
		if inv.GetAccount() == "" {
			log.Warn("Not linked invoice found", zap.String("uuid", inv.GetUuid()))
			preWarns++
			continue
		}
		if accountPools[inv.GetAccount()] == nil {
			acc, err := s.accounts.Get(ctx, inv.GetAccount())
			if err != nil {
				log.Error("Error getting account for invoice list", zap.Error(err))
				preErrs++
				continue
			}
			accountPools[inv.GetAccount()] = &accountPool{
				Account:        acc,
				Instances:      []*ipb.Instance{},
				Invoices:       []*graph.Invoice{},
				InstExpRecords: map[string][]*dpb.ExpirationRecord{},
			}
		}
		accountPools[inv.GetAccount()].Invoices = append(accountPools[inv.GetAccount()].Invoices, inv)
	}

	started = true
	for _, pool := range accountPools {
		ok, _errs, _warns := s.processAccountRenewalInvoices(ctx, log, pool, isExpiring, currConf.Currency)
		if _errs < 0 || _warns < 0 {
			gotPanic = true
			continue
		}
		errs += _errs
		warns += _warns
		if ok {
			created++
		}
	}

end:
	// Too lazy to watch logs, so just look at final result to see if everything is ok
	var conclusion string
	var preConclusion string
	if preErrs > 0 {
		preConclusion = "ERRORS"
	} else if preWarns > 0 {
		preConclusion = "WARNS"
	} else {
		preConclusion = "OK"
	}
	if !started {
		conclusion = "DIDN'T EVEN STARTED"
	} else if gotPanic {
		conclusion = "PANIC"
	} else if errs > 0 {
		conclusion = "ERRORS"
	} else if warns > 0 {
		conclusion = "OK"
	} else {
		conclusion = "PERFECT"
	}
	result := fmt.Sprintf("Invoices Cron Job Result - Initialization [%s] Processing [%s]: %d errors, %d warnings, %d init_errors, %d init_warns. Created %d invoices.", preConclusion, conclusion, errs, warns, preErrs, preWarns, created)
	if gotPanic {
		result += " GOT PANIC IN PROCESSING!"
	}
	log.Info(result, zap.Int("errs", errs), zap.Int("warns", warns), zap.Bool("got_panic", gotPanic), zap.Int("init_errs", preErrs), zap.Int("init_warns", preWarns))
}

type instanceExpData struct {
	Instance *ipb.Instance
	ExpireAt int64
	Period   int64
}

func (s *BillingServiceServer) processAccountRenewalInvoices(ctx context.Context, log *zap.Logger, data *accountPool, isExp func(now, expiringAt, period int64) bool, defCurr *pb.Currency) (created bool, errCount int, warnsCount int) {
	log = log.Named("ProcessAccount").With(zap.String("account", data.Account.GetUuid()))
	defer func(errs *int, warns *int) {
		if err := recover(); err != nil {
			log.Error("Recovered from panic", zap.Error(fmt.Errorf("%v", err)))
			*errs = -1
			*warns = -1
		}
	}(&errCount, &warnsCount)

	expData := make([]*instanceExpData, 0)
	instances := filterInstances(data.Instances)
	for _, inst := range instances {
		product, ok := inst.GetBillingPlan().GetProducts()[inst.GetProduct()]
		if !ok {
			log.Warn("Product not found for instance", zap.String("product", inst.GetProduct()), zap.String("instance", inst.GetUuid()))
			warnsCount++
			continue
		}
		var expires int64
		var period int64
		for _, rec := range data.InstExpRecords[inst.GetUuid()] {
			if rec.Product != "" {
				expires = rec.Expires
				period = rec.Period
				break
			}
		}
		if expires == 0 {
			log.Warn("No expiration record found for instance product", zap.String("product", inst.GetProduct()), zap.String("instance", inst.GetUuid()))
			warnsCount++
			continue
		}
		if product.GetKind() == pb.Kind_POSTPAID {
			expires += period
		}
		if product.GetPeriod() != period {
			log.Error("Product period mismatch", zap.String("product", inst.GetProduct()), zap.String("instance", inst.GetUuid()))
			errCount++
			continue
		}
		if !isExp(time.Now().Unix(), expires, period) {
			continue
		}

		// Find at least 1 invoice with billing data on the same timestamp
		found := false
		invoices := filterInvoices(data.Invoices)
		for _, inv := range invoices {
			bData := inv.BillingData()
			if bData != nil && bData.RenewalData != nil {
				if val, ok := bData.RenewalData[inst.GetUuid()]; ok && val.ExpirationTs == expires {
					found = true
					break
				}
			}
		}
		if found {
			continue
		}

		expData = append(expData, &instanceExpData{
			Instance: inst,
			ExpireAt: expires,
			Period:   period,
		})
	}

	if err := s.createRenewalInvoice(ctx, log, &data.Account, expData, defCurr); err != nil {
		if !errors.Is(err, errNothingToRenew) {
			log.Error("Error creating renewal invoice", zap.Error(err))
			errCount++
		}
		return false, errCount, warnsCount
	}

	return true, errCount, warnsCount
}

var errNothingToRenew = fmt.Errorf("nothing to renew")

func (s *BillingServiceServer) createRenewalInvoice(ctx context.Context, log *zap.Logger, _acc *graph.Account, data []*instanceExpData, defCurr *pb.Currency) error {
	now := time.Now().Unix()
	fDateNum := func(d int) string {
		if d < 10 {
			return fmt.Sprintf("0%d", d)
		}
		return fmt.Sprintf("%d", d)
	}

	acc, err := s.accounts.GetAccountOrOwnerAccountIfPresent(ctx, _acc.GetUuid())
	if err != nil {
		log.Error("Error getting instance owner when getting subaccount", zap.Error(err))
		return fmt.Errorf("error getting instance owner: %w", err)
	}

	if acc.Currency == nil {
		acc.Currency = defCurr
	}
	rate, _, err := s.currencies.GetExchangeRate(ctx, defCurr, acc.Currency)
	if err != nil {
		log.Error("Error getting exchange rate", zap.Error(err))
		return fmt.Errorf("error getting exchange rate: %w", err)
	}

	tax := acc.GetTaxRate()

	inv := &graph.Invoice{
		Invoice: &pb.Invoice{
			Status:   pb.BillingStatus_UNPAID,
			Items:    []*pb.Item{},
			Type:     pb.ActionType_INSTANCE_RENEWAL,
			Created:  now,
			Account:  acc.GetUuid(),
			Currency: acc.Currency,
			Meta: map[string]*structpb.Value{
				"creator":               structpb.NewStringValue(schema.ROOT_ACCOUNT_KEY),
				graph.InvoiceTaxMetaKey: structpb.NewNumberValue(tax),
			},
		},
	}

	var (
		total           float64
		totalNoDiscount float64
		expirations     = make([]int64, 0)
	)
	const monthSecs = 3600 * 24 * 30
	const daySecs = 3600 * 24
	billingData := graph.BillingData{
		RenewalData: make(map[string]graph.RenewalData),
	}
	for _, d := range data {
		inst := d.Instance
		initCost, _ := s.instances.CalculateInstanceEstimatePrice(inst, false)
		cost, err := s.promocodes.GetDiscountPriceByInstance(inst, false)
		if err != nil {
			log.Error("Error calculating instance estimate price", zap.Error(err))
			continue
		}

		cost *= rate // Convert from NCU to  account's currency
		cost = cost + (cost * tax)
		initCost *= rate
		initCost = initCost + (initCost * tax)
		total += cost
		totalNoDiscount += initCost

		expireDate := time.Unix(d.ExpireAt, 0)
		var untilDate time.Time
		if d.Period == monthSecs {
			untilDate = expireDate.AddDate(0, 1, 0)
		} else {
			untilDate = expireDate.Add(time.Duration(d.Period) * time.Second)
		}
		if untilDate.Unix()-expireDate.Unix() > daySecs {
			untilDate = untilDate.AddDate(0, 0, -1)
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

		billingData.RenewalData[inst.GetUuid()] = graph.RenewalData{
			ExpirationTs: d.ExpireAt,
		}
		item := &pb.Item{
			Description: renewDescription,
			Amount:      1,
			Unit:        "Pcs",
			Price:       cost,
			Instance:    d.Instance.GetUuid(),
		}
		inv.Items = append(inv.Items, item)
		expirations = append(expirations, d.ExpireAt)
	}
	inv.SetBillingData(&billingData)

	if len(inv.Items) == 0 {
		return errNothingToRenew
	}

	slices.Sort(expirations)
	var dueDate = expirations[len(expirations)-1]
	if dueDate < now {
		dueDate = now + monthSecs
	}
	inv.Deadline = dueDate

	inv.Total = total
	inv.Meta["no_discount_price"] = structpb.NewStringValue(fmt.Sprintf("%.2f %s", totalNoDiscount, acc.Currency.GetTitle()))

	if val, ok := ctx.Value("create_as_draft").(bool); ok && val {
		inv.Status = pb.BillingStatus_DRAFT
	}

	if inv.Total < 0 {
		log.Warn("Total less than 0, skipping invoice creation. Wtf?")
		return errNothingToRenew
	}
	resp, err := s.CreateInvoice(ctx, connect.NewRequest(&pb.CreateInvoiceRequest{
		IsSendEmail: true,
		Invoice:     inv.Invoice,
	}))
	if err != nil {
		log.Error("Error creating invoice", zap.Error(err))
		return fmt.Errorf("error creating invoice: %w", err)
	}

	log.Info("Created invoice", zap.String("uuid", resp.Msg.GetUuid()), zap.Int("item_count", len(inv.Items)))
	return nil
}

func filterInvoices(invoices []*graph.Invoice) []*graph.Invoice {
	var filteredInvoices []*graph.Invoice
	for _, inv := range invoices {
		if inv.GetStatus() == pb.BillingStatus_PAID || inv.GetStatus() == pb.BillingStatus_CANCELED ||
			inv.GetStatus() == pb.BillingStatus_TERMINATED || inv.GetStatus() == pb.BillingStatus_RETURNED {
			continue
		}
		filteredInvoices = append(filteredInvoices, inv)
	}
	return filteredInvoices
}

func filterInstances(instances []*ipb.Instance) []*ipb.Instance {
	var filteredInstances []*ipb.Instance
	for _, inst := range instances {
		if inst.BillingPlan == nil {
			continue
		}
		if inst.GetStatus() == statuses.NoCloudStatus_DEL || inst.GetStatus() == statuses.NoCloudStatus_INIT || inst.GetStatus() == statuses.NoCloudStatus_UNSPECIFIED {
			continue
		}
		if inst.GetState().GetState() == states.NoCloudState_DELETED || inst.GetState().GetState() == states.NoCloudState_PENDING ||
			inst.GetState().GetState() == states.NoCloudState_UNKNOWN || inst.GetState().GetState() == states.NoCloudState_INIT {
			continue
		}
		filteredInstances = append(filteredInstances, inst)
	}
	return filteredInstances
}
