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

	"github.com/arangodb/go-driver"
	"github.com/rabbitmq/amqp091-go"
	"github.com/slntopp/nocloud-proto/access"
	hasher "github.com/slntopp/nocloud-proto/hasher"
	spb "github.com/slntopp/nocloud-proto/services"
	stpb "github.com/slntopp/nocloud-proto/statuses"
	nocloud "github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	driver.DocumentMeta
	*spb.Service
}

type Provision struct {
	From  driver.DocumentID `json:"_from"`
	To    driver.DocumentID `json:"_to"`
	Group string            `json:"group"`

	driver.DocumentMeta
}

type ServicesController struct {
	log *zap.Logger

	col     driver.Collection // Services Collection
	ig_ctrl *InstancesGroupsController

	db driver.Database
}

func NewServicesController(log *zap.Logger, db driver.Database, conn *amqp091.Connection) ServicesController {
	log.Debug("New Services Controller creating")
	ctx := context.TODO()

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.SERVICES_COL)
	/* #nosec */
	col.EnsurePersistentIndex(ctx, []string{"status"}, &driver.EnsurePersistentIndexOptions{
		Unique: false, Sparse: true, InBackground: true, Name: "service-status",
	})
	GraphGetEdgeEnsure(log, ctx, graph, schema.NS2SERV, schema.NAMESPACES_COL, schema.SERVICES_COL)

	return ServicesController{log: log, col: col, ig_ctrl: NewInstancesGroupsController(log, db, conn), db: db}
}

func (ctrl *ServicesController) IGController() *InstancesGroupsController {
	return ctrl.ig_ctrl
}

// Create Service and underlaying entities and store in DB
func (ctrl *ServicesController) Create(ctx context.Context, service *spb.Service) (*spb.Service, error) {
	log := ctrl.log.Named("Create")
	log.Debug("Creating Service", zap.Any("service", service))

	service.Status = stpb.NoCloudStatus_INIT

	err := hasher.SetHash(service.ProtoReflect())
	if err != nil {
		return nil, err
	}

	obj := proto.Clone(service).(*spb.Service)
	obj.InstancesGroups = nil

	meta, err := ctrl.col.CreateDocument(ctx, obj)
	if err != nil {
		log.Debug("Error creating document(Service)", zap.Error(err))
		return nil, err
	}
	service.Uuid = meta.Key

	log.Debug("Creating Groups", zap.Any("method", service.GetInstancesGroups()), zap.Any("direct", service.InstancesGroups))
	for _, ig := range service.GetInstancesGroups() {
		err := ctrl.ig_ctrl.Create(ctx, meta.ID, ig)
		if err != nil {
			log.Error("Error creating InstancesGroup", zap.Error(err))
			continue
		}
	}

	return service, nil
}

// Update Service and underlaying entities and store in DB
func (ctrl *ServicesController) Update(ctx context.Context, service *spb.Service, hash bool) error {
	log := ctrl.log.Named("Update")
	log.Debug("Updating Service", zap.Any("service", service))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	oldService, err := ctrl.Get(ctx, requestor, service.GetUuid())
	if err != nil {
		log.Debug("Error recieving document(Service)", zap.Error(err))
		return err
	}

	// deleting missing groups
	for _, oldIg := range oldService.GetInstancesGroups() {
		var oldIgFound = false
		for _, ig := range service.GetInstancesGroups() {
			if oldIg.GetUuid() == ig.GetUuid() {
				oldIgFound = true
				break
			}
		}
		if !oldIgFound {
			err = ctrl.ig_ctrl.Delete(ctx, service.GetUuid(), oldIg)
			if err != nil {
				log.Error("Error while deleting instances group", zap.Error(err))
				return err
			}
		}
	}

	// creating missing and updating existing groups
	for _, ig := range service.GetInstancesGroups() {
		var igFound = false
		for _, oldIg := range oldService.GetInstancesGroups() {
			if ig.GetUuid() == oldIg.GetUuid() {
				igFound = true
				err = ctrl.ig_ctrl.Update(ctx, ig, oldIg)
				if err != nil {
					log.Error("Error while updating instances group", zap.Error(err))
					return err
				}
				break
			}
		}
		if !igFound {
			docID := driver.NewDocumentID(schema.SERVICES_COL, service.Uuid)
			err = ctrl.ig_ctrl.Create(ctx, docID, ig)
			if err != nil {
				log.Error("Error while creating instances group", zap.Error(err))
				return err
			}
			if err := ctrl.ig_ctrl.Provide(ctx, ig.GetUuid(), ig.GetSp()); err != nil {
				log.Error("Error while providing instances group", zap.Error(err))
				return err
			}
		}
	}

	err = hasher.SetHash(service.ProtoReflect())
	if err != nil {
		log.Error("Failed to calculate hash", zap.Error(err))
		return err
	}

	mask := &spb.Service{
		Uuid:    service.GetUuid(),
		Context: service.GetContext(),
	}

	if oldService.GetVersion() != service.GetVersion() {
		mask.Version = service.GetVersion()
	}
	if oldService.GetTitle() != service.GetTitle() {
		mask.Title = service.GetTitle()
	}
	if hash {
		mask.Hash = service.GetHash()
	}

	_, err = ctrl.col.UpdateDocument(ctx, mask.Uuid, mask)
	if err != nil {
		log.Debug("Error updating document(Service)", zap.Error(err))
		return err
	}

	return nil
}

// Get Service from DB
const getServiceQuery = `
LET service = (
    FOR path IN OUTBOUND K_SHORTEST_PATHS @account TO @service
    GRAPH @permissions SORT path.edges[0].level DESC
		LET perm = path.edges[0]
    	RETURN MERGE(path.vertices[-1], { access: { level: perm.level, role: perm.role, namespace: path.vertices[-2]._key }})
)[0]
LET instances_groups = (
    FOR group IN 1 OUTBOUND service
    GRAPH @permissions
        LET instances = (
            FOR i IN 1 OUTBOUND group
            GRAPH @permissions
            FILTER IS_SAME_COLLECTION(@instances, i)
				LET bp = DOCUMENT(CONCAT(@bps, "/", i.billing_plan.uuid))
                RETURN MERGE(i, { 
                    uuid: i._key, 
                    access: service.access, 
                    billing_plan: {
                        uuid: bp._key,
                        title: bp.title,
                        type: bp.type,
                        kind: bp.kind,
                        resources: bp.resources,
                        products: {
                            [i.product]: bp.products[i.product],
                        },
						meta: bp.meta,
						fee: bp.fee,
						software: bp.software
                    } 
                }))
        LET sp = (
            FOR s IN 1 OUTBOUND group
            GRAPH @permissions
            FILTER IS_SAME_COLLECTION(@sps, s)
                RETURN s._key )
        RETURN MERGE(group, { uuid: group._key, instances, sp: sp[0], access: service.access })
)

RETURN MERGE(service, { uuid: service._key, instances_groups })
`

func (ctrl *ServicesController) Get(ctx context.Context, acc, key string) (*spb.Service, error) {
	ctrl.log.Debug("Getting Service", zap.String("key", key))

	requestor := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
	id := driver.NewDocumentID(schema.SERVICES_COL, key)
	c, err := ctrl.db.Query(ctx, getServiceQuery, map[string]interface{}{
		"account":     requestor,
		"service":     id,
		"instances":   schema.INSTANCES_COL,
		"sps":         schema.SERVICES_PROVIDERS_COL,
		"bps":         schema.BILLING_PLANS_COL,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
	})
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var service spb.Service

	_, err = c.ReadDocument(ctx, &service)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

const getServiceInstancesUuidQuery = `
FOR group IN 1 OUTBOUND @service GRAPH @permissions
    FOR i IN 1 OUTBOUND group
    GRAPH @permissions
    FILTER IS_SAME_COLLECTION(@instances, i)
    	RETURN i._key
`

func (ctrl *ServicesController) GetServiceInstancesUuids(key string) ([]string, error) {
	ctrl.log.Debug("Getting Service", zap.String("key", key))

	id := driver.NewDocumentID(schema.SERVICES_COL, key)
	ctx := context.Background()
	c, err := ctrl.db.Query(ctx, getServiceInstancesUuidQuery, map[string]interface{}{
		"service":     id,
		"instances":   schema.INSTANCES_COL,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
	})
	if err != nil {
		return nil, err
	}
	defer c.Close()

	uuids := make([]string, 0)
	for {
		var uuid string
		_, err = c.ReadDocument(ctx, &uuid)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			ctrl.log.Error("Failed to recieve Service Instances UUIDs", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to recieve Service Instances UUIDs")
		} else {
			uuids = append(uuids, uuid)
		}
	}

	return uuids, nil
}

var getServiceList = `
LET list = (FOR service, e, path IN 0..@depth OUTBOUND @account
    GRAPH @permissions_graph
	OPTIONS {order: "bfs", uniqueVertices: "global"}
    FILTER IS_SAME_COLLECTION(@@services, service)
        LET perm = path.edges[0]
		%s
		LET instances_groups = (
    	FOR group IN 1 OUTBOUND service
    	GRAPH @permissions_graph
    		LET instances = (
    			FOR i IN 1 OUTBOUND group
    			GRAPH @permissions_graph
    			FILTER IS_SAME_COLLECTION(@instances, i)
				%s
					LET bp = DOCUMENT(CONCAT(@bps, "/", i.billing_plan.uuid))
					RETURN MERGE(i, { 
						uuid: i._key, 
						billing_plan: {
							uuid: bp._key,
							title: bp.title,
							type: bp.type,
							kind: bp.kind,
							resources: bp.resources,
							products: {
								[i.product]: bp.products[i.product],
							},
							meta: bp.meta,
							fee: bp.fee,
							software: bp.software
						} 
					}))
    		RETURN MERGE(group, { uuid: group._key, instances })
        )
    RETURN MERGE(service, {uuid:service._key, instances_groups, access: { level: perm.level, role: perm.role, namespace: path.vertices[-2]._key }})
	)

RETURN { 
	result: (@limit > 0) ? SLICE(list, @offset, @limit) : list,
	count: LENGTH(list)
}
`

type ServicesResult struct {
	Result []*spb.Service `json:"result"`
	Count  int            `json:"count"`
}

// List Services in DB
func (ctrl *ServicesController) List(ctx context.Context, requestor string, request *spb.ListRequest) (*ServicesResult, error) {
	ctrl.log.Debug("Getting Services", zap.String("requestor", requestor))

	depth := request.GetDepth()
	if depth < 2 {
		depth = 5
	}
	showDeleted := request.GetShowDeleted()

	limit, page := request.GetLimit(), request.GetPage()
	offset := (page - 1) * limit

	var query, instQuery string
	bindVars := map[string]interface{}{
		"depth":             depth,
		"account":           driver.NewDocumentID(schema.ACCOUNTS_COL, requestor),
		"permissions_graph": schema.PERMISSIONS_GRAPH.Name,
		"@services":         schema.SERVICES_COL,
		"instances":         schema.INSTANCES_COL,
		"bps":               schema.BILLING_PLANS_COL,
		"offset":            offset,
		"limit":             limit,
	}

	if request.Field != nil && request.Sort != nil {
		subQuery := ` SORT service.%s %s`
		field, sort := request.GetField(), request.GetSort()

		query += fmt.Sprintf(subQuery, field, sort)
	}

	for key, val := range request.GetFilters() {
		if key == "search_param" {
			query += fmt.Sprintf(` FILTER LOWER(service.title) LIKE LOWER("%s")`, "%"+val.GetStringValue()+"%")
		} else if key == "access.level" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += ` FILTER perm.level in @levels`
			bindVars["levels"] = values
		} else if key == "account" {
			value := val.GetStringValue()
			query += ` FILTER path.vertices[-3]._key == @search_account`
			bindVars["search_account"] = value
		} else if key == "namespace" {
			value := val.GetStringValue()
			query += ` FILTER path.vertices[-2]._key == @search_namespace`
			bindVars["search_namespace"] = value
		} else {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER service["%s"] in @%s`, key, key)
			bindVars[key] = values
		}
	}

	if showDeleted {
		instQuery = ""
	} else {
		query += fmt.Sprintf(` FILTER service.status != %d`, stpb.NoCloudStatus_DEL)
		instQuery = fmt.Sprintf(` FILTER i.status != %d`, stpb.NoCloudStatus_DEL)
	}

	query = fmt.Sprintf(getServiceList, query, instQuery)

	ctrl.log.Debug("Query", zap.Any("q", query))
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars), zap.Bool("show_deleted", showDeleted))

	c, err := ctrl.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var result ServicesResult
	_, err = c.ReadDocument(ctx, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Join Service into Namespace
func (ctrl *ServicesController) Join(ctx context.Context, service *spb.Service, ns *Namespace, access access.Level, role string) error {
	ctrl.log.Debug("Joining service to namespace")
	edge, _ := ctrl.db.Collection(ctx, schema.NS2SERV)
	_, err := edge.CreateDocument(ctx, Access{
		From:  ns.ID,
		To:    driver.NewDocumentID(schema.SERVICES_COL, service.Uuid),
		Level: access,
		Role:  role,
	})
	return err
}

func (ctrl *ServicesController) Delete(ctx context.Context, s *spb.Service) (err error) {
	log := ctrl.log.Named("Service.Delete")
	log.Debug("Deleting Service", zap.String("status", s.GetStatus().String()))
	if s.GetStatus() != stpb.NoCloudStatus_INIT && s.GetStatus() != stpb.NoCloudStatus_DOWN {
		return fmt.Errorf("cannot delete Service, status: %s", s.GetStatus())
	}

	return ctrl.SetStatus(ctx, s, stpb.NoCloudStatus_DEL)
}

func (ctrl *ServicesController) SetStatus(ctx context.Context, s *spb.Service, status stpb.NoCloudStatus) (err error) {
	mask := &spb.Service{
		Status: status,
	}
	_, err = ctrl.col.UpdateDocument(ctx, s.Uuid, mask)
	return err
}
