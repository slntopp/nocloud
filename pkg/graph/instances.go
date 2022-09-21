/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

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
	"errors"
	"fmt"

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"

	bpb "github.com/slntopp/nocloud/pkg/billing/proto"
	"github.com/slntopp/nocloud/pkg/hasher"
	pb "github.com/slntopp/nocloud/pkg/instances/proto"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	sppb "github.com/slntopp/nocloud/pkg/services_providers/proto"
)

const (
	INSTANCES_COL = "Instances"
)

type InstancesController struct {
	col   driver.Collection // Instances Collection
	graph driver.Graph

	log *zap.Logger

	db driver.Database

	ig2inst driver.Collection
}

func NewInstancesController(log *zap.Logger, db driver.Database) *InstancesController {
	ctx := context.TODO()

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.INSTANCES_COL)
	ig2inst := GraphGetEdgeEnsure(log, ctx, graph, schema.IG2INST, schema.INSTANCES_GROUPS_COL, schema.INSTANCES_COL)

	return &InstancesController{log: log.Named("InstancesController"), col: col, graph: graph, db: db, ig2inst: ig2inst}
}

func (ctrl *InstancesController) Create(ctx context.Context, group driver.DocumentID, sp string, i *pb.Instance) error {
	log := ctrl.log.Named("Create")
	log.Debug("Creating Instance", zap.Any("instance", i))

	// ensure status is INIT
	i.Uuid = ""
	i.Status = pb.InstanceStatus_INIT

	err := hasher.SetHash(i.ProtoReflect())
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

func (ctrl *InstancesController) Update(ctx context.Context, sp string, inst, oldInst *pb.Instance) error {
	log := ctrl.log.Named("Update")
	log.Debug("Updating Instance", zap.Any("instance", inst))

	inst.Uuid = ""
	inst.Status = pb.InstanceStatus_INIT
	inst.Data = nil
	inst.State = nil

	err := hasher.SetHash(inst.ProtoReflect())
	if err != nil {
		return err
	}

	ctrl.log.Debug("instance for hash calculating while Updating", zap.Any("inst", inst))

	mask := &pb.Instance{
		Config:    inst.GetConfig(),
		Resources: inst.GetResources(),
		Hash:      inst.GetHash(),
	}

	if inst.GetTitle() != oldInst.GetTitle() {
		mask.Title = inst.GetTitle()
	}

	_, err = ctrl.col.UpdateDocument(ctx, oldInst.Uuid, mask)
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

func (ctrl *InstancesController) ValidateBillingPlan(ctx context.Context, spUuid string, i *pb.Instance) error {
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

func (ctrl *InstancesController) SetStatus(ctx context.Context, inst *pb.Instance, status pb.InstanceStatus) (err error) {
	mask := &pb.Instance{
		Status: status,
	}
	_, err = ctrl.col.UpdateDocument(ctx, inst.Uuid, mask)
	return err
}
