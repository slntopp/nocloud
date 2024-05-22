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
	"reflect"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/wI2L/jsondiff"
	"go.uber.org/zap"

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

const (
	INSTANCES_COL = "Instances"
)

type Instance struct {
	*pb.Instance
	driver.DocumentMeta
}

type InstancesController struct {
	col   driver.Collection // Instances Collection
	graph driver.Graph

	log *zap.Logger

	db driver.Database

	ig2inst driver.Collection

	bp_ctrl *BillingPlansController
}

func NewInstancesController(log *zap.Logger, db driver.Database) *InstancesController {
	ctx := context.TODO()

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GetEnsureCollection(log, ctx, db, schema.INSTANCES_COL)
	ig2inst := GraphGetEdgeEnsure(log, ctx, graph, schema.IG2INST, schema.INSTANCES_GROUPS_COL, schema.INSTANCES_COL)

	bp_ctrl := NewBillingPlansController(log, db)

	return &InstancesController{log: log.Named("InstancesController"), col: col, graph: graph, db: db, ig2inst: ig2inst,
		bp_ctrl: &bp_ctrl}
}

// CalculateInstanceEstimatePeriodicPrice return estimate periodic price for current instance in NCU currency
func (ctrl *InstancesController) CalculateInstanceEstimatePeriodicPrice(i *pb.Instance) (float64, error) {
	plan, err := ctrl.bp_ctrl.Get(context.Background(), i.GetBillingPlan())
	if err != nil {
		return 0, err
	}

	cost := 0.0

	for _, res := range plan.GetResources() {

		if res.GetPeriod() == 0 {
			continue
		}

		switch res.GetKey() {
		case "cpu":
			value := i.GetResources()["cpu"].GetNumberValue()
			cost += value * res.GetPrice()
		case "ram":
			value := i.GetResources()["ram"].GetNumberValue() / 1024
			cost += value * res.GetPrice()
		case "ips_public":
			value := i.GetResources()["ips_public"].GetNumberValue()
			cost += value * res.GetPrice()
		case "ips_private":
			value := i.GetResources()["ips_private"].GetNumberValue()
			cost += value * res.GetPrice()
		default:
			// Calculate drive size billing
			if strings.Contains(res.GetKey(), "drive") {
				driveType := i.GetResources()["drive_type"].GetStringValue()
				if res.GetKey() != "drive_"+strings.ToLower(driveType) {
					continue
				}
				count := i.GetResources()["drive_size"].GetNumberValue() / 1024
				cost += res.GetPrice() * count
			}
		}
	}
	return cost, nil
}

// GetInstancePeriod returns billing period for the whole instance
//
// Now it simply returns product's period or period of some random resource if product is not defined or it's period is 0
func (ctrl *InstancesController) GetInstancePeriod(i *pb.Instance) (int64, error) {
	plan, err := ctrl.bp_ctrl.Get(context.Background(), i.GetBillingPlan())
	if err != nil {
		return 0, err
	}

	product, ok := plan.GetProducts()[i.GetProduct()]
	if ok {
		if product.GetPeriod() > 0 {
			return product.GetPeriod(), nil
		}
	}

	for _, res := range plan.GetResources() {
		if _, ok := i.GetResources()[res.GetKey()]; ok && res.GetPeriod() > 0 {
			return res.GetPeriod(), nil
		}
	}

	return 0, nil
}

func (ctrl *InstancesController) Create(ctx context.Context, group driver.DocumentID, sp string, i *pb.Instance) error {
	log := ctrl.log.Named("Create")
	log.Debug("Creating Instance", zap.Any("instance", i))

	// ensure status is INIT
	i.Uuid = ""
	i.Status = spb.NoCloudStatus_INIT
	i.Created = time.Now().Unix()

	// Set estimate and period values
	estimate, err := ctrl.CalculateInstanceEstimatePeriodicPrice(i)
	if err != nil {
		log.Error("Failed to calculate estimate instance periodic price", zap.Error(err))
		return err
	}
	i.Estimate = estimate
	period, err := ctrl.GetInstancePeriod(i)
	if err != nil {
		log.Error("Failed to get instance period", zap.Error(err))
		return err
	}
	i.Period = period

	err = hasher.SetHash(i.ProtoReflect())
	if err != nil {
		log.Error("Failed to calculate hash", zap.Error(err))
		return err
	}

	ctrl.log.Debug("instance for hash calculating while Creating", zap.Any("inst", i))

	// Attempt create document
	meta, err := ctrl.col.CreateDocument(ctx, i)
	if err != nil {
		log.Error("Failed to create Instance", zap.Error(err))
		return err
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
		return err
	}

	return nil
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

func (ctrl *InstancesController) Update(ctx context.Context, sp string, inst, oldInst *pb.Instance) error {
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
	estimate, err := ctrl.CalculateInstanceEstimatePeriodicPrice(inst)
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
		Config:    inst.GetConfig(),
		Resources: inst.GetResources(),
		Hash:      inst.GetHash(),
		Period:    period,
		Estimate:  estimate,
	}

	if inst.GetTitle() != oldInst.GetTitle() {
		mask.Title = inst.GetTitle()
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

	return nil
}

func (ctrl *InstancesController) UpdateNotes(ctx context.Context, inst *pb.Instance) error {
	log := ctrl.log.Named("UpdateNotes")
	log.Debug("Updating Instance", zap.Any("instance", inst))

	_, err := ctrl.col.UpdateDocument(ctx, inst.Uuid, inst)
	if err != nil {
		log.Error("Failed to update Instance", zap.Error(err))
		return err
	}

	return nil
}

func (ctrl *InstancesController) Delete(ctx context.Context, group string, i *pb.Instance) error {
	log := ctrl.log.Named("Delete")
	log.Debug("Deleting Instance", zap.Any("instance", i))

	_, err := ctrl.col.RemoveDocument(ctx, i.Uuid)
	if err != nil {
		log.Error("Failed to delete Instance", zap.Error(err))
		return err
	}

	ctrl.log.Debug("Deleting Edge", zap.String("fromCollection", schema.INSTANCES_GROUPS_COL), zap.String("toCollection",
		schema.INSTANCES_COL), zap.String("fromKey", group), zap.String("toKey", i.GetUuid()))
	err = DeleteEdge(ctx, ctrl.col.Database(), schema.INSTANCES_GROUPS_COL, schema.INSTANCES_COL, group, i.GetUuid())
	if err != nil {
		log.Error("Failed to delete edge "+schema.INSTANCES_GROUPS_COL+"2"+schema.INSTANCES_COL, zap.Error(err))
		return err
	}

	return nil
}

func (ctrl *InstancesController) Get(ctx context.Context, uuid string) (*Instance, error) {
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
	ctrl.log.Debug("ReadDocument.Result", zap.Any("meta", meta), zap.Error(err), zap.Any("isnt", &inst))

	if inst == nil {
		return nil, err
	}

	// If values not presented in existing instance calculate estimate and period dynamically
	// TODO: make migrations or smth instead of calculating it everytime
	if inst.GetEstimate() == 0 {
		inst.Estimate, _ = ctrl.CalculateInstanceEstimatePeriodicPrice(inst)
	}
	if inst.GetPeriod() == 0 {
		inst.Period, _ = ctrl.GetInstancePeriod(inst)
	}

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

func (ctrl *InstancesController) GetGroup(ctx context.Context, i string) (*GroupWithSP, error) {
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

func (ctrl *InstancesController) CheckEdgeExist(ctx context.Context, spUuid string, i *pb.Instance) error {
	log := ctrl.log.Named("ValidateBillingPlan").Named(i.Title)
	if i.BillingPlan == nil {
		log.Debug("Billing plan is not provided, skipping")
		return nil
	}

	ok, err := EdgeExist(ctx, ctrl.db, schema.SERVICES_PROVIDERS_COL, schema.BILLING_PLANS_COL, spUuid, i.BillingPlan.Uuid)
	if err != nil {
		return err
	}
	if !ok {
		ctrl.log.Error("SP and Billing Plan are not binded", zap.Any("sp", spUuid), zap.Any("plan", i.BillingPlan.Uuid))
		return errors.New("SP and Billing Plan are not binded")
	}

	return nil
}

func (ctrl *InstancesController) ValidateBillingPlan(ctx context.Context, spUuid string, i *pb.Instance) error {
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

func (ctrl *InstancesController) SetStatus(ctx context.Context, inst *pb.Instance, status spb.NoCloudStatus) (err error) {
	mask := &pb.Instance{
		Status: status,
	}
	if status == spb.NoCloudStatus_DEL {
		mask.Deleted = time.Now().Unix()
	}
	_, err = ctrl.col.UpdateDocument(ctx, inst.Uuid, mask)
	return err
}

func (ctrl *InstancesController) SetState(ctx context.Context, inst *pb.Instance, state stpb.NoCloudState) (err error) {
	mask := &pb.Instance{
		State: &stpb.State{
			State: state,
		},
	}
	_, err = ctrl.col.UpdateDocument(ctx, inst.Uuid, mask)
	return err
}

func (ctrl *InstancesController) TransferInst(ctx context.Context, oldIGEdge string, newIG driver.DocumentID, inst driver.DocumentID) error {
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

func (ctrl *InstancesController) GetEdge(ctx context.Context, inboundNode string, collection string) (string, error) {
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
