package graph

import (
	"context"

	bpb "github.com/slntopp/nocloud-proto/billing"
	pb "github.com/slntopp/nocloud-proto/instances"
	"google.golang.org/protobuf/proto"
)

// LockProductKey is the billing plan meta flag that leaves the product and plan of its instances to
// platform admins, for plans whose products grant something the moment a period is paid.
const LockProductKey = "lock_product"

// AdminConfigKeys are instance config keys only platform admins set, like admin-ui's "start without
// payment" (skip_next_payment).
var AdminConfigKeys = []string{"skip_next_payment"}

// ChangesProductOrPlan reports whether updating old with inst moves the instance to another product
// or billing plan. A product or plan left out of the update keeps its stored value.
func ChangesProductOrPlan(inst, old *pb.Instance) bool {
	if inst.Product != nil && inst.GetProduct() != old.GetProduct() {
		return true
	}
	return inst.GetBillingPlan() != nil && inst.GetBillingPlan().GetUuid() != old.GetBillingPlan().GetUuid()
}

// LockedProductChange reports whether updating old with inst changes the product or plan of an
// instance whose stored or requested plan has LockProductKey set. The flag is read from the plans
// collection, never from the plan copy in the request or on the instance.
func LockedProductChange(ctx context.Context, plans BillingPlansController, inst, old *pb.Instance) (bool, error) {
	if !ChangesProductOrPlan(inst, old) {
		return false, nil
	}
	for _, uuid := range []string{old.GetBillingPlan().GetUuid(), inst.GetBillingPlan().GetUuid()} {
		if uuid == "" {
			continue
		}
		plan, err := plans.Get(ctx, &bpb.Plan{Uuid: uuid})
		if err != nil {
			return false, err
		}
		if plan.GetMeta()[LockProductKey].GetBoolValue() {
			return true, nil
		}
	}
	return false, nil
}

// KeepAdminConfig gives the AdminConfigKeys in a non-admin's instance config their stored values,
// removing those stored has not, and reports whether the request tried to change one. stored is nil
// for a new instance. An update merges config, so a key left out of the request keeps its value.
func KeepAdminConfig(inst, stored *pb.Instance) bool {
	changed := false
	for _, key := range AdminConfigKeys {
		value, ok := inst.GetConfig()[key]
		if !ok {
			continue
		}
		old, had := stored.GetConfig()[key]
		if had && proto.Equal(value, old) {
			continue
		}
		changed = true
		if had {
			inst.Config[key] = old
		} else {
			delete(inst.Config, key)
		}
	}
	return changed
}
