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

	stpb "github.com/slntopp/nocloud-proto/statuses"

	"github.com/arangodb/go-driver"
	ipb "github.com/slntopp/nocloud-proto/instances"
	spb "github.com/slntopp/nocloud-proto/services"
	pb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type ServicesProvider struct {
	*pb.ServicesProvider
	driver.DocumentMeta
}

type ServicesProvidersController struct {
	col   driver.Collection // Services Providers Collection
	ig2sp driver.Collection
	sp2bp driver.Collection

	log *zap.Logger

	graph driver.Graph
}

func NewServicesProvidersController(logger *zap.Logger, db driver.Database) ServicesProvidersController {
	ctx := context.TODO()
	log := logger.Named("ServicesProvidersController")
	log.Debug("New ServicesProvider Controller Creating")

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.SERVICES_PROVIDERS_COL)

	ig2sp := GraphGetEdgeEnsure(log, ctx, graph, schema.IG2SP, schema.INSTANCES_GROUPS_COL, schema.SERVICES_PROVIDERS_COL)
	sp2pb := GraphGetEdgeEnsure(log, ctx, graph, schema.SP2BP, schema.SERVICES_PROVIDERS_COL, schema.BILLING_PLANS_COL)

	return ServicesProvidersController{log: log, col: col, graph: graph, ig2sp: ig2sp, sp2bp: sp2pb}
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

const deleteEdgesQuery = `
FOR edge IN @@collection
    FILTER edge._from == @id || edge._to == @id
    REMOVE edge._key IN @@collection
`

func (ctrl *ServicesProvidersController) DeleteEdges(ctx context.Context, id string) error {
	bindVars := map[string]interface{}{
		"@collection": schema.SP2BP,
		"id":          id,
	}
	c, err := ctrl.col.Database().Query(ctx, deleteEdgesQuery, bindVars)
	if err != nil {
		return err
	}
	c.Close()

	bindVars["@collection"] = schema.IG2SP
	c, err = ctrl.col.Database().Query(ctx, deleteEdgesQuery, bindVars)
	if err != nil {
		return err
	}
	c.Close()

	return nil
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
func (ctrl *ServicesProvidersController) List(ctx context.Context, requestor string, isRoot bool) ([]*ServicesProvider, error) {
	ctrl.log.Debug("Getting Services", zap.String("requestor", requestor))

	var query string

	if requestor != "" {
		query = `FOR sp IN @@sps %s RETURN MERGE(UNSET(sp, ['secrets', 'vars']), {uuid: sp._key})`

		if !isRoot {
			query = fmt.Sprintf(query, "FILTER sp.public == true")
		} else {
			query = fmt.Sprintf(query, "")
		}
	} else {
		// anonymous query
		query = `FOR sp IN @@sps FILTER sp.public == true RETURN {uuid: sp._key, type: sp.type, title: sp.title, public_data: sp.public_data, locations: sp.locations, meta: sp.meta}`
	}
	bindVars := map[string]interface{}{
		"@sps": schema.SERVICES_PROVIDERS_COL,
	}
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars), zap.String("query", query))

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

func (ctrl *ServicesProvidersController) BindPlan(ctx context.Context, uuid, planUuid string) error {
	ctrl.log.Debug("Binding Plan")

	exist, err := EdgeExist(ctx, ctrl.col.Database(), schema.SERVICES_PROVIDERS_COL, schema.BILLING_PLANS_COL, uuid, planUuid)

	if err != nil {
		return err
	}

	if exist {
		ctrl.log.Debug("Plan Already Binded")
		return nil
	}

	// Attempt get edge collection
	edge, err := ctrl.col.Database().Collection(ctx, schema.SP2BP)
	if err != nil {
		ctrl.log.Error("Failed to get EdgeCollection", zap.Error(err))
		return err
	}

	// Attempt create edge
	spDocId := driver.NewDocumentID(schema.SERVICES_PROVIDERS_COL, uuid)
	planDocId := driver.NewDocumentID(schema.BILLING_PLANS_COL, planUuid)
	_, err = edge.CreateDocument(ctx, Access{
		From: spDocId, To: planDocId,
	})
	if err != nil {
		ctrl.log.Error("Failed to create Edge", zap.Error(err))
		return err
	}

	return nil
}

const listDeployedServicesQuery = `
FOR srv IN 2 INBOUND @sp
GRAPH @permissions
OPTIONS { order: "bfs", uniqueVertices: "global" }
FILTER IS_SAME_COLLECTION(@services, srv)
FILTER srv.status != @status_init && srv.status != @status_del
    RETURN MERGE(srv, { uuid: srv._key })
`

const listDeployedGroupsQueryWithInstances = `
FOR group IN 1 INBOUND @sp
GRAPH @permissions
OPTIONS { order: "bfs", uniqueVertices: "global" }
FILTER group.status == @up_status || group.status == @suspend_status 
FILTER IS_SAME_COLLECTION(@groups, group)
    LET instances = (
        FOR instance IN OUTBOUND group
        GRAPH @permissions
        FILTER IS_SAME_COLLECTION(@instances, instance)
			LET bp = DOCUMENT(CONCAT(@bps, "/", instance.billing_plan.uuid))
			RETURN MERGE(instance, { 
				uuid: instance._key, 
				billing_plan: {
					uuid: bp.uuid,
					title: bp.title,
					type: bp.type,
					kind: bp.kind,
					resources: bp.resources,
					products: {
						[instance.product]: bp.products[instance.product],
					},
					meta: bp.meta,
					fee: bp.fee,
					software: bp.software
				} 
			}))
    RETURN MERGE(group, { uuid: group._key, instances })`

func (ctrl *ServicesProvidersController) GetGroups(ctx context.Context, sp *ServicesProvider) ([]*ipb.InstancesGroup, error) {
	bindVars := map[string]interface{}{
		"groups":         schema.INSTANCES_GROUPS_COL,
		"bps":            schema.BILLING_PLANS_COL,
		"sp":             sp.DocumentMeta.ID,
		"permissions":    schema.PERMISSIONS_GRAPH.Name,
		"instances":      schema.INSTANCES_COL,
		"up_status":      stpb.NoCloudStatus_UP,
		"suspend_status": stpb.NoCloudStatus_SUS,
	}
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	query := listDeployedGroupsQueryWithInstances

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
		ctrl.log.Debug("Got document", zap.Any("group", &ig))
		r = append(r, &ig)
	}

	return r, nil
}

func (ctrl *ServicesProvidersController) GetServices(ctx context.Context, sp *ServicesProvider) ([]*spb.Service, error) {
	bindVars := map[string]interface{}{
		"services":    schema.SERVICES_COL,
		"sp":          sp.DocumentMeta.ID,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"status_init": stpb.NoCloudStatus_INIT,
		"status_del":  stpb.NoCloudStatus_DEL,
	}
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	query := listDeployedServicesQuery

	c, err := ctrl.col.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var r []*spb.Service
	for {
		var s spb.Service
		_, err := c.ReadDocument(ctx, &s)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		ctrl.log.Debug("Got document", zap.Any("group", &s))
		r = append(r, &s)
	}

	return r, nil
}
