/*
Copyright © 2021 Nikita Ivanovski info@slnt-opp.xyz

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

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud/pkg/services/proto"
	"go.uber.org/zap"
)

const (
	SERVICES_COL = "Services"
	NS2SERV = NAMESPACES_COL + "2" + SERVICES_COL
	SP2SERV = SERVICES_PROVIDERS_COL + "2" + SERVICES_COL
)

type Service struct {
	*pb.Service
	driver.DocumentMeta
}

type Provision struct {
	From driver.DocumentID `json:"_from"`
	To driver.DocumentID `json:"_to"`
	Group string `json:"group"`

	driver.DocumentMeta
}

type ServicesController struct {
	log *zap.Logger

	col driver.Collection // Services Collection
	ig_ctrl InstancesGroupsController

	db driver.Database
}

func NewServicesController(log *zap.Logger, db driver.Database) ServicesController {
	col, _ := db.Collection(nil, SERVICES_COL)
	return ServicesController{log: log, col: col, ig_ctrl: NewInstancesGroupsController(log, db), db:db}
}

// Create Service and underlaying entities and store in DB
func (ctrl *ServicesController) Create(ctx context.Context, service *pb.Service) (*Service, error) {
	ctrl.log.Debug("Creating Service", zap.Any("service", service))
	for _, ig := range service.GetInstancesGroups() {
		err := ctrl.ig_ctrl.Create(ctx, ig)
		if err != nil {
			return nil, err
		}
	}
	service.Status = "init"
	meta, err := ctrl.col.CreateDocument(ctx, service)
	if err != nil {
		ctrl.log.Debug("Error creating document", zap.Error(err))
		return nil, errors.New("Error creating document")
	}
	service.Uuid = meta.ID.Key()
	return &Service{service, meta}, nil
}

func (ctrl *ServicesController) Update(ctx context.Context, service *pb.Service) (error) {
	ctrl.log.Debug("Updating Service", zap.Any("service", service))
	for _, ig := range service.GetInstancesGroups() {
		err := ctrl.ig_ctrl.Update(ctx, ig)
		if err != nil {
			return err
		}
	}
	meta, err := ctrl.col.UpdateDocument(ctx, service.GetUuid(), service)
	ctrl.log.Debug("UpdateDocument.Result", zap.Any("meta", meta), zap.Error(err))
	return err
}

// Get Service from DB
func (ctrl *ServicesController) Get(ctx context.Context, id string) (*Service, error) {
	ctrl.log.Debug("Getting Service", zap.String("id", id))
	var service pb.Service
	meta, err := ctrl.col.ReadDocument(ctx, id, &service)
	if err != nil {
		ctrl.log.Debug("Error reading document(Service)", zap.Error(err))
		return nil, errors.New("Error reading document")
	}
	ctrl.log.Debug("ReadDocument.Result", zap.Any("meta", meta), zap.Any("service", &service))
	service.Uuid = meta.ID.Key()
	return &Service{&service, meta}, nil
}

// List Services in DB
func (ctrl *ServicesController) List(ctx context.Context, requestor string, req_depth *int32) ([]*Service, error) {
	ctrl.log.Debug("Getting Services", zap.String("requestor", requestor))

	var depth int32
	if req_depth == nil || *req_depth < 2 {
		depth = 5
	} else {
		depth = *req_depth
	}

	query := `FOR node IN 0..@depth OUTBOUND @account GRAPH @permissions_graph OPTIONS {order: "bfs", uniqueVertices: "global"} FILTER IS_SAME_COLLECTION(@@services, node) RETURN node`
	bindVars := map[string]interface{}{
		"depth": depth,
		"account": driver.NewDocumentID(ACCOUNTS_COL, requestor),
		"permissions_graph": PERMISSIONS_GRAPH.Name,
		"@services": SERVICES_COL,
	}
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	c, err := ctrl.col.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var r []*Service
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
		r = append(r, &Service{&s, meta})
	}

	return r,  nil
}

// Join Service into Namespace
func (ctrl *ServicesController) Join(ctx context.Context, service *Service, namespace *Namespace, access int32, role string) (error) {
	ctrl.log.Debug("Joining service to namespace")
	edge, _ := ctrl.db.Collection(ctx, NS2SERV)
	_, err := edge.CreateDocument(ctx, Access{
		From: namespace.ID,
		To: service.ID,
		Level: access,
		Role: role,
	})
	return err
}

func (ctrl *ServicesController) Provide(ctx context.Context, sp, service driver.DocumentID, group string) (error) {
	ctrl.log.Debug("Providing group to service provider")
	edge, _ := ctrl.db.Collection(ctx, SP2SERV)
	_, err := edge.CreateDocument(ctx, Provision{
		From: sp,
		To: service,
		Group: group,
		DocumentMeta: driver.DocumentMeta{Key: group},
	})
	return err
}

func (ctrl *ServicesController) Unprovide(ctx context.Context, group string) (err error) {
	ctrl.log.Debug("Unproviding group from service provider")
	g, _ := ctrl.db.Graph(ctx, SERVICES_GRAPH.Name)
	edge, _, _ := g.EdgeCollection(ctx, SP2SERV)
	_, err = edge.RemoveDocument(ctx, group)
	return err
}

func (ctrl *ServicesController) GetProvisions(ctx context.Context, service string) (r map[string]string, err error) {
	ctrl.log.Debug("Getting groups provisions")
	query := `FOR service, provision IN INBOUND @service GRAPH @@services RETURN provision`
	bindVars := map[string]interface{}{
		"service": service,
		"@services": SERVICES_GRAPH.Name,
	}

	c, err := ctrl.col.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	r = make(map[string]string)
	for {
		var p Provision
		_, err = c.ReadDocument(ctx, &p)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		ctrl.log.Debug("Got document", zap.Any("provision", p))
		r[p.Group] = p.From.Key()
	}

	return r, nil
}