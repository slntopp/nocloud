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

	ipb "github.com/slntopp/nocloud/pkg/instances/proto"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	pb "github.com/slntopp/nocloud/pkg/services_providers/proto"
)

type ServicesProvider struct {
	*pb.ServicesProvider
	driver.DocumentMeta
}

type ServicesProvidersController struct {
	col driver.Collection // Services Providers Collection

	log *zap.Logger
}

func NewServicesProvidersController(logger *zap.Logger, db driver.Database) ServicesProvidersController {
	ctx := context.TODO()
	log := logger.Named("ServicesProvidersController")

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.SERVICES_PROVIDERS_COL)

	GraphGetEdgeEnsure(log, ctx, graph, schema.IG2SP, schema.INSTANCES_GROUPS_COL, schema.SERVICES_PROVIDERS_COL)

	return ServicesProvidersController{log: log, col: col}
}

func (ctrl *ServicesProvidersController) Create(ctx context.Context, sp *ServicesProvider) (err error) {
	ctrl.log.Debug("Creating Document for ServicesProvider", zap.Any("config", sp))
	meta, err := ctrl.col.CreateDocument(ctx, *sp)
	sp.Uuid = meta.Key
	return err
}

// Update ServicesProvider in DB
func (ctrl *ServicesProvidersController) Update(ctx context.Context, sp *pb.ServicesProvider) error {
	ctrl.log.Debug("Updating ServicesProvider", zap.Any("sp", sp))

	meta, err := ctrl.col.ReplaceDocument(ctx, sp.GetUuid(), sp)
	ctrl.log.Debug("ReplaceDocument.Result", zap.Any("meta", meta), zap.Error(err))
	return err
}

func (ctrl *ServicesProvidersController) Delete(ctx context.Context, id string) (err error) {
	ctrl.log.Debug("Deleting ServicesProvider Document", zap.Any("uuid", id))
	_, err = ctrl.col.RemoveDocument(ctx, id)
	return err
}

func (ctrl *ServicesProvidersController) Get(ctx context.Context, id string) (r *ServicesProvider, err error) {
	ctrl.log.Debug("Getting ServicesProvider", zap.Any("sp", id))
	var sp pb.ServicesProvider
	query := `RETURN DOCUMENT(@sp)`
	c, err := ctrl.col.Database().Query(ctx, query, map[string]interface{}{
		"sp": driver.NewDocumentID(schema.SERVICES_PROVIDERS_COL, id),
	})
	if err != nil {
		ctrl.log.Debug("Error reading document(ServiceProvider)", zap.Error(err))
		return nil, err
	}
	defer c.Close()

	meta, err := c.ReadDocument(ctx, &sp)
	ctrl.log.Debug("ReadDocument.Result", zap.Any("meta", meta), zap.Error(err), zap.Any("sp", &sp))
	sp.Uuid = meta.ID.Key()
	return &ServicesProvider{&sp, meta}, err
}

// List Services Providers in DB
func (ctrl *ServicesProvidersController) List(ctx context.Context, requestor string) ([]*ServicesProvider, error) {
	ctrl.log.Debug("Getting Services", zap.String("requestor", requestor))

	var query string

	if requestor != "" {
		query = `FOR sp IN @@sps RETURN sp`
	} else {
		// anonymous query
		query = `FOR sp IN @@sps RETURN {uuid: sp.uuid, type: sp.type, title: sp.title, public_data: sp.public_data}`
	}
	bindVars := map[string]interface{}{
		"@sps": schema.SERVICES_PROVIDERS_COL,
	}
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	c, err := ctrl.col.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var r []*ServicesProvider
	for {
		var s pb.ServicesProvider
		meta, err := c.ReadDocument(ctx, &s)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		ctrl.log.Debug("Got document", zap.Any("service_provider", &s))
		r = append(r, &ServicesProvider{&s, meta})
	}

	return r, nil
}

const listDeployedGroupsQuery = `
FOR group, edge, path IN 1 INBOUND @sp
GRAPH @permissions
OPTIONS { order: "bfs", uniqueVertices: "global" }
FILTER IS_SAME_COLLECTION(@groups, group)
    RETURN MERGE(group, { uuid: group._key })`
const listDeployedGroupsQueryWithInstances = `
FOR group IN 1 INBOUND @sp
GRAPH @permissions
OPTIONS { order: "bfs", uniqueVertices: "global" }
FILTER IS_SAME_COLLECTION(@groups, group)
    LET instances = (
        FOR instance IN OUTBOUND group
        GRAPH @permissions
        FILTER IS_SAME_COLLECTION(@instances, instance)
            RETURN MERGE(instance, { uuid: instance._key }) )
    RETURN MERGE(group, { uuid: group._key, instances })`

func (ctrl *ServicesProvidersController) ListDeployments(ctx context.Context, sp *ServicesProvider, includeInstances bool) ([]*ipb.InstancesGroup, error) {
	bindVars := map[string]interface{}{
		"groups":      schema.INSTANCES_GROUPS_COL,
		"sp":          sp.DocumentMeta.ID,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"instances":   schema.INSTANCES_COL,
	}
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	query := listDeployedGroupsQuery
	if includeInstances {
		query = listDeployedGroupsQueryWithInstances
	}
	c, err := ctrl.col.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var r []*ipb.InstancesGroup
	for {
		var ig ipb.InstancesGroup
		_, err := c.ReadDocument(ctx, &ig)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		ctrl.log.Debug("Got document", zap.Any("group", ig))
		r = append(r, &ig)
	}

	return r, nil
}
