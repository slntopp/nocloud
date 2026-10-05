package billing

import (
	"strconv"
	"strings"

	bpb "github.com/slntopp/nocloud-proto/billing"
	ipb "github.com/slntopp/nocloud-proto/instances"
	"google.golang.org/protobuf/types/known/structpb"
)

const defaultSiteCreditPerSite = 5.0

const siteCreditConfigKey = "site_credit"

const siteCreditMetaKey = "site_credit"

func hostingSites(inst *ipb.Instance) int {
	if inst == nil {
		return 0
	}
	if n := resourceSites(inst.GetResources()); n > 0 {
		return n
	}
	plan := inst.GetBillingPlan()
	if plan == nil || plan.GetProducts() == nil {
		return 0
	}
	product := plan.GetProducts()[inst.GetProduct()]
	if product == nil {
		return 0
	}
	return resourceSites(product.GetResources())
}

func siteCreditAmount(inst *ipb.Instance, sites int) float64 {
	if sites <= 0 {
		return 0
	}
	per := defaultSiteCreditPerSite
	if n, ok := productMetaNumber(inst, siteCreditMetaKey); ok {
		if n < 0 {
			n = 0
		}
		per = n
	}
	return per * float64(sites)
}

func hostingProduct(inst *ipb.Instance) *bpb.Product {
	if inst == nil || inst.GetBillingPlan() == nil {
		return nil
	}
	return inst.GetBillingPlan().GetProducts()[inst.GetProduct()]
}

func productMetaNumber(inst *ipb.Instance, key string) (float64, bool) {
	product := hostingProduct(inst)
	if product == nil || product.GetMeta() == nil {
		return 0, false
	}
	value := product.GetMeta()[key]
	if value == nil {
		return 0, false
	}
	switch kind := value.GetKind().(type) {
	case *structpb.Value_NumberValue:
		return kind.NumberValue, true
	case *structpb.Value_StringValue:
		n, err := strconv.ParseFloat(strings.TrimSpace(kind.StringValue), 64)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}

func siteCreditGranted(config map[string]*structpb.Value) bool {
	if config == nil {
		return false
	}
	value := config[siteCreditConfigKey]
	return value != nil && value.GetBoolValue()
}

func resourceSites(resources map[string]*structpb.Value) int {
	if resources == nil {
		return 0
	}
	for _, key := range []string{"sites", "websites"} {
		if n := resourceCount(resources[key]); n > 0 {
			return n
		}
	}
	return 0
}

func resourceCount(value *structpb.Value) int {
	if value == nil {
		return 0
	}
	switch kind := value.GetKind().(type) {
	case *structpb.Value_NumberValue:
		if kind.NumberValue > 0 {
			return int(kind.NumberValue)
		}
	case *structpb.Value_StringValue:
		text := strings.TrimSpace(kind.StringValue)
		if n, err := strconv.Atoi(text); err == nil && n > 0 {
			return n
		}
		if n, err := strconv.ParseFloat(text, 64); err == nil && n > 0 {
			return int(n)
		}
	}
	return 0
}
