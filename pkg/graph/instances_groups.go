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
	"fmt"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"reflect"

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/slntopp/nocloud-proto/hasher"
	pb "github.com/slntopp/nocloud-proto/instances"
	spb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

type InstancesGroupsController interface {
	Instances() InstancesController
	Create(ctx context.Context, service driver.DocumentID, g *pb.InstancesGroup) error
	GetWithAccess(ctx context.Context, from driver.DocumentID, id string) (InstancesGroup, error)
	Delete(ctx context.Context, service string, g *pb.InstancesGroup) error
	Update(ctx context.Context, ig, oldIg *pb.InstancesGroup) error
	TransferIG(ctx context.Context, oldSrvEdge string, newSrv driver.DocumentID, ig driver.DocumentID) error
	Provide(ctx context.Context, group, sp string) error
	SetStatus(ctx context.Context, ig *pb.InstancesGroup, status spb.NoCloudStatus) (err error)
	GetEdge(ctx context.Context, inboundNode string, collection string) (string, error)
	GetSP(ctx context.Context, group string) (ServicesProvider, error)
}

type InstancesGroup struct {
	driver.DocumentMeta
	*pb.InstancesGroup
}

type instancesGroupsController struct {
	col   driver.Collection // Instances Groups Collection
	graph driver.Graph

	inst_ctrl InstancesController

	log *zap.Logger

	db driver.Database

	serv2ig driver.Collection
	ig2sp   driver.Collection
}

func NewInstancesGroupsController(log *zap.Logger, db driver.Database, conn rabbitmq.Connection) InstancesGroupsController {
	log.Debug("New InstancesGroups Controller Creating")

	ctx := context.TODO()

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.INSTANCES_GROUPS_COL)

	/* #nosec */
	col.EnsurePersistentIndex(ctx, []string{"type"}, &driver.EnsurePersistentIndexOptions{
		Unique: false, Sparse: true, InBackground: true, Name: "sp-type",
	})

	serv2ig := GraphGetEdgeEnsure(log, ctx, graph, schema.SERV2IG, schema.SERVICES_COL, schema.INSTANCES_GROUPS_COL)
	ig2sp := GraphGetEdgeEnsure(log, ctx, graph, schema.IG2SP, schema.INSTANCES_GROUPS_COL, schema.SERVICES_PROVIDERS_COL)
	inst := NewInstancesController(log, db, conn)

	return &instancesGroupsController{
		log: log.Named("InstancesGroupsController"), inst_ctrl: inst,
		col: col, graph: graph, db: db,
		serv2ig: serv2ig,
		ig2sp:   ig2sp,
	}
}

func (ctrl *instancesGroupsController) Instances() InstancesController {
	return ctrl.inst_ctrl
}

func (ctrl *instancesGroupsController) Create(ctx context.Context, service driver.DocumentID, g *pb.InstancesGroup) error {
	log := ctrl.log.Named("Create")
	log.Debug("Creating InstancesGroup", zap.Any("group", g))

	g.Status = spb.NoCloudStatus_INIT

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

	_, err = ctrl.serv2ig.CreateDocument(ctx, Access{
		From: service, To: meta.ID,
		Role: roles.OWNER,
	})
	if err != nil {
		log.Error("Failed to create Edge", zap.Error(err))
		if _, err = ctrl.col.RemoveDocument(ctx, meta.Key); err != nil {
			log.Warn("Failed to cleanup", zap.String("uuid", meta.Key), zap.Error(err))
		}
		return err
	}

	isImported := g.GetData()["imported"].GetBoolValue()

	for _, instance := range g.GetInstances() {
		if isImported {
			log.Error("Can't create new instance to imported IG")
			return fmt.Errorf("can't create new instance to imported IG")
		}
		_, err := ctrl.inst_ctrl.Create(ctx, meta.ID, *g.Sp, instance)
		if err != nil {
			log.Error("Failed to create Instance", zap.Error(err))
			return err
		}
	}

	return nil
}

func (ctrl *instancesGroupsController) GetWithAccess(ctx context.Context, from driver.DocumentID, id string) (InstancesGroup, error) {
	return getWithAccess[InstancesGroup](ctx, ctrl.col.Database(), from, driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, id))
}

func (ctrl *instancesGroupsController) Delete(ctx context.Context, service string, g *pb.InstancesGroup) error {
	log := ctrl.log.Named("Delete")
	log.Debug("Deleting InstancesGroup", zap.Any("group", g))

	_, err := ctrl.col.RemoveDocument(ctx, g.GetUuid())
	if err != nil {
		log.Error("Failed to delete InstancesGroup", zap.Error(err))
		return err
	}

	ctrl.log.Debug("Deleting Edge", zap.String("fromCollection", schema.SERVICES_COL), zap.String("toCollection",
		schema.INSTANCES_GROUPS_COL), zap.String("fromKey", service), zap.String("toKey", g.GetUuid()))
	err = deleteEdge(ctx, ctrl.col.Database(), schema.SERVICES_COL, schema.INSTANCES_GROUPS_COL, service, g.GetUuid())
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

func (ctrl *instancesGroupsController) Update(ctx context.Context, ig, oldIg *pb.InstancesGroup) error {
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
		if !oldInstFound && oldInst.GetStatus() != spb.NoCloudStatus_DEL {
			err := ctrl.inst_ctrl.Delete(ctx, ig.GetUuid(), oldInst)
			if err != nil {
				log.Error("Error while deleting instance", zap.Error(err))
				return err
			}
		}
	}

	isImported := ig.GetData()["imported"].GetBoolValue()

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
			if isImported {
				log.Error("Can't create new instance to imported IG")
				return fmt.Errorf("can't create new instance to imported IG")
			}

			docID := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, ig.Uuid)
			_, err := ctrl.inst_ctrl.Create(ctx, docID, *oldIg.Sp, inst)
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
	if reflect.DeepEqual(ig.GetData(), oldIg.GetData()) {
		mask.Data = ig.GetData()
	}

	_, err = ctrl.col.UpdateDocument(ctx, mask.Uuid, mask)
	if err != nil {
		log.Debug("Error updating document(InstancesGroup)", zap.Error(err))
		return err
	}

	return nil
}

func (ctrl *instancesGroupsController) TransferIG(ctx context.Context, oldSrvEdge string, newSrv driver.DocumentID, ig driver.DocumentID) error {
	log := ctrl.log.Named("Transfer")
	log.Debug("Transfer InstancesGroup", zap.String("group", ig.String()), zap.String("srvEdge", oldSrvEdge), zap.String("to", newSrv.String()))

	_, err := ctrl.serv2ig.RemoveDocument(ctx, oldSrvEdge)
	if err != nil {
		log.Error("Failed to remove old Edge", zap.Error(err))
		return err
	}

	_, err = ctrl.serv2ig.CreateDocument(ctx, Access{From: newSrv, To: ig, Role: roles.OWNER})
	if err != nil {
		log.Error("Failed to create Edge", zap.Error(err))
		return err
	}

	return nil
}

const getSPQuery = `
LET instanceGroup = DOCUMENT(@instanceGroup)
LET sp = (
    FOR sp IN 1 OUTBOUND instanceGroup
    FILTER IS_SAME_COLLECTION(@sps, sp)
    GRAPH @permissions
        RETURN sp )[0]
        
RETURN MERGE(sp, {uuid: sp._key})`

func (ctrl *instancesGroupsController) GetSP(ctx context.Context, group string) (ServicesProvider, error) {
	log := ctrl.log.Named("GetSP")
	log.Debug("Getting Service Provider connected with IG", zap.String("group", group))
	c, err := ctrl.db.Query(ctx, getSPQuery, map[string]interface{}{
		"permissions":   schema.PERMISSIONS_GRAPH.Name,
		"sps":           schema.SERVICES_PROVIDERS_COL,
		"instanceGroup": driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, group),
	})
	if err != nil {
		log.Error("Error while querying", zap.Error(err))
		return ServicesProvider{}, err
	}
	defer c.Close()

	var r ServicesProvider
	meta, err := c.ReadDocument(ctx, &r)
	if err != nil {
		log.Error("Error while reading document", zap.Error(err))
		return r, err
	}
	r.Uuid = meta.Key

	return r, nil
}

func (ctrl *instancesGroupsController) Provide(ctx context.Context, group, sp string) error {
	_, err := ctrl.ig2sp.CreateDocument(ctx, Access{
		From: driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, group),
		To:   driver.NewDocumentID(schema.SERVICES_PROVIDERS_COL, sp),
	})
	if err != nil {
		return err
	}
	_, err = ctrl.col.UpdateDocument(ctx, group, map[string]interface{}{"sp": sp})
	return err
}

func (ctrl *instancesGroupsController) SetStatus(ctx context.Context, ig *pb.InstancesGroup, status spb.NoCloudStatus) (err error) {
	mask := &pb.InstancesGroup{
		Status: status,
	}
	_, err = ctrl.col.UpdateDocument(ctx, ig.Uuid, mask)
	return err
}

func (ctrl *instancesGroupsController) GetEdge(ctx context.Context, inboundNode string, collection string) (string, error) {
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
