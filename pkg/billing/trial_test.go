package billing

import (
	"fmt"
	"testing"

	"github.com/arangodb/go-driver"
	bpb "github.com/slntopp/nocloud-proto/billing"
	ipb "github.com/slntopp/nocloud-proto/instances"
	"google.golang.org/protobuf/types/known/structpb"
)

func trialProduct(t *testing.T, trial map[string]any) *bpb.Product {
	t.Helper()
	if trial == nil {
		return &bpb.Product{}
	}
	s, err := structpb.NewStruct(trial)
	if err != nil {
		t.Fatal(err)
	}
	return &bpb.Product{Meta: map[string]*structpb.Value{"trial": structpb.NewStructValue(s)}}
}

func TestProductTrial(t *testing.T) {
	plan := &bpb.Plan{Type: "scanning"}

	if conf, err := productTrial(plan, trialProduct(t, nil)); conf != nil || err != nil {
		t.Fatalf("product without trial: got %+v, %v", conf, err)
	}

	conf, err := productTrial(plan, trialProduct(t, map[string]any{"period": 604800}))
	if err != nil {
		t.Fatal(err)
	}
	if conf.Period != 604800 || conf.Scope != "account" || conf.Group != "scanning" {
		t.Fatalf("defaults: got %+v, want account scope grouped by plan type", conf)
	}

	conf, err = productTrial(plan, trialProduct(t, map[string]any{"period": 60, "scope": "config:domain", "group": "easm"}))
	if err != nil || conf.Scope != "config:domain" || conf.Group != "easm" {
		t.Fatalf("explicit scope and group: got %+v, %v", conf, err)
	}

	for name, bad := range map[string]map[string]any{
		"no period":       {"scope": "account"},
		"negative period": {"period": -1},
		"unknown scope":   {"period": 60, "scope": "namespace"},
		"empty key":       {"period": 60, "scope": "config:"},
	} {
		if conf, err := productTrial(plan, trialProduct(t, bad)); err == nil {
			t.Errorf("%s: accepted as %+v", name, conf)
		}
	}
}

func TestTrialScopeValue(t *testing.T) {
	inst := &ipb.Instance{Config: map[string]*structpb.Value{
		"domain": structpb.NewStringValue("  Example.COM "),
		"blank":  structpb.NewStringValue(" "),
		"list":   structpb.NewListValue(&structpb.ListValue{}),
	}}

	if v, ok := (&trialConf{Scope: "account"}).scopeValue("acc-1", inst); !ok || v != "acc-1" {
		t.Errorf("account scope: got %q, %v", v, ok)
	}
	if _, ok := (&trialConf{Scope: "account"}).scopeValue("", inst); ok {
		t.Error("account scope without an account must not yield a trial")
	}
	// The same domain must map to the same trial however it is typed.
	if v, ok := (&trialConf{Scope: "config:domain"}).scopeValue("acc-1", inst); !ok || v != "example.com" {
		t.Errorf("config scope: got %q, %v, want normalised domain", v, ok)
	}
	for _, key := range []string{"missing", "blank", "list"} {
		if v, ok := (&trialConf{Scope: "config:" + key}).scopeValue("acc-1", inst); ok {
			t.Errorf("config:%s yielded trial value %q; an instance with nothing to scope to gets no trial", key, v)
		}
	}
}

func TestTrialInvoiceDate(t *testing.T) {
	plan := &bpb.Plan{Type: "scanning"}
	const now, day = int64(1_800_000_000), int64(86400)

	// A week's trial is invoiced two days before it ends by default.
	conf, err := productTrial(plan, trialProduct(t, map[string]any{"period": 7 * day}))
	if err != nil {
		t.Fatal(err)
	}
	if got := conf.invoiceDate(now); got != now+5*day {
		t.Errorf("default lead: invoice at now+%ds, want now+5d", got-now)
	}

	conf, err = productTrial(plan, trialProduct(t, map[string]any{"period": 7 * day, "invoice_before": day}))
	if err != nil || conf.invoiceDate(now) != now+6*day {
		t.Errorf("explicit lead: got %+v, %v", conf, err)
	}

	// Zero means at the very end of the trial.
	conf, err = productTrial(plan, trialProduct(t, map[string]any{"period": 7 * day, "invoice_before": 0}))
	if err != nil || conf.invoiceDate(now) != now+7*day {
		t.Errorf("zero lead: got %+v, %v", conf, err)
	}

	// A trial shorter than the lead is invoiced immediately, never in the past.
	conf, err = productTrial(plan, trialProduct(t, map[string]any{"period": day}))
	if err != nil || conf.invoiceDate(now) != now {
		t.Errorf("short trial: got %+v, %v", conf, err)
	}

	if conf, err := productTrial(plan, trialProduct(t, map[string]any{"period": day, "invoice_before": -1})); err == nil {
		t.Errorf("negative lead accepted as %+v", conf)
	}
}

func TestIsWriteConflict(t *testing.T) {
	if isWriteConflict(nil) {
		t.Error("nil is not a conflict")
	}
	if !isWriteConflict(fmt.Errorf("AQL: write-write conflict [node #4: UpdateNode] (while executing)")) {
		t.Error("wrapped AQL conflict not recognised")
	}
	if !isWriteConflict(driver.ArangoError{HasError: true, Code: 409, ErrorNum: driver.ErrArangoConflict}) {
		t.Error("ArangoDB conflict error not recognised")
	}
	if isWriteConflict(fmt.Errorf("document not found")) {
		t.Error("unrelated error taken for a conflict")
	}
}
