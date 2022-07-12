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
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/hasher"
	nocloud "github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	pb "github.com/slntopp/nocloud/pkg/services/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

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

func NewServicesController(log *zap.Logger, db driver.Database) ServicesController {
	log.Debug("New Services Controller creating")
	ctx := context.TODO()

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.SERVICES_COL)
	col.EnsurePersistentIndex(ctx, []string{"status"}, &driver.EnsurePersistentIndexOptions{
		Unique: false, Sparse: true, InBackground: true, Name: "service-status",
	})
	GraphGetEdgeEnsure(log, ctx, graph, schema.NS2SERV, schema.NAMESPACES_COL, schema.SERVICES_COL)

	return ServicesController{log: log, col: col, ig_ctrl: NewInstancesGroupsController(log, db), db: db}
}

func (ctrl *ServicesController) IGController() *InstancesGroupsController {
	return ctrl.ig_ctrl
}

// Create Service and underlaying entities and store in DB
func (ctrl *ServicesController) Create(ctx context.Context, service *pb.Service) (*pb.Service, error) {
	log := ctrl.log.Named("Create")
	log.Debug("Creating Service", zap.Any("service", service))

	service.Status = pb.ServiceStatus_INIT

	err := hasher.SetHash(service.ProtoReflect())
	if err != nil {
		return nil, err
	}

	obj := proto.Clone(service).(*pb.Service)
	obj.InstancesGroups = nil

	meta, err := ctrl.col.CreateDocument(ctx, obj)
	if err != nil {
		log.Debug("Error creating document(Service)", zap.Error(err))
		return nil, err
	}
	service.Uuid = meta.Key

	log.Debug("Groups", zap.Any("method", service.GetInstancesGroups()), zap.Any("direct", service.InstancesGroups))
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
func (ctrl *ServicesController) Update(ctx context.Context, service *pb.Service, hash bool) error {
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
		}
	}

	err = hasher.SetHash(service.ProtoReflect())
	if err != nil {
		log.Error("Failed to calculate hash", zap.Error(err))
		return err
	}

	mask := &pb.Service{
		Uuid:    service.GetUuid(),
		Context: service.GetContext(),
	}

	if oldService.GetVersion() != service.GetVersion() {
		mask.Version = service.GetVersion()
	}
	if oldService.GetTitle() != service.GetTitle() {
		mask.Title = service.GetTitle()
	}
	if oldService.GetStatus() != service.GetStatus() {
		mask.Status = service.GetStatus()
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
    GRAPH @permissions SORT path.edges[0].level
    	RETURN MERGE(path.vertices[-1], {
    	    access_level: path.edges[0].level ? : 0
    	})
)[0]

LET instances_groups = (
    FOR group IN 1 OUTBOUND service
    GRAPH @permissions
        LET instances = (
            FOR i IN 1 OUTBOUND group
            GRAPH @permissions
            FILTER IS_SAME_COLLECTION(@instances, i)
                RETURN MERGE(i, { uuid: i._key }) )
        LET sp = (
            FOR s IN 1 OUTBOUND group
            GRAPH @permissions
            FILTER IS_SAME_COLLECTION(@sps, s)
                RETURN s._key )
        RETURN MERGE(group, { uuid: group._key, instances, sp: sp[0] })
)

RETURN MERGE(service, { uuid: service._key, instances_groups })
`

func (ctrl *ServicesController) Get(ctx context.Context, acc, key string) (*pb.Service, error) {
	ctrl.log.Debug("Getting Service", zap.String("key", key))

	requestor := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
	id := driver.NewDocumentID(schema.SERVICES_COL, key)
	c, err := ctrl.db.Query(ctx, getServiceQuery, map[string]interface{}{
		"account":     requestor,
		"service":     id,
		"instances":   schema.INSTANCES_COL,
		"sps":         schema.SERVICES_PROVIDERS_COL,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
	})
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var service pb.Service

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
    	RETURN i.uuid
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

// List Services in DB
func (ctrl *ServicesController) List(ctx context.Context, requestor string, request *pb.ListRequest) ([]*pb.Service, error) {
	ctrl.log.Debug("Getting Services", zap.String("requestor", requestor))

	depth := request.GetDepth()
	if depth < 2 {
		depth = 5
	}
	showDeleted := request.GetShowDeleted() == "true"
	var query string
	if showDeleted {
		query = `FOR node IN 0..@depth OUTBOUND @account GRAPH @permissions_graph OPTIONS {order: "bfs", uniqueVertices: "global"} FILTER IS_SAME_COLLECTION(@@services, node) RETURN node`
	} else {
		query = `FOR node IN 0..@depth OUTBOUND @account GRAPH @permissions_graph OPTIONS {order: "bfs", uniqueVertices: "global"} FILTER IS_SAME_COLLECTION(@@services, node) FILTER node.status != "del" RETURN node`
	}
	bindVars := map[string]interface{}{
		"depth":             depth,
		"account":           driver.NewDocumentID(schema.ACCOUNTS_COL, requestor),
		"permissions_graph": schema.PERMISSIONS_GRAPH.Name,
		"@services":         schema.SERVICES_COL,
	}
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars), zap.Bool("show_deleted", showDeleted))

	c, err := ctrl.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var r []*pb.Service
	for {
		var s pb.Service
		meta, err := c.ReadDocument(ctx, &s)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		ctrl.log.Debug("Got document", zap.Any("service", &s))
		s.Uuid = meta.ID.Key()
		r = append(r, &s)
	}

	return r, nil
}

// Join Service into Namespace
func (ctrl *ServicesController) Join(ctx context.Context, service *pb.Service, namespace *Namespace, access int32, role string) error {
	ctrl.log.Debug("Joining service to namespace")
	edge, _ := ctrl.db.Collection(ctx, schema.NS2SERV)
	_, err := edge.CreateDocument(ctx, Access{
		From:  namespace.ID,
		To:    driver.NewDocumentID(schema.SERVICES_COL, service.Uuid),
		Level: access,
		Role:  role,
	})
	return err
}

func (ctrl *ServicesController) Delete(ctx context.Context, s *pb.Service) (err error) {
	log := ctrl.log.Named("Service.Delete")
	log.Debug("Deleting Service", zap.String("status", s.GetStatus().String()))
	if s.GetStatus() != pb.ServiceStatus_INIT && s.GetStatus() != pb.ServiceStatus_DOWN {
		return fmt.Errorf("cannot delete Service, status: %s", s.GetStatus())
	}

	return ctrl.SetStatus(ctx, s, pb.ServiceStatus_DEL)
}

func (ctrl *ServicesController) SetStatus(ctx context.Context, s *pb.Service, status pb.ServiceStatus) (err error) {
	mask := &pb.Service{
		Status: status,
	}
	_, err = ctrl.col.UpdateDocument(ctx, s.Uuid, mask)
	return err
}
