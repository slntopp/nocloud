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

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"

	"github.com/slntopp/nocloud/pkg/hasher"
	pb "github.com/slntopp/nocloud/pkg/instances/proto"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

const (
	INSTANCES_COL = "Instances"
)

type InstancesController struct {
	col driver.Collection // Instances Collection
	graph driver.Graph

	log *zap.Logger
}

func NewInstancesController(log *zap.Logger, db driver.Database) InstancesController {
	ctx := context.TODO()

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.INSTANCES_COL)
	GraphGetEdgeEnsure(log, ctx, graph, schema.IG2INST, schema.INSTANCES_GROUPS_COL, schema.INSTANCES_COL)

	return InstancesController{log: log.Named("InstancesController"), col: col, graph: graph}
}

func (ctrl *InstancesController) Create(ctx context.Context, group driver.DocumentID, i *pb.Instance) error {
	log := ctrl.log.Named("Create")
	log.Debug("Creating Instance", zap.Any("instance", i))

	// ensure status is INIT
	i.Status = pb.InstanceStatus_INIT

	err := hasher.SetHash(i.ProtoReflect())
	if err != nil {
		log.Error("Failed to calculate hash", zap.Error(err))
		return err
	}

	// Attempt create document
	meta, err := ctrl.col.CreateDocument(ctx, i)
	if err != nil {
		log.Error("Failed to create Instance", zap.Error(err))
		return err
	}
	i.Uuid = meta.Key

	// Attempt get edge collection
	edge, _, err := ctrl.graph.EdgeCollection(ctx, schema.IG2INST)
	if err != nil {
		log.Error("Failed to get EdgeCollection", zap.Error(err))
		ctrl.col.RemoveDocument(ctx, meta.Key) // if failed - remove instance from DataBase
		return err
	}

	// Attempt create edge
	_, err = edge.CreateDocument(ctx, Access{
		From: group, To: meta.ID,
	})
	if err != nil {
		log.Error("Failed to create Edge", zap.Error(err))
		ctrl.col.RemoveDocument(ctx, meta.Key) // if failed - remove instance from DataBase
		return err
	}

	return nil
}