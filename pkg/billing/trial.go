package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
	bpb "github.com/slntopp/nocloud-proto/billing"
	ipb "github.com/slntopp/nocloud-proto/instances"
	"github.com/slntopp/nocloud/pkg/graph"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// A product offers a free trial through its meta:
//
//	meta:
//	  trial:
//	    period: 604800      # seconds
//	    scope: account      # or "config:<key>", e.g. "config:domain"
//	    group: scanning     # optional, defaults to the billing plan type
//	    invoice_before: 172800  # optional, seconds; defaults to two days
//
// One trial exists per scope value within a group: per account, or per value
// of an instance config key (so a trial can be tied to a domain rather than
// to whoever orders it). An instance that gets the trial starts without a
// start invoice and carries config.trial_ends; its driver bills from there.
//
// The first renewal invoice is issued invoice_before seconds ahead of the end
// of the trial rather than by the usual rule, which for a monthly product is
// ten days ahead of expiry and so would bill a week's trial on its first day.

// TrialEndsKey is the instance config key holding the Unix time a granted
// trial ends. It is one of graph.AdminConfigKeys: customers cannot set it.
const TrialEndsKey = "trial_ends"

// defaultTrialInvoiceBefore is how long before the end of a trial its first
// renewal invoice is issued unless the product says otherwise.
const defaultTrialInvoiceBefore = int64(2 * 24 * 60 * 60)

const (
	trialStartAttempts = 5
	trialStartBackoff  = 200 * time.Millisecond
)

// isWriteConflict reports whether ArangoDB refused a write because another
// transaction was changing the same document.
func isWriteConflict(err error) bool {
	return err != nil && (driver.IsArangoErrorWithErrorNum(err, driver.ErrArangoConflict) || strings.Contains(err.Error(), "write-write conflict"))
}

const (
	trialScopeAccount      = "account"
	trialScopeConfigPrefix = "config:"
)

type trialConf struct {
	Period int64
	Scope  string
	Group  string
	// InvoiceBefore is how long before the end of the trial the first
	// renewal invoice is issued.
	InvoiceBefore int64
}

// invoiceDate is when the first renewal invoice of a trial started at now is
// due. A trial shorter than the lead time is invoiced straight away.
func (c *trialConf) invoiceDate(now int64) int64 {
	if date := now + c.Period - c.InvoiceBefore; date > now {
		return date
	}
	return now
}

// productTrial reads a product's trial configuration; nil means it has none.
func productTrial(plan *bpb.Plan, product *bpb.Product) (*trialConf, error) {
	raw := product.GetMeta()["trial"].GetStructValue().GetFields()
	if raw == nil {
		return nil, nil
	}
	conf := &trialConf{
		Period: int64(raw["period"].GetNumberValue()),
		Scope:  raw["scope"].GetStringValue(),
		Group:  raw["group"].GetStringValue(),
	}
	if conf.Period <= 0 {
		return nil, fmt.Errorf("trial period must be positive")
	}
	if conf.Scope == "" {
		conf.Scope = trialScopeAccount
	}
	if conf.Scope != trialScopeAccount &&
		(!strings.HasPrefix(conf.Scope, trialScopeConfigPrefix) || conf.Scope == trialScopeConfigPrefix) {
		return nil, fmt.Errorf("trial scope %q is neither %q nor %q", conf.Scope, trialScopeAccount, trialScopeConfigPrefix+"<key>")
	}
	if conf.Group == "" {
		conf.Group = plan.GetType()
	}
	conf.InvoiceBefore = defaultTrialInvoiceBefore
	if v, ok := raw["invoice_before"]; ok {
		conf.InvoiceBefore = int64(v.GetNumberValue())
		if conf.InvoiceBefore < 0 {
			return nil, fmt.Errorf("trial invoice_before must not be negative")
		}
	}
	return conf, nil
}

// scopeValue is what this instance's trial would be recorded against. ok is
// false when a config-scoped trial's key is missing or empty on the instance:
// with nothing to tie the trial to, none is given.
func (c *trialConf) scopeValue(account string, inst *ipb.Instance) (value string, ok bool) {
	if c.Scope == trialScopeAccount {
		return account, account != ""
	}
	v := inst.GetConfig()[strings.TrimPrefix(c.Scope, trialScopeConfigPrefix)]
	switch v.GetKind().(type) {
	case *structpb.Value_StringValue:
		value = strings.ToLower(strings.TrimSpace(v.GetStringValue()))
	case *structpb.Value_NumberValue, *structpb.Value_BoolValue:
		value = fmt.Sprint(v.AsInterface())
	}
	return value, value != ""
}

// tryStartTrial starts the instance on its product's free trial if it has one
// that is still unused, and reports whether it did. An instance started this
// way gets no start invoice.
func (s *BillingServiceServer) tryStartTrial(ctx context.Context, log *zap.Logger, inst *ipb.Instance, account string, now int64) (bool, error) {
	conf, err := productTrial(inst.GetBillingPlan(), inst.GetBillingPlan().GetProducts()[inst.GetProduct()])
	if err != nil {
		// A misconfigured trial must not block orders: bill normally.
		log.Warn("Ignoring invalid product trial", zap.Error(err))
		return false, nil
	}
	if conf == nil {
		return false, nil
	}
	value, ok := conf.scopeValue(account, inst)
	if !ok {
		log.Info("Instance has nothing to scope a trial to", zap.String("scope", conf.Scope))
		return false, nil
	}

	ends := now + conf.Period
	claimed, err := s.trials.Claim(ctx, graph.Trial{
		Group: conf.Group, Scope: conf.Scope, Value: value,
		Account: account, Instance: inst.GetUuid(), Started: now, Ends: ends,
	})
	if err != nil {
		return false, fmt.Errorf("failed to claim trial: %w", err)
	}
	if !claimed {
		log.Info("Trial already used", zap.String("group", conf.Group), zap.String("scope", conf.Scope))
		return false, nil
	}

	started := proto.Clone(inst).(*ipb.Instance)
	if started.Config == nil {
		started.Config = map[string]*structpb.Value{}
	}
	started.Config["auto_start"] = structpb.NewBoolValue(true)
	started.Config[TrialEndsKey] = structpb.NewNumberValue(float64(ends))
	// The renewal job clears this once it has issued the invoice, so later
	// renewals follow the usual rule.
	invoiceAt := conf.invoiceDate(now)
	if started.Meta == nil {
		started.Meta = &ipb.InstanceMeta{}
	}
	started.Meta.NextForcedRenewInvoice = &invoiceAt
	// The order is still being brought up by other services when this runs,
	// and their writes to the instance can collide with this one. Left to
	// the queue, a collision would delay the trial by a redelivery (a minute).
	err = nil
	for attempt := 1; attempt <= trialStartAttempts; attempt++ {
		// Update strips the UUID from what it is given: hand it a fresh copy.
		if err = s.instances.Update(ctx, "", proto.Clone(started).(*ipb.Instance), inst); err == nil || !isWriteConflict(err) {
			break
		}
		log.Warn("Write conflict starting instance on trial, retrying", zap.Int("attempt", attempt), zap.Error(err))
		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(attempt) * trialStartBackoff):
		}
	}
	if err != nil {
		// Without this the trial would be spent on an instance that never ran.
		if relErr := s.trials.Release(ctx, conf.Group, conf.Scope, value); relErr != nil {
			log.Error("Failed to release trial after failed start", zap.Error(relErr))
		}
		return false, fmt.Errorf("failed to start instance on trial: %w", err)
	}
	log.Info("Started instance on trial", zap.Int64("ends", ends), zap.Int64("invoice_at", invoiceAt), zap.String("group", conf.Group), zap.String("scope", conf.Scope))
	return true, nil
}
