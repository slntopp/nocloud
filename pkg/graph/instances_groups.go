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

	pb "github.com/slntopp/nocloud/pkg/instances/proto"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

const (
	INSTANCES_GROUPS_COL = "InstancesGroups"
)

type InstancesGroupsController struct {
	col driver.Collection // Instances Collection
	graph driver.Graph

	inst_ctrl InstancesController
	
	log *zap.Logger
}

func NewInstancesGroupsController(log *zap.Logger, db driver.Database) InstancesGroupsController {
	ctx := context.TODO()

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.INSTANCES_GROUPS_COL)

	col.EnsurePersistentIndex(ctx, []string{"type"}, &driver.EnsurePersistentIndexOptions{
		Unique: false, Sparse: true, InBackground: true, Name: "sp-type",
	})

	GraphGetEdgeEnsure(log, ctx, graph, schema.SERV2IG, schema.SERVICES_COL, schema.INSTANCES_GROUPS_COL)
	GraphGetEdgeEnsure(log, ctx, graph, schema.IG2SP, schema.INSTANCES_GROUPS_COL, schema.SERVICES_PROVIDERS_COL)

	return InstancesGroupsController{
		log: log.Named("InstancesGroupsController"), inst_ctrl: NewInstancesController(log, db),
		col: col, graph: graph,
	}
}

func (ctrl *InstancesGroupsController) Create(ctx context.Context, service driver.DocumentID, g *pb.InstancesGroup) (error) {
	log := ctrl.log.Named("Create")
	log.Debug("Creating InstancesGroup", zap.Any("group", g))
	
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
		err := ctrl.inst_ctrl.Create(ctx, meta.ID, instance)
		if err != nil {
			log.Error("Failed to create Instance", zap.Error(err))
			continue
		}
	}

	return nil
}