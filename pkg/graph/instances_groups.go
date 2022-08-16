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
	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/proto"

	"github.com/slntopp/nocloud/pkg/hasher"
	pb "github.com/slntopp/nocloud/pkg/instances/proto"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

const (
	INSTANCES_GROUPS_COL = "InstancesGroups"
)

type InstancesGroupsController struct {
	col   driver.Collection // Instances Collection
	graph driver.Graph

	inst_ctrl *InstancesController

	log *zap.Logger
}

func NewInstancesGroupsController(log *zap.Logger, db driver.Database) *InstancesGroupsController {
	log.Debug("New InstancesGroups Controller Creating")

	ctx := context.TODO()

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.INSTANCES_GROUPS_COL)

	col.EnsurePersistentIndex(ctx, []string{"type"}, &driver.EnsurePersistentIndexOptions{
		Unique: false, Sparse: true, InBackground: true, Name: "sp-type",
	})

	GraphGetEdgeEnsure(log, ctx, graph, schema.SERV2IG, schema.SERVICES_COL, schema.INSTANCES_GROUPS_COL)
	GraphGetEdgeEnsure(log, ctx, graph, schema.IG2SP, schema.INSTANCES_GROUPS_COL, schema.SERVICES_PROVIDERS_COL)

	return &InstancesGroupsController{
		log: log.Named("InstancesGroupsController"), inst_ctrl: NewInstancesController(log, db),
		col: col, graph: graph,
	}
}

func (ctrl *InstancesGroupsController) Instances() *InstancesController {
	return ctrl.inst_ctrl
}

func (ctrl *InstancesGroupsController) Create(ctx context.Context, service driver.DocumentID, g *pb.InstancesGroup) error {
	log := ctrl.log.Named("Create")
	log.Debug("Creating InstancesGroup", zap.Any("group", g))

	g.Status = pb.InstanceStatus_INIT

	err := hasher.SetHash(g.ProtoReflect())
	if err != nil {
		return err
	}

	obj := proto.Clone(g).(*pb.InstancesGroup)
	obj.Instances = nil
	meta, err := ctrl.col.CreateDocument(ctx, obj)
	if err != nil {
		log.Error("Failed to create InstancesGroup", zap.Error(err))
		return err
	}
	g.Uuid = meta.Key

	edge, _, err := ctrl.graph.EdgeCollection(ctx, schema.SERV2IG)
	if err != nil {
		log.Error("Failed to get edge collection", zap.Error(err))
		ctrl.col.RemoveDocument(ctx, meta.Key)
		return err
	}

	_, err = edge.CreateDocument(ctx, Access{
		From: service, To: meta.ID,
	})
	if err != nil {
		log.Error("Failed to create Edge", zap.Error(err))
		ctrl.col.RemoveDocument(ctx, meta.Key)
		return err
	}

	for _, instance := range g.GetInstances() {
		err := ctrl.inst_ctrl.Create(ctx, meta.ID, *g.Sp, instance)
		if err != nil {
			log.Error("Failed to create Instance", zap.Error(err))
			continue
		}
	}

	return nil
}

func (ctrl *InstancesGroupsController) Delete(ctx context.Context, service string, g *pb.InstancesGroup) error {
	log := ctrl.log.Named("Delete")
	log.Debug("Deleting InstancesGroup", zap.Any("group", g))

	_, err := ctrl.col.RemoveDocument(ctx, g.GetUuid())
	if err != nil {
		log.Error("Failed to delete InstancesGroup", zap.Error(err))
		return err
	}

	ctrl.log.Debug("Deleting Edge", zap.String("fromCollection", schema.SERVICES_COL), zap.String("toCollection",
		schema.INSTANCES_GROUPS_COL), zap.String("fromKey", service), zap.String("toKey", g.GetUuid()))
	err = DeleteEdge(ctx, ctrl.col.Database(), schema.SERVICES_COL, schema.INSTANCES_GROUPS_COL, service, g.GetUuid())
	if err != nil {
		log.Error("Failed to delete edge "+schema.SERVICES_COL+"2"+schema.INSTANCES_GROUPS_COL, zap.Error(err))
		return err
	}

	for _, instance := range g.GetInstances() {
		err := ctrl.inst_ctrl.Delete(ctx, g.GetUuid(), instance)
		if err != nil {
			log.Error("Failed to delete Instance", zap.Error(err))
			continue
		}
	}

	return nil
}

func (ctrl *InstancesGroupsController) Update(ctx context.Context, ig, oldIg *pb.InstancesGroup) error {
	log := ctrl.log.Named("Update")
	log.Debug("Updating InstancesGroup", zap.Any("group", ig))

	// deleting missing instances
	for _, oldInst := range oldIg.GetInstances() {
		var oldInstFound = false
		for _, inst := range ig.GetInstances() {
			if oldInst.GetUuid() == inst.GetUuid() {
				oldInstFound = true
				break
			}
		}
		if !oldInstFound {
			err := ctrl.inst_ctrl.Delete(ctx, ig.GetUuid(), oldInst)
			if err != nil {
				log.Error("Error while deleting instance", zap.Error(err))
				return err
			}
		}
	}

	// creating missing and updating existing instances
	for _, inst := range ig.GetInstances() {
		var instFound = false
		for _, oldInst := range oldIg.GetInstances() {
			if inst.GetUuid() == oldInst.GetUuid() {
				instFound = true
				err := ctrl.inst_ctrl.Update(ctx, *oldIg.Sp, inst, oldInst)
				if err != nil {
					log.Error("Error while updating instance", zap.Error(err))
					return err
				}
				break
			}
		}
		if !instFound {
			docID := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, ig.Uuid)
			err := ctrl.inst_ctrl.Create(ctx, docID, *oldIg.Sp, inst)
			if err != nil {
				log.Error("Error while creating instance", zap.Error(err))
				return err
			}
		}
	}

	ig.Data = nil

	err := hasher.SetHash(ig.ProtoReflect())
	if err != nil {
		return err
	}

	mask := &pb.InstancesGroup{
		Uuid:      ig.GetUuid(),
		Config:    ig.GetConfig(),
		Resources: ig.GetResources(),
		Hash:      ig.GetHash(),
	}

	if ig.GetType() != oldIg.GetType() {
		mask.Type = ig.GetType()
	}
	if ig.GetTitle() != oldIg.GetTitle() {
		mask.Title = ig.GetTitle()
	}

	_, err = ctrl.col.UpdateDocument(ctx, mask.Uuid, mask)
	if err != nil {
		log.Debug("Error updating document(InstancesGroup)", zap.Error(err))
		return err
	}

	return nil
}
func (ctrl *InstancesGroupsController) Provide(ctx context.Context, group, sp string) error {
	edge, _, err := ctrl.graph.EdgeCollection(ctx, schema.IG2SP)
	if err != nil {
		ctrl.log.Error("Failed to get edge collection", zap.Error(err))
		return err
	}

	_, err = edge.CreateDocument(ctx, Access{
		From: driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, group),
		To:   driver.NewDocumentID(schema.SERVICES_PROVIDERS_COL, sp),
	})
	return err
}

func (ctrl *InstancesGroupsController) SetStatus(ctx context.Context, ig *pb.InstancesGroup, status pb.InstanceStatus) (err error) {
	mask := &pb.InstancesGroup{
		Status: status,
	}
	_, err = ctrl.col.UpdateDocument(ctx, ig.Uuid, mask)
	return err
}
