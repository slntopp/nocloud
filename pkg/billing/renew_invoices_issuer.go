package billing

import (
	"context"
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
	const days15 = int64(3600 * 24 * 15)
	const days10 = int64(3600 * 24 * 10)
	isExpiring := func(now, expiringAt, period int64) bool {
		if period > days15 {
			return (expiringAt - days10) < now
		}
		return (expiringAt - int64(float64(period)*(1.0-expiringPercentage))) < now
	}

	accountPools := map[string]*accountPool{}
	for _, srv := range list.Result {
		for _, ig := range srv.GetInstancesGroups() {
			for _, inst := range ig.GetInstances() {
				log := log.With(zap.String("instance", inst.GetUuid()))
				acc, err := s.instances.GetInstanceOwner(ctx, inst.GetUuid())
				if err != nil {
					log.Error("Error getting instance owner", zap.Error(err))
					continue
				}
				instId := driver.NewDocumentID(schema.INSTANCES_COL, inst.GetUuid()).String()
				spResp, err := s.instances.GetGroup(ctx, instId)
				if err != nil {
					log.Error("Error getting instance group and sp", zap.Error(err))
					continue
				}
				if spResp.SP == nil || spResp.Group == nil {
					log.Error("Error getting instance group and sp. Nil values", zap.Error(err))
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
					continue
				}
				expResp, err := client.GetExpiration(ctx, &dpb.GetExpirationRequest{Instance: inst, ServicesProvider: spResp.SP})
				if err != nil {
					log.Error("Error getting instance expirations", zap.Error(err))
					continue
				}
				accountPools[acc.GetUuid()].InstExpRecords[inst.GetUuid()] = expResp.GetRecords()
			}
		}
	}

	invoices, err := s.invoices.List(ctx, "")
	if err != nil {
		log.Error("Error listing invoices", zap.Error(err))
		return
	}
	for _, inv := range invoices {
		if inv.GetAccount() == "" {
			log.Warn("Not linked invoice found", zap.String("uuid", inv.GetUuid()))
			continue
		}
		if accountPools[inv.GetAccount()] == nil {
			acc, err := s.accounts.Get(ctx, inv.GetAccount())
			if err != nil {
				log.Error("Error getting account for invoice list", zap.Error(err))
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

	for _, pool := range accountPools {
		s.processAccountRenewalInvoices(ctx, log, pool, isExpiring)
	}

	log.Info("Finished Invoice Expiring Instances Cron Job")
}

type instanceExpData struct {
	Instance *ipb.Instance
	ExpireAt int64
	Period   int64
}

func (s *BillingServiceServer) processAccountRenewalInvoices(ctx context.Context, log *zap.Logger, data *accountPool, isExp func(now, expiringAt, period int64) bool) {
	log = log.Named("ProcessAccount").With(zap.String("account", data.Account.GetUuid()))
	log.Info("Processing account", zap.Int("instance_count", len(data.Instances)), zap.Int("invoice_count", len(data.Invoices)), zap.Int("expirings", len(data.InstExpRecords)))
	defer func() {
		if err := recover(); err != nil {
			log.Error("Recovered from panic", zap.Error(fmt.Errorf("%v", err)))
		}
	}()

	expData := make([]*instanceExpData, 0)
	instances := filterInstances(data.Instances)
	for _, inst := range instances {
		product, ok := inst.GetBillingPlan().GetProducts()[inst.GetProduct()]
		if !ok {
			log.Warn("Product not found for instance", zap.String("product", inst.GetProduct()), zap.String("instance", inst.GetUuid()))
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
			continue
		}
		if product.GetPeriod() != period {
			log.Error("Product period mismatch", zap.String("product", inst.GetProduct()), zap.String("instance", inst.GetUuid()))
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
