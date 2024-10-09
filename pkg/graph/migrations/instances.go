package migrations

import (
	"context"
	"fmt"
	addonspb "github.com/slntopp/nocloud-proto/billing/addons"
	"github.com/slntopp/nocloud-proto/instances"
	servicespb "github.com/slntopp/nocloud-proto/services"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"math"
	"slices"
)

func MigrateInstancesToNewAddons(log *zap.Logger, instCtrl graph.InstancesController, srvCtrl graph.ServicesController, bpCtrl graph.BillingPlansController, addCtrl graph.AddonsController) {
	log = log.Named("MigrateInstancesToNewAddons")
	log.Debug("Starting MigrateInstancesToNewAddons")
	page := uint64(1)
	depth := int32(math.MaxInt32)
	limit := uint64(math.MaxUint64)
	services, err := srvCtrl.List(context.Background(), schema.ROOT_ACCOUNT_KEY, &servicespb.ListRequest{
		Page:  &page,
		Limit: &limit,
		Depth: &depth,
	})
	if err != nil {
		log.Fatal("Failed to list services", zap.Error(err))
		return
	}
	instances := make([]string, 0)
	for _, srv := range services.Result {
		for _, ig := range srv.GetInstancesGroups() {
			for _, inst := range ig.GetInstances() {
				instances = append(instances, inst.GetUuid())
			}
		}
	}
	addons, err := addCtrl.List(context.Background(), &addonspb.ListAddonsRequest{})
	if err != nil {
		log.Fatal("Failed to list addons", zap.Error(err))
		return
	}
	for _, id := range instances {
		log := log.With(zap.String("instance_id", id))
		inst, err := instCtrl.Get(context.Background(), id)
		if err != nil {
			log.Fatal("Failed to get instance", zap.Error(err))
			return
		}
		if err := migrateInstance(log, inst, instCtrl, bpCtrl, addons); err != nil {
			log.Fatal("Failed to migrate instance", zap.Error(err))
			return
		}
	}
	log.Debug("Finished MigrateInstancesToNewAddons")
}

func migrateInstance(log *zap.Logger, inst *graph.Instance, instCtrl graph.InstancesController, bpCtrl graph.BillingPlansController, addons []*addonspb.Addon) error {
	oldInst := proto.Clone(inst.Instance).(*instances.Instance)

	bp, err := bpCtrl.Get(context.Background(), inst.GetBillingPlan())
	if err != nil {
		return err
	}

	product, ok := bp.GetProducts()[inst.GetProduct()]
	if !ok {
		return fmt.Errorf("product '%s' not found in billing plan %s", inst.GetProduct(), inst.GetBillingPlan().GetUuid())
	}

	instAddons := inst.GetAddons()
	instData := inst.GetData()
	oldAddonKeys := inst.GetConfig()["addons"].GetListValue().GetValues()                          // Old addons from config (ovh, virtual, keyweb...)
	oldAddonKeys = append(oldAddonKeys, product.GetMeta()["addons"].GetListValue().GetValues()...) // Old addons from product meta (virtual driver)
	_ = inst.GetConfig()["duration"].GetStringValue()

	for _, oldAddonKey := range oldAddonKeys {
		var a *addonspb.Addon = nil
		oldKey := oldAddonKey.GetStringValue()
		for _, addon := range addons {
			key := addon.GetMeta()["key"].GetStringValue()
			if key == oldKey {
				a = addon
				break
			}
		}
		if a == nil {
			return fmt.Errorf("addon with key '%s' not found in new addons", oldKey)
		}
		if slices.Contains(instAddons, a.GetUuid()) {
			continue
		}
		instAddons = append(instAddons, a.GetUuid())
		lmVal, ok := instData[oldKey+"_last_monitoring"]
		if !ok {
			continue
		}
		lm := int64(lmVal.GetNumberValue())
		instData["addon_"+a.GetUuid()+"_last_monitoring"] = structpb.NewNumberValue(float64(lm))
	}

	inst.Addons = instAddons
	inst.Data = instData

	log.Debug("Updating instance")
	if err := instCtrl.Update(context.Background(), "", inst.Instance, oldInst); err != nil {
		return err
	}

	return nil
}
