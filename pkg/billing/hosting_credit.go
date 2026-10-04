package billing

import (
	"strconv"
	"strings"

	ipb "github.com/slntopp/nocloud-proto/instances"
	"google.golang.org/protobuf/types/known/structpb"
)

const siteCreditPerSite = 5.0

const siteCreditConfigKey = "site_credit"

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

func siteCreditAmount(sites int) float64 {
	if sites <= 0 {
		return 0
	}
	return siteCreditPerSite * float64(sites)
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
