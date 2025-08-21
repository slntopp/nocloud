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
package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/slntopp/nocloud-proto/access"
	addonspb "github.com/slntopp/nocloud-proto/billing/addons"
	dpb "github.com/slntopp/nocloud-proto/billing/descriptions"
	epb "github.com/slntopp/nocloud-proto/events"
	instancespb "github.com/slntopp/nocloud-proto/instances"
	"github.com/slntopp/nocloud-proto/notes"
	servicespb "github.com/slntopp/nocloud-proto/services"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	ps "github.com/slntopp/nocloud/pkg/pubsub"
	"github.com/slntopp/nocloud/pkg/pubsub/services_registry"
	"google.golang.org/protobuf/types/known/structpb"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/wI2L/jsondiff"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	bpb "github.com/slntopp/nocloud-proto/billing"
	elpb "github.com/slntopp/nocloud-proto/events_logging"
	"github.com/slntopp/nocloud-proto/hasher"
	pb "github.com/slntopp/nocloud-proto/instances"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	stpb "github.com/slntopp/nocloud-proto/states"
	spb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

type InstancesController interface {
	CalculateInstanceEstimatePrice(i *pb.Instance, includeOneTimePayments bool) (float64, error)
	GetInstancePeriod(i *pb.Instance) (*int64, error)
	Create(ctx context.Context, group driver.DocumentID, sp string, i *pb.Instance) (string, error)
	Update(ctx context.Context, sp string, inst, oldInst *pb.Instance) error
	UpdateWithPatch(ctx context.Context, sp string, inst, oldInst *pb.Instance) error
	Exists(ctx context.Context, uuid string) (bool, error)
	UpdateNotes(ctx context.Context, inst *pb.Instance) error
	Delete(ctx context.Context, group string, i *pb.Instance) error
	Get(ctx context.Context, uuid string) (*Instance, error)
	GetWithAccess(ctx context.Context, from driver.DocumentID, id string) (Instance, error)
	GetGroup(ctx context.Context, i string) (*GroupWithSP, error)
	CheckEdgeExist(ctx context.Context, spUuid string, i *pb.Instance) error
	ValidateBillingPlan(ctx context.Context, spUuid string, i *pb.Instance) error
	SetStatus(ctx context.Context, inst *pb.Instance, status spb.NoCloudStatus) (err error)
	SetState(ctx context.Context, inst *pb.Instance, state stpb.NoCloudState) (err error)
	TransferInst(ctx context.Context, oldIGEdge string, newIG driver.DocumentID, inst driver.DocumentID) error
	GetEdge(ctx context.Context, inboundNode string, collection string) (string, error)
	GetInstanceOwner(ctx context.Context, uuid string) (Account, error)
	getSp(ctx context.Context, uuid string) (string, error)
}

const (
	INSTANCES_COL                   = "Instances"
	CreationPromocodeKey ContextKey = "creation-promocode"
)

type Instance struct {
	*pb.Instance
	driver.DocumentMeta
}

type instancesController struct {
	col   driver.Collection // Instances Collection
	graph driver.Graph

	log *zap.Logger

	db driver.Database

	ig2inst driver.Collection

	inv     InvoicesController
	acc     AccountsController
	cur     CurrencyController
	addons  AddonsController
	channel rabbitmq.Channel

	bp_ctrl BillingPlansController

	ps    *ps.PubSub[*epb.Event]
	ansPs *ps.PubSub[*pb.Context]
}

// Migrations

func MigrateInstancesToNewAddons(log *zap.Logger, instCtrl InstancesController, srvCtrl ServicesController, bpCtrl BillingPlansController, addCtrl AddonsController, descCtrl DescriptionsController) {
	log = log.Named("MigrateInstancesToNewAddons")
	log.Debug("Starting MigrateInstancesToNewAddons")
	page := uint64(1)
	depth := int32(150)
	limit := uint64(0)
	services, err := srvCtrl.List(context.Background(), schema.ROOT_ACCOUNT_KEY, &servicespb.ListRequest{
		Page:  &page,
		Limit: &limit,
		Depth: &depth,
	})
	if err != nil {
		log.Fatal("Failed to list services", zap.Error(err))
		return
	}
	log.Debug("Found " + fmt.Sprint(len(services.Result)) + " services")
	instances := make([]string, 0)
	for _, srv := range services.Result {
		for _, ig := range srv.GetInstancesGroups() {
			for _, inst := range ig.GetInstances() {
				instances = append(instances, inst.GetUuid())
			}
		}
	}
	log.Debug("Creating virtual addons")
	bps, err := bpCtrl.List(context.Background(), "")
	if err != nil {
		log.Fatal("Failed to list billing plans", zap.Error(err))
		return
	}
	if err := createVirtualAddons(log.Named("CreateVirtualAddons"), bps, addCtrl, descCtrl, bpCtrl); err != nil {
		log.Fatal("Failed to create virtual addons", zap.Error(err))
		return
	}
	log.Debug("Found " + fmt.Sprint(len(instances)) + " instances")
	addons, err := addCtrl.List(context.Background(), &addonspb.ListAddonsRequest{})
	if err != nil {
		log.Fatal("Failed to list addons", zap.Error(err))
		return
	}
	log.Debug("Found " + fmt.Sprint(len(addons)) + " addons")
	var migrated = make([]string, 0)
	for _, id := range instances {
		log := log.With(zap.String("instance_id", id))
		inst, err := instCtrl.Get(context.Background(), id)
		if err != nil {
			log.Fatal("Failed to get instance", zap.Error(err))
			return
		}
		if err := migrateInstance(log, inst, instCtrl, bpCtrl, addons, &migrated); err != nil {
			log.Fatal("Failed to migrate instance", zap.Error(err))
			return
		}
	}
	log.Debug("Migrated instances", zap.Any("uuids", migrated))
	log.Debug("Finished MigrateInstancesToNewAddons")
}

func createVirtualAddons(log *zap.Logger, _bps []*BillingPlan, addCtrl AddonsController, descCtrl DescriptionsController, bpCtrl BillingPlansController) error {
	// Collect only 'empty' plans
	bps := make([]*BillingPlan, 0)
	for _, bp := range _bps {
		if bp.Type == "empty" {
			bps = append(bps, bp)
		}
	}
	log.Debug("Found " + fmt.Sprint(len(bps)) + " empty type billing plans")

	// Collect resources which will be used to create new addons, exclude 'ram', 'cpu', 'drive_ssd', 'drive_hdd', 'ips_public', 'ips_private' just in case
	resources := make([]*bpb.ResourceConf, 0)
	resToBp := make([]*BillingPlan, 0)
	for _, bp := range bps {
		for _, res := range bp.Resources {
			if res.GetKey() == "ram" || res.GetKey() == "cpu" || res.GetKey() == "drive_ssd" ||
				res.GetKey() == "drive_hdd" || res.GetKey() == "ips_public" || res.GetKey() == "ips_private" {
				continue
			}
			resources = append(resources, res)
			resToBp = append(resToBp, bp)
		}
	}
	log.Debug("Found " + fmt.Sprint(len(resources)) + " resources which will be used to create new addons")

	addons, err := addCtrl.List(context.Background(), &addonspb.ListAddonsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list addons: %w", err)
	}

	for i, res := range resources {
		isSkip := false
		for _, addon := range addons {
			if addon.Group == resToBp[i].GetUuid() && addon.GetMeta()["key"].GetStringValue() == res.GetKey() {
				log.Debug("Found existing addon", zap.String("key", res.GetKey()), zap.String("group", resToBp[i].GetUuid()))
				isSkip = true
				break
			}
		}
		if isSkip {
			continue
		}

		var descrId string
		if desc, ok := res.GetMeta()["description"]; ok && desc.GetStringValue() != "" {
			newDesc, err := descCtrl.Create(context.Background(), &dpb.Description{
				Text: desc.GetStringValue(),
			})
			if err != nil {
				return fmt.Errorf("failed to create description: %w", err)
			}
			descrId = newDesc.GetUuid()
		}

		var title = res.GetTitle()
		if title == "" {
			title = res.GetKey()
		}

		addon := &addonspb.Addon{
			Title:         title,
			Public:        true,
			Group:         resToBp[i].GetUuid(),
			DescriptionId: descrId,
			Meta: map[string]*structpb.Value{
				"key": structpb.NewStringValue(res.GetKey()),
			},
			Kind:   addonspb.Kind(res.GetKind()),
			System: false,
			Periods: map[int64]float64{
				res.GetPeriod(): res.GetPrice(),
			},
			Created: time.Now().Unix(),
		}
		newAddon, err := addCtrl.Create(context.Background(), addon)
		if err != nil {
			return fmt.Errorf("failed to create addon: %w", err)
		}
		if !slices.Contains(resToBp[i].GetAddons(), newAddon.GetUuid()) {
			resToBp[i].Addons = append(resToBp[i].GetAddons(), newAddon.GetUuid())
			if _, err := bpCtrl.Update(context.Background(), resToBp[i].Plan); err != nil {
				return fmt.Errorf("failed to update billing plan: %w", err)
			}
		}
	}

	return nil
}

func migrateInstance(log *zap.Logger, inst *Instance, instCtrl InstancesController, bpCtrl BillingPlansController, addons []*addonspb.Addon, migrated *[]string) error {
	oldInst := proto.Clone(inst.Instance).(*instancespb.Instance)

	bp, err := bpCtrl.Get(context.Background(), inst.GetBillingPlan())
	if err != nil {
		return fmt.Errorf("failed to get billing plan: %w", err)
	}

	if inst.GetProduct() == "" {
		log.Debug("Skipping instance without product")
		return nil
	}

	product, ok := bp.GetProducts()[inst.GetProduct()]
	if !ok {
		return fmt.Errorf("product '%s' not found in billing plan %s", inst.GetProduct(), inst.GetBillingPlan().GetUuid())
	}

	instAddons := inst.GetAddons()
	instData := inst.GetData()
	oldAddonKeys := inst.GetConfig()["addons"].GetListValue().GetValues() // Old addons from config (ovh, virtual, keyweb...)
	//oldAddonKeys = append(oldAddonKeys, product.GetMeta()["addons"].GetListValue().GetValues()...) // Old addons from product meta (virtual driver)
	_ = inst.GetConfig()["duration"].GetStringValue()

	modified := false
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
		if _, ok := a.GetPeriods()[product.GetPeriod()]; !ok {
			return fmt.Errorf("addon '%s' with key '%s' don't have period that instances's product have. Addon missing '%d' period", a.GetUuid(), oldKey, product.GetPeriod())
		}
		if slices.Contains(instAddons, a.GetUuid()) {
			continue
		}
		instAddons = append(instAddons, a.GetUuid())
		modified = true
		lmVal, ok := instData[oldKey+"_last_monitoring"]
		if !ok {
			continue
		}
		lm := int64(lmVal.GetNumberValue())
		instData["addon_"+a.GetUuid()+"_last_monitoring"] = structpb.NewNumberValue(float64(lm))
	}

	if !modified {
		return nil
	}

	inst.Addons = instAddons
	inst.Data = instData

	log.Debug("Updating instance")
	if err := instCtrl.Update(context.WithValue(context.Background(), nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY), "", inst.Instance, oldInst); err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}
	*migrated = append(*migrated, inst.GetUuid())

	return nil
}

//

func NewInstancesController(log *zap.Logger, db driver.Database, conn rabbitmq.Connection) InstancesController {
	ctx := context.TODO()

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GetEnsureCollection(log, ctx, db, schema.INSTANCES_COL)
	ig2inst := GraphGetEdgeEnsure(log, ctx, graph, schema.IG2INST, schema.INSTANCES_GROUPS_COL, schema.INSTANCES_COL)

	bp_ctrl := NewBillingPlansController(log, db)
	addons := NewAddonsController(log, db)
	acc := NewAccountsController(log, db)
	inv := NewInvoicesController(log, db)
	cur := NewCurrencyController(log, db)

	return &instancesController{log: log.Named("InstancesController"), col: col, graph: graph, db: db, ig2inst: ig2inst, bp_ctrl: bp_ctrl,
		addons: addons, inv: inv, acc: acc, cur: cur, ps: ps.NewPubSub[*epb.Event](conn, log), ansPs: ps.NewPubSub[*pb.Context](conn, log)}
}

// CalculateInstanceEstimatePrice return estimate periodic price for current instance in NCU currency
func (ctrl *instancesController) CalculateInstanceEstimatePrice(i *pb.Instance, includeOneTimePayments bool) (float64, error) {
	if i == nil {
		return -1, fmt.Errorf("instance is nil")
	}

	plan, err := ctrl.bp_ctrl.Get(context.Background(), i.GetBillingPlan())
	if err != nil {
		return -1, err
	}

	cost := 0.0

	product, ok := plan.GetProducts()[i.GetProduct()]
	prodPeriod := product.GetPeriod()
	if ok && product != nil {
		charge := (prodPeriod == 0 && includeOneTimePayments) || prodPeriod != 0
		if charge {
			cost += product.Price
		}

		for _, addonId := range i.GetAddons() {
			addon, err := ctrl.addons.Get(context.Background(), addonId)
			if err != nil {
				return -1, fmt.Errorf("failed to get addon %s: %w", addonId, err)
			}
			price, hasPrice := addon.GetPeriods()[prodPeriod]
			if !hasPrice {
				return -1, fmt.Errorf("addon %s has no price for period %d", addonId, prodPeriod)
			}
			if charge {
				cost += price
			}
		}
	}

	for _, res := range plan.GetResources() {

		if res.GetPeriod() == 0 && !includeOneTimePayments {
			continue
		}

		// ram and drive_size calculates in GB for value
		if res.GetKey() == "ram" {
			value := i.GetResources()["ram"].GetNumberValue() / 1024
			cost += value * res.GetPrice()
		} else if strings.Contains(res.GetKey(), "drive") {
			driveType := i.GetResources()["drive_type"].GetStringValue()
			if res.GetKey() != "drive_"+strings.ToLower(driveType) {
				continue
			}
			count := i.GetResources()["drive_size"].GetNumberValue() / 1024
			cost += res.GetPrice() * count
		} else if res.GetKey() == "who_is_privacy" {
			value := float64(0)
			if i.GetResources()["who_is_privacy"].GetBoolValue() {
				value = 1
			}
			cost += value * res.GetPrice()
		} else {
			count := i.GetResources()[res.GetKey()].GetNumberValue()
			cost += res.GetPrice() * count
		}
	}

	return cost, nil
}

// GetInstancePeriod returns billing period for the whole instance
//
// Now it simply returns product's period
func (ctrl *instancesController) GetInstancePeriod(i *pb.Instance) (*int64, error) {
	zero := int64(0)
	_err := int64(-1)

	if i == nil {
		return &_err, fmt.Errorf("instance is nil")
	}

	plan, err := ctrl.bp_ctrl.Get(context.Background(), i.GetBillingPlan())
	if err != nil {
		return &_err, err
	}

	product, ok := plan.GetProducts()[i.GetProduct()]
	if ok {
		if product.GetPeriod() > 0 {
			period := product.GetPeriod()
			return &period, nil
		} else {
			return &zero, nil
		}
	}

	return nil, nil
}

const instanceOwner = `
LET account = LAST( // Find Instance owner Account
    FOR node, edge, path IN 4
    INBOUND DOCUMENT(@@instances, @instance)
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner","owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@accounts)
        RETURN node
    )
FILTER account
RETURN account`

func (ctrl *instancesController) GetInstanceOwner(ctx context.Context, uuid string) (Account, error) {
	log := ctrl.log.Named("GetInstanceOwner")

	cur, err := ctrl.db.Query(ctx, instanceOwner, map[string]interface{}{
		"instance":    uuid,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"@instances":  schema.INSTANCES_COL,
		"@accounts":   schema.ACCOUNTS_COL,
	})
	if err != nil {
		log.Error("Error getting instance owner. Failed to execute query", zap.Error(err))
		return Account{}, fmt.Errorf("error getting instance owner: %w", err)
	}
	if !cur.HasMore() {
		return Account{}, fmt.Errorf("no instance owner")
	}
	var acc Account
	meta, err := cur.ReadDocument(ctx, &acc)
	if err != nil {
		log.Error("Error getting instance owner. Failed to read from cursor", zap.Error(err))
		return Account{}, fmt.Errorf("failed to get instance owner: %w", err)
	}
	log.Debug("GetInstanceOwner", zap.String("instance", uuid), zap.Any("account", acc), zap.Any("meta", meta))
	if meta.Key == "" {
		log.Error("Instance owner not found. Uuid is empty")
		return Account{}, fmt.Errorf("instance owner not found. Uuid is empty")
	}
	acc.Uuid = meta.Key
	return acc, nil
}

func (ctrl *instancesController) GetWithAccess(ctx context.Context, from driver.DocumentID, id string) (Instance, error) {
	var o Instance
	vars := map[string]interface{}{
		"account":     from,
		"node":        driver.NewDocumentID(schema.INSTANCES_COL, id),
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"bps":         schema.BILLING_PLANS_COL,
	}
	c, err := ctrl.db.Query(ctx, getInstanceWithAccessLevel, vars)
	if err != nil {
		return o, err
	}
	defer c.Close()

	meta, err := c.ReadDocument(ctx, &o)
	if err != nil {
		return o, err
	}

	if from.String() == meta.ID.String() {
		o.GetAccess().Level = access.Level_ROOT
	}

	return o, nil
}

func (ctrl *instancesController) Exists(ctx context.Context, uuid string) (bool, error) {
	return ctrl.col.DocumentExists(ctx, uuid)
}

func (ctrl *instancesController) Create(ctx context.Context, group driver.DocumentID, sp string, i *pb.Instance) (string, error) {
	log := ctrl.log.Named("Create")
	log.Debug("Creating Instance", zap.Any("instance", i))

	// ensure status is INIT
	i.Uuid = ""
	i.Status = spb.NoCloudStatus_INIT
	i.Created = time.Now().Unix()

	// Set estimate and period values
	estimate, err := ctrl.CalculateInstanceEstimatePrice(i, false)
	if err != nil {
		log.Error("Failed to calculate estimate instance periodic price", zap.Error(err))
		return "", err
	}
	i.Estimate = estimate
	period, err := ctrl.GetInstancePeriod(i)
	if err != nil {
		log.Error("Failed to get instance period", zap.Error(err))
		return "", err
	}
	i.Period = period

	err = hasher.SetHash(i.ProtoReflect())
	if err != nil {
		log.Error("Failed to calculate hash", zap.Error(err))
		return "", err
	}

	log.Debug("instance for hash calculating while Creating", zap.Any("inst", i))

	if i.GetConfig()["auto_start"].GetBoolValue() {
		if i.Meta == nil {
			i.Meta = &pb.InstanceMeta{}
		}
		i.Meta.Started = time.Now().Unix()
	}

	log.Debug("period and estimate", zap.Any("period", period), zap.Any("estimate", estimate))
	// Attempt create document
	meta, err := ctrl.col.CreateDocument(driver.WithWaitForSync(ctx, true), i)
	if err != nil {
		log.Error("Failed to create Instance", zap.Error(err))
		return "", err
	}
	i.Uuid = meta.Key

	var event = &elpb.Event{
		Entity:    INSTANCES_COL,
		Uuid:      i.GetUuid(),
		Scope:     "database",
		Action:    "create",
		Rc:        0,
		Requestor: ctx.Value(nocloud.NoCloudAccount).(string),
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	nocloud.Log(log, event)

	// Attempt create edge
	_, err = ctrl.ig2inst.CreateDocument(ctx, Access{
		From: group, To: meta.ID,
		Role: roles.OWNER,
	})
	if err != nil {
		log.Error("Failed to create Edge", zap.Error(err)) // if failed - remove instance from DataBase
		if _, err = ctrl.col.RemoveDocument(ctx, meta.Key); err != nil {
			log.Warn("Failed to cleanup", zap.String("uuid", meta.Key), zap.Error(err))
		}
		return "", err
	}

	if i.Config == nil {
		i.Config = make(map[string]*structpb.Value)
	}
	// Send for ansible hooks
	c := pb.Context{
		Instance: i.GetUuid(),
		Sp:       sp,
		Event:    spb.NoCloudStatus_INIT.String(),
	}
	if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
		log.Error("Failed to publish ansible hook", zap.Error(err))
	}
	if i.GetConfig()["auto_start"].GetBoolValue() {
		c := pb.Context{
			Instance: i.GetUuid(),
			Sp:       sp,
			Event:    "START",
		}
		if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
			log.Error("Failed to publish ansible hook", zap.Error(err))
		}
		for _, a := range i.GetAddons() {
			addon, err := ctrl.addons.Get(ctx, a)
			if err != nil {
				log.Error("Failed to get instance addon", zap.Error(err), zap.String("addon", a))
				continue
			}
			if addon.Action != nil && addon.Action.GetPlaybook() != "" {
				c := pb.Context{
					Instance: i.GetUuid(),
					Sp:       sp,
					Event:    "START",
					Addon:    &a,
				}
				if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
					log.Error("Failed to publish ansible hook addon event", zap.Error(err))
				}
			}
		}
	}
	e := epb.Event{
		Uuid: i.GetUuid(),
		Key:  services_registry.InstanceCreated,
		Data: make(map[string]*structpb.Value),
	}
	if promo, ok := ctx.Value(CreationPromocodeKey).(string); ok {
		e.Data["promocode"] = structpb.NewStringValue(promo)
	}
	if err = ctrl.ps.Publish(ps.DEFAULT_EXCHANGE, services_registry.Topic("instances"), &e); err != nil {
		log.Error("Failed to publish instance creating", zap.Error(err))
	}

	return meta.Key, nil
}

const removeDataQuery = `
UPDATE DOCUMENT(@key) WITH { data: null } IN @@collection 
`

const updateDataQuery = `
UPDATE DOCUMENT(@key) WITH { data: @data } IN @@collection 
`

const removePlanQuery = `
UPDATE DOCUMENT(@key) WITH { billing_plan: null } IN @@collection 
`

const updatePlanQuery = `
UPDATE DOCUMENT(@key) WITH { billing_plan: @billingPlan } IN @@collection
`

func (ctrl *instancesController) Update(ctx context.Context, _ string, inst, oldInst *pb.Instance) error {
	log := ctrl.log.Named("Update")
	log.Debug("Updating Instance", zap.Any("instance", inst))

	uuid := inst.GetUuid()

	if oldInst.GetStatus() == spb.NoCloudStatus_DEL {
		log.Info("Inst cannot be updated. Status DEL", zap.String("uuid", oldInst.GetUuid()))
		return nil
	}
	inst.Uuid = ""
	inst.Status = spb.NoCloudStatus_INIT
	inst.State = nil

	err := hasher.SetHash(inst.ProtoReflect())
	if err != nil {
		return err
	}

	ctrl.log.Debug("instance for hash calculating while Updating", zap.Any("inst", inst))

	// Recalculate estimate price and period
	// Values would change if plan was updated or replaced
	estimate, err := ctrl.CalculateInstanceEstimatePrice(inst, false)
	if err != nil {
		log.Error("Failed to calculate estimate instance periodic price", zap.Error(err))
		return err
	}
	period, err := ctrl.GetInstancePeriod(inst)
	if err != nil {
		log.Error("Failed to get instance period", zap.Error(err))
		return err
	}

	log.Debug("period and estimate", zap.Any("period", period), zap.Any("estimate", estimate))
	mask := &pb.Instance{
		Config:    inst.GetConfig(),
		Resources: inst.GetResources(),
		Hash:      inst.GetHash(),
		Period:    period,
		Estimate:  estimate,
		Meta:      inst.Meta,
	}

	if inst.GetTitle() != oldInst.GetTitle() {
		mask.Title = inst.GetTitle()
	}

	if !reflect.DeepEqual(inst.GetAddons(), oldInst.GetAddons()) {
		mask.Addons = inst.GetAddons()
	}

	if inst.GetProduct() != oldInst.GetProduct() {
		mask.Product = inst.Product
	}

	if inst.GetCreated() != oldInst.GetCreated() {
		mask.Created = inst.GetCreated()
	}

	equalPlans := reflect.DeepEqual(inst.GetBillingPlan(), oldInst.GetBillingPlan())

	if !equalPlans {
		log.Debug("Update plan")
		_, err := ctrl.db.Query(ctx, removePlanQuery, map[string]interface{}{
			"@collection": schema.INSTANCES_COL,
			"key":         driver.NewDocumentID(schema.INSTANCES_COL, oldInst.Uuid),
		})
		if err != nil {
			log.Error("Failed to remove plan")
			return err
		}

		_, err = ctrl.db.Query(ctx, updatePlanQuery, map[string]interface{}{
			"@collection": schema.INSTANCES_COL,
			"key":         driver.NewDocumentID(schema.INSTANCES_COL, oldInst.Uuid),
			"billingPlan": inst.GetBillingPlan(),
		})
		if err != nil {
			log.Error("Failed to update plan")
			return err
		}
	}

	equalDatas := reflect.DeepEqual(inst.GetData(), oldInst.GetData())

	if !equalDatas {
		_, err := ctrl.db.Query(ctx, removeDataQuery, map[string]interface{}{
			"@collection": schema.INSTANCES_COL,
			"key":         driver.NewDocumentID(schema.INSTANCES_COL, oldInst.Uuid),
		})
		if err != nil {
			log.Error("Failed to remove data")
			return err
		}

		_, err = ctrl.db.Query(ctx, updateDataQuery, map[string]interface{}{
			"@collection": schema.INSTANCES_COL,
			"key":         driver.NewDocumentID(schema.INSTANCES_COL, oldInst.Uuid),
			"data":        inst.Data,
		})
		if err != nil {
			log.Error("Failed to update data")
			return err
		}
	}

	if !oldInst.GetConfig()["auto_start"].GetBoolValue() && inst.GetConfig()["auto_start"].GetBoolValue() {
		if mask.Meta == nil {
			mask.Meta = &pb.InstanceMeta{}
		}
		mask.Meta.Started = time.Now().Unix()
	}

	_, err = ctrl.col.UpdateDocument(ctx, oldInst.Uuid, mask)
	if err != nil {
		log.Error("Failed to update Instance", zap.Error(err))
		return err
	}

	instMarshal, _ := json.Marshal(inst)
	oldInstMarshal, _ := json.Marshal(oldInst)
	diff, err := jsondiff.CompareJSON(oldInstMarshal, instMarshal)
	if err != nil {
		log.Error("Failed to calculate diff", zap.Error(err))
		return err
	}

	var event = &elpb.Event{
		Entity:    INSTANCES_COL,
		Uuid:      uuid,
		Scope:     "database",
		Action:    "update",
		Rc:        0,
		Requestor: ctx.Value(nocloud.NoCloudAccount).(string),
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: diff.String(),
		},
	}

	nocloud.Log(log, event)

	sp, err := ctrl.getSp(ctx, uuid)
	if err != nil {
		log.Error("Failed to get sp to publish ansible hook", zap.Error(err))
		return nil
	}

	if inst.Config == nil {
		inst.Config = make(map[string]*structpb.Value)
	}
	if oldInst.Config == nil {
		oldInst.Config = make(map[string]*structpb.Value)
	}
	if !oldInst.GetConfig()["auto_start"].GetBoolValue() && inst.GetConfig()["auto_start"].GetBoolValue() {
		c := pb.Context{
			Instance: uuid,
			Sp:       sp,
			Event:    "START",
		}
		if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
			log.Error("Failed to publish ansible hook", zap.Error(err))
		}
		for _, a := range inst.GetAddons() {
			addon, err := ctrl.addons.Get(ctx, a)
			if err != nil {
				log.Error("Failed to get instance addon", zap.Error(err), zap.String("addon", a))
				continue
			}
			if addon.Action != nil && addon.Action.GetPlaybook() != "" {
				c := pb.Context{
					Instance: inst.GetUuid(),
					Sp:       sp,
					Event:    "START",
					Addon:    &a,
				}
				if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
					log.Error("Failed to publish ansible hook addon event", zap.Error(err))
				}
			}
		}
	}
	c := pb.Context{
		Instance: uuid,
		Sp:       sp,
		Event:    "UPDATE",
	}
	if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
		log.Error("Failed to publish ansible hook", zap.Error(err))
	}

	return nil
}

func (ctrl *instancesController) UpdateWithPatch(ctx context.Context, _ string, inst, oldInst *pb.Instance) error {
	log := ctrl.log.Named("UpdateWithPatch")
	log.Debug("Updating Instance", zap.Any("instance", inst))
	uuid := inst.GetUuid()

	if oldInst.GetStatus() == spb.NoCloudStatus_DEL {
		log.Info("Inst cannot be updated. Status DEL", zap.String("uuid", oldInst.GetUuid()))
		return nil
	}
	inst.Uuid = ""
	inst.Status = spb.NoCloudStatus_INIT
	inst.State = nil

	err := hasher.SetHash(inst.ProtoReflect())
	if err != nil {
		return err
	}

	// Recalculate estimate price and period
	// Values would change if plan was updated or replaced
	estimate, err := ctrl.CalculateInstanceEstimatePrice(inst, false)
	if err != nil {
		log.Error("Failed to calculate estimate instance periodic price", zap.Error(err))
		return err
	}
	period, err := ctrl.GetInstancePeriod(inst)
	if err != nil {
		log.Error("Failed to get instance period", zap.Error(err))
		return err
	}

	mask := &pb.Instance{
		Config:    inst.Config,
		Resources: inst.Resources,
		Hash:      inst.Hash,
		Period:    period,
		Estimate:  estimate,
		Meta:      inst.Meta,
	}
	if inst.GetTitle() != oldInst.GetTitle() {
		mask.Title = inst.GetTitle()
	}
	if !reflect.DeepEqual(inst.GetAddons(), oldInst.GetAddons()) {
		mask.Addons = inst.GetAddons()
	}
	if inst.GetProduct() != oldInst.GetProduct() {
		mask.Product = inst.Product
	}
	if inst.GetCreated() != oldInst.GetCreated() {
		mask.Created = inst.GetCreated()
	}

	equalPlans := reflect.DeepEqual(inst.GetBillingPlan(), oldInst.GetBillingPlan())
	if !equalPlans && inst.BillingPlan != nil {
		log.Debug("Update plan")
		_, err := ctrl.db.Query(ctx, removePlanQuery, map[string]interface{}{
			"@collection": schema.INSTANCES_COL,
			"key":         driver.NewDocumentID(schema.INSTANCES_COL, oldInst.Uuid),
		})
		if err != nil {
			log.Error("Failed to remove plan")
			return err
		}

		_, err = ctrl.db.Query(ctx, updatePlanQuery, map[string]interface{}{
			"@collection": schema.INSTANCES_COL,
			"key":         driver.NewDocumentID(schema.INSTANCES_COL, oldInst.Uuid),
			"billingPlan": inst.GetBillingPlan(),
		})
		if err != nil {
			log.Error("Failed to update plan")
			return err
		}
	}

	equalDatas := reflect.DeepEqual(inst.GetData(), oldInst.GetData())
	if !equalDatas && inst.Data != nil {
		_, err := ctrl.db.Query(ctx, removeDataQuery, map[string]interface{}{
			"@collection": schema.INSTANCES_COL,
			"key":         driver.NewDocumentID(schema.INSTANCES_COL, oldInst.Uuid),
		})
		if err != nil {
			log.Error("Failed to remove data")
			return err
		}

		_, err = ctrl.db.Query(ctx, updateDataQuery, map[string]interface{}{
			"@collection": schema.INSTANCES_COL,
			"key":         driver.NewDocumentID(schema.INSTANCES_COL, oldInst.Uuid),
			"data":        inst.Data,
		})
		if err != nil {
			log.Error("Failed to update data")
			return err
		}
	}

	if inst.Config != nil && !oldInst.GetConfig()["auto_start"].GetBoolValue() && inst.GetConfig()["auto_start"].GetBoolValue() {
		if mask.Meta == nil {
			mask.Meta = &pb.InstanceMeta{}
		}
		mask.Meta.Started = time.Now().Unix()
	}

	toUpdate := ToMapClean(mask)
	log.Debug("Instance patch", zap.Any("patch", toUpdate))
	_, err = ctrl.col.UpdateDocument(ctx, oldInst.Uuid, toUpdate)
	if err != nil {
		log.Error("Failed to update Instance", zap.Error(err))
		return err
	}

	instMarshal, _ := json.Marshal(inst)
	oldInstMarshal, _ := json.Marshal(oldInst)
	diff, err := jsondiff.CompareJSON(oldInstMarshal, instMarshal)
	if err != nil {
		log.Error("Failed to calculate diff", zap.Error(err))
		return err
	}

	var event = &elpb.Event{
		Entity:    INSTANCES_COL,
		Uuid:      uuid,
		Scope:     "database",
		Action:    "update",
		Rc:        0,
		Requestor: ctx.Value(nocloud.NoCloudAccount).(string),
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: diff.String(),
		},
	}

	nocloud.Log(log, event)

	sp, err := ctrl.getSp(ctx, uuid)
	if err != nil {
		log.Error("Failed to get sp to publish ansible hook", zap.Error(err))
		return nil
	}

	if oldInst.Config == nil {
		oldInst.Config = make(map[string]*structpb.Value)
	}
	if inst.Config != nil && !oldInst.GetConfig()["auto_start"].GetBoolValue() && inst.GetConfig()["auto_start"].GetBoolValue() {
		c := pb.Context{
			Instance: uuid,
			Sp:       sp,
			Event:    "START",
		}
		if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
			log.Error("Failed to publish ansible hook", zap.Error(err))
		}
		for _, a := range inst.GetAddons() {
			addon, err := ctrl.addons.Get(ctx, a)
			if err != nil {
				log.Error("Failed to get instance addon", zap.Error(err), zap.String("addon", a))
				continue
			}
			if addon.Action != nil && addon.Action.GetPlaybook() != "" {
				c := pb.Context{
					Instance: inst.GetUuid(),
					Sp:       sp,
					Event:    "START",
					Addon:    &a,
				}
				if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
					log.Error("Failed to publish ansible hook addon event", zap.Error(err))
				}
			}
		}
	}
	c := pb.Context{
		Instance: uuid,
		Sp:       sp,
		Event:    "UPDATE",
	}
	if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
		log.Error("Failed to publish ansible hook", zap.Error(err))
	}

	return nil
}

func (ctrl *instancesController) UpdateNotes(ctx context.Context, inst *pb.Instance) error {
	log := ctrl.log.Named("UpdateNotes")
	log.Debug("Updating Instance", zap.Any("instance", inst))

	if len(inst.GetAdminNotes()) == 0 {
		_, err := ctrl.col.UpdateDocument(ctx, inst.Uuid, map[string]interface{}{
			"admin_notes": make([]*notes.AdminNote, 0),
		})
		if err != nil {
			log.Error("Failed to update Instance", zap.Error(err))
			return err
		}
		return nil
	}

	_, err := ctrl.col.UpdateDocument(ctx, inst.Uuid, inst)
	if err != nil {
		log.Error("Failed to update Instance", zap.Error(err))
		return err
	}

	return nil
}

func (ctrl *instancesController) Delete(ctx context.Context, group string, i *pb.Instance) error {
	log := ctrl.log.Named("Delete")
	log.Debug("Deleting Instance", zap.Any("instance", i))

	_, err := ctrl.col.RemoveDocument(ctx, i.Uuid)
	if err != nil {
		log.Error("Failed to delete Instance", zap.Error(err))
		return err
	}

	ctrl.log.Debug("Deleting Edge", zap.String("fromCollection", schema.INSTANCES_GROUPS_COL), zap.String("toCollection",
		schema.INSTANCES_COL), zap.String("fromKey", group), zap.String("toKey", i.GetUuid()))
	err = deleteEdge(ctx, ctrl.col.Database(), schema.INSTANCES_GROUPS_COL, schema.INSTANCES_COL, group, i.GetUuid())
	if err != nil {
		log.Error("Failed to delete edge "+schema.INSTANCES_GROUPS_COL+"2"+schema.INSTANCES_COL, zap.Error(err))
		return err
	}

	sp, err := ctrl.getSp(ctx, i.GetUuid())
	if err != nil {
		log.Error("Failed to get sp to publish ansible hook", zap.Error(err))
		return nil
	}
	c := pb.Context{
		Instance: i.GetUuid(),
		Sp:       sp,
		Event:    spb.NoCloudStatus_DEL.String(),
	}
	if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
		log.Error("Failed to publish ansible hook", zap.Error(err))
	}

	return nil
}

func (ctrl *instancesController) Get(ctx context.Context, uuid string) (*Instance, error) {
	ctrl.log.Debug("Getting Instance", zap.Any("sp", uuid))
	var inst *pb.Instance
	query := `RETURN DOCUMENT(@inst)`
	c, err := ctrl.col.Database().Query(ctx, query, map[string]interface{}{
		"inst": driver.NewDocumentID(schema.INSTANCES_COL, uuid),
	})
	if err != nil {
		ctrl.log.Debug("Error reading document(Instance)", zap.Error(err))
		return nil, err
	}
	defer c.Close()

	meta, err := c.ReadDocument(ctx, &inst)

	if inst == nil {
		return nil, err
	}

	// If values not presented in existing instance calculate estimate and period dynamically
	// TODO: make migrations or smth instead of calculating it everytime
	if inst.GetEstimate() == 0 {
		inst.Estimate, _ = ctrl.CalculateInstanceEstimatePrice(inst, false)
	}
	if inst.GetPeriod() == 0 {
		inst.Period, _ = ctrl.GetInstancePeriod(inst)
	}

	inst.Uuid = uuid
	return &Instance{inst, meta}, nil
}

const getGroupWithSPQuery = `
LET instance = DOCUMENT(@instance)
LET group = (
    FOR group IN 1 INBOUND instance
    GRAPH @permissions
        RETURN group )[0]

LET sp = (
    FOR s IN 1 OUTBOUND group
    GRAPH @permissions
    FILTER IS_SAME_COLLECTION(@sps, s)
        RETURN s )[0]
        
RETURN {
  group: MERGE(group, { uuid: group._key }),
  sp: MERGE(sp, { uuid: sp._key })
}`

type GroupWithSP struct {
	Group *pb.InstancesGroup     `json:"group"`
	SP    *sppb.ServicesProvider `json:"sp"`
}

func (ctrl *instancesController) GetGroup(ctx context.Context, i string) (*GroupWithSP, error) {
	log := ctrl.log.Named("GetGroup")
	log.Debug("Getting Instance Group", zap.String("instance", i))
	c, err := ctrl.db.Query(ctx, getGroupWithSPQuery, map[string]interface{}{
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"sps":         schema.SERVICES_PROVIDERS_COL,
		"instance":    i,
	})
	if err != nil {
		log.Error("Error while querying", zap.Error(err))
		return nil, err
	}
	defer c.Close()

	var r GroupWithSP
	_, err = c.ReadDocument(ctx, &r)
	if err != nil {
		log.Error("Error while reading document", zap.Error(err))
		return nil, err
	}

	return &r, nil
}

func (ctrl *instancesController) CheckEdgeExist(ctx context.Context, spUuid string, i *pb.Instance) error {
	log := ctrl.log.Named("ValidateBillingPlan").Named(i.Title)
	if i.BillingPlan == nil {
		log.Debug("Billing plan is not provided, skipping")
		return nil
	}

	ok, err := edgeExist(ctx, ctrl.db, schema.SERVICES_PROVIDERS_COL, schema.BILLING_PLANS_COL, spUuid, i.BillingPlan.Uuid)
	if err != nil {
		return err
	}
	if !ok {
		ctrl.log.Error("SP and Billing Plan are not binded", zap.Any("sp", spUuid), zap.Any("plan", i.BillingPlan.Uuid))
		return errors.New("SP and Billing Plan are not binded")
	}

	return nil
}

func (ctrl *instancesController) ValidateBillingPlan(ctx context.Context, spUuid string, i *pb.Instance) error {
	log := ctrl.log.Named("ValidateBillingPlan").Named(i.Title)
	if i.BillingPlan == nil {
		log.Debug("Billing plan is not provided, skipping")
		return nil
	}

	if i.BillingPlan.Software != nil {
	check_software:
		for _, s := range i.BillingPlan.Software {
			for _, is := range i.Software {
				if s.Playbook == is.Playbook {
					log.Debug("Software is valid", zap.String("software", s.String()))
					continue check_software
				}
			}
			return fmt.Errorf("software %s is not defined in Instance", s.Playbook)
		}
	}

	if i.BillingPlan.Kind < 2 { // If Kind is Dynamic or Unknown
		log.Debug("Billing plan Dynamic, nothing else to validate")
		i.BillingPlan.Kind = bpb.PlanKind_DYNAMIC // Ensuring Kind is set
		return nil
	}

	if i.BillingPlan.Kind == bpb.PlanKind_STATIC {
		log.Debug("Billing plan is Static, checking if it is valid")
		if i.Product == nil {
			return errors.New("product is required for static billing plan")
		}

		product, ok := i.BillingPlan.Products[*i.Product]
		if !ok {
			return fmt.Errorf("product %s is not defined in billing plan", *i.Product)
		}

		for key, amount := range product.Resources {
			res, ok := i.Resources[key]
			if !ok {
				return fmt.Errorf("resource %s is not defined in instance", key)
			}
			if res.AsInterface() != amount.AsInterface() {
				return fmt.Errorf("resource %s has different amount in billing plan and instance: %v != %v", key, res, amount)
			}
		}

		return nil
	}

	return nil
}

func (ctrl *instancesController) SetStatus(ctx context.Context, inst *pb.Instance, status spb.NoCloudStatus) (err error) {
	log := ctrl.log.Named("SetStatus")

	mask := &pb.Instance{
		Status: status,
	}
	if status == spb.NoCloudStatus_DEL {
		mask.Deleted = time.Now().Unix()
	}
	_, err = ctrl.col.UpdateDocument(ctx, inst.Uuid, mask)
	if err != nil {
		log.Error("Failed to update", zap.Error(err))
		return err
	}

	sp, err := ctrl.getSp(ctx, inst.GetUuid())
	if err != nil {
		log.Error("Failed to get sp", zap.Error(err))
		return nil
	}
	c := pb.Context{
		Instance: inst.GetUuid(),
		Sp:       sp,
		Event:    status.String(),
	}
	if err = ctrl.ansPs.Publish("hooks", services_registry.Topic("ansible_hooks"), &c); err != nil {
		log.Error("Failed to publish ansible hook", zap.Error(err))
	}

	return nil
}

func (ctrl *instancesController) SetState(ctx context.Context, inst *pb.Instance, state stpb.NoCloudState) (err error) {
	mask := &pb.Instance{
		State: &stpb.State{
			State: state,
		},
	}
	_, err = ctrl.col.UpdateDocument(ctx, inst.Uuid, mask)
	return err
}

func (ctrl *instancesController) TransferInst(ctx context.Context, oldIGEdge string, newIG driver.DocumentID, inst driver.DocumentID) error {
	log := ctrl.log.Named("Transfer")
	log.Debug("Transfer InstancesGroup", zap.String("group", inst.String()), zap.String("srvEdge", oldIGEdge), zap.String("to", newIG.String()))

	_, err := ctrl.ig2inst.RemoveDocument(ctx, oldIGEdge)
	if err != nil {
		log.Error("Failed to remove old Edge", zap.Error(err))
		return err
	}

	_, err = ctrl.ig2inst.CreateDocument(ctx, Access{From: newIG, To: inst, Role: roles.OWNER})
	if err != nil {
		log.Error("Failed to create Edge", zap.Error(err))
		return err
	}

	return nil
}

func (ctrl *instancesController) GetEdge(ctx context.Context, inboundNode string, collection string) (string, error) {
	log := ctrl.log.Named("GetEdge")
	log.Debug("Getting edge", zap.String("nodeId", inboundNode))
	c, err := ctrl.db.Query(ctx, getEdge, map[string]interface{}{
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"inboundNode": inboundNode,
		"collection":  collection,
	})

	if err != nil {
		log.Error("Error while querying", zap.Error(err))
		return "", err
	}
	defer c.Close()
	var edgeId string
	_, err = c.ReadDocument(ctx, &edgeId)
	if err != nil {
		log.Error("Error while reading document", zap.Error(err))
		return "", err
	}

	return edgeId, nil
}

const getSp = `
LET inboundNode = DOCUMENT(@node)

LET ig = LAST(
	FOR ig IN 1 Inbound inboundNode
	GRAPH @permissions
	FILTER IS_SAME_COLLECTION(@ig, ig)
	RETURN ig
)

LET sp = LAST(
	FOR sp IN 1 OUTBOUND ig
	GRAPH @permissions
	FILTER IS_SAME_COLLECTION(@sp, sp)
	RETURN sp
)

return sp._key
`

func (ctrl *instancesController) getSp(ctx context.Context, uuid string) (string, error) {
	log := ctrl.log.Named("GetSp")
	c, err := ctrl.db.Query(ctx, getSp, map[string]interface{}{
		"node":        driver.NewDocumentID(schema.INSTANCES_COL, uuid),
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"ig":          schema.INSTANCES_GROUPS_COL,
		"sp":          schema.SERVICES_PROVIDERS_COL,
	})

	if err != nil {
		log.Error("Error while querying", zap.Error(err))
		return "", err
	}
	defer c.Close()
	var sp string
	_, err = c.ReadDocument(ctx, &sp)
	if err != nil {
		log.Error("Error while reading document", zap.Error(err))
		return "", err
	}

	return sp, nil
}

const getInstanceWithAccessLevel = `
FOR path IN OUTBOUND K_SHORTEST_PATHS @account TO @node
GRAPH @permissions SORT path.edges[0].level
	LET bp = DOCUMENT(CONCAT(@bps, "/", path.vertices[-1].billing_plan.uuid))
    RETURN MERGE(path.vertices[-1], {
        uuid: path.vertices[-1]._key,
        billing_plan: {
			uuid: bp._key,
			title: bp.title,
			type: bp.type,
			kind: bp.kind,
			resources: bp.resources,
			products: {
			    [path.vertices[-1].product]: bp.products[path.vertices[-1].product],
            },
			meta: bp.meta,
			fee: bp.fee,
			software: bp.software,
            addons: bp.addons,
            properties: bp.properties
        },
	    access: {level: path.edges[0].level ? : 0, role: path.edges[0].role ? : "none", namespace: path.vertices[-2]._key }
	})
`

func ToMapClean(src any) map[string]any {
	res := make(map[string]any)

	val := reflect.ValueOf(src)
	if !val.IsValid() {
		return res
	}
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return res
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return res
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		var name = typ.Field(i).Name
		if jsonTag := typ.Field(i).Tag.Get("json"); jsonTag != "" {
			split := strings.Split(jsonTag, ",")
			var fieldName string
			for _, v := range split {
				if v == "omitempty" || v == "-" {
					continue
				}
				fieldName = v
				break
			}
			if fieldName != "" {
				name = fieldName
			}
		}

		if !field.CanInterface() {
			continue
		}

		add := func(v any) { res[name] = v }

		switch field.Kind() {
		case reflect.Pointer:
			if field.IsNil() {
				continue
			}
			if field.Elem().Kind() == reflect.Struct {
				if ma := ToMapClean(field.Interface()); len(ma) != 0 {
					add(ma)
				}
			} else if !isZero(field.Elem()) {
				add(field.Elem().Interface())
			}

		case reflect.Struct:
			if ma := ToMapClean(field.Interface()); len(ma) != 0 {
				add(ma)
			}

		case reflect.Slice, reflect.Array:
			if field.Len() == 0 {
				continue
			}
			var out []any
			for j := 0; j < field.Len(); j++ {
				item := field.Index(j)
				if item.Kind() == reflect.Struct ||
					(item.Kind() == reflect.Pointer && item.Elem().Kind() == reflect.Struct) {
					if ma := ToMapClean(item.Interface()); len(ma) != 0 {
						out = append(out, ma)
					}
				} else if !isZero(item) {
					out = append(out, item.Interface())
				}
			}
			if len(out) != 0 {
				add(out)
			}

		case reflect.Map:
			if field.Len() == 0 {
				continue
			}
			mp := make(map[string]any)
			iter := field.MapRange()
			for iter.Next() {
				key := fmt.Sprint(iter.Key().Interface())
				v := iter.Value()
				if v.Kind() == reflect.Struct ||
					(v.Kind() == reflect.Pointer && v.Elem().Kind() == reflect.Struct) {
					if ma := ToMapClean(v.Interface()); len(ma) != 0 {
						mp[key] = ma
					}
				} else if !isZero(v) {
					mp[key] = v.Interface()
				}
			}
			if len(mp) != 0 {
				add(mp)
			}

		default:
			if !isZero(field) {
				add(field.Interface())
			}
		}
	}
	return res
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return false // Do not consider false as zero value
	case reflect.String:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	default:
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
