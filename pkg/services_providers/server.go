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
package services_providers

import (
	"context"
	"fmt"
	stpb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"time"

	elpb "github.com/slntopp/nocloud-proto/events_logging"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	proto "github.com/slntopp/nocloud-proto/states"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	p "github.com/slntopp/nocloud/pkg/public_data"
	s "github.com/slntopp/nocloud/pkg/states"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServicesProviderServer struct {
	sppb.UnimplementedServicesProvidersServiceServer

	drivers map[string]driverpb.DriverServiceClient

	extention_servers map[string]sppb.ServicesProvidersExtentionsServiceClient
	db                driver.Database
	ctrl              graph.ServicesProvidersController
	ns_ctrl           graph.NamespacesController
	ca                graph.CommonActionsController

	log *zap.Logger

	rdb redisdb.Client
}

func NewServicesProviderServer(log *zap.Logger, db driver.Database, rbmq rabbitmq.Connection, rdb redisdb.Client) *ServicesProviderServer {
	s := s.NewStatesPubSub(log, &db, rbmq)
	p := p.NewPublicDataPubSub(log, &db, rbmq)
	statesCh := s.Channel()
	publicDataCh := p.Channel()
	s.TopicExchange(statesCh, "states") // init Exchange with name "states" of type "topic"
	p.TopicExchange(publicDataCh, "public_data")
	s.StatesConsumerInit(statesCh, "states", "sp", schema.SERVICES_PROVIDERS_COL) // init Consumer queue of topic "states.sp"
	p.PublicDataConsumerInit(publicDataCh, "public_data", "sp", schema.SERVICES_PROVIDERS_COL)

	return &ServicesProviderServer{
		log: log, db: db, ctrl: graph.NewServicesProvidersController(log, db),
		ns_ctrl:           graph.NewNamespacesController(log, db),
		ca:                graph.NewCommonActionsController(log, db),
		drivers:           make(map[string]driverpb.DriverServiceClient),
		extention_servers: make(map[string]sppb.ServicesProvidersExtentionsServiceClient),
		rdb:               rdb,
	}
}

func (s *ServicesProviderServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

func (s *ServicesProviderServer) RegisterExtentionServer(type_key string, client sppb.ServicesProvidersExtentionsServiceClient) {
	s.extention_servers[type_key] = client
}

func (s *ServicesProviderServer) ListExtentions(ctx context.Context, req *sppb.ListRequest) (res *sppb.ListExtentionsResponse, err error) {
	s.log.Debug("ListExtentions request received", zap.Any("request", req))

	keys := make([]string, 0, len(s.extention_servers))
	for k := range s.extention_servers {
		keys = append(keys, k)
	}

	return &sppb.ListExtentionsResponse{Types: keys}, nil
}

func (s *ServicesProviderServer) Test(ctx context.Context, req *sppb.ServicesProvider) (*sppb.TestResponse, error) {
	s.log.Debug("Test request received", zap.Any("request", req))

	title := req.GetTitle()
	if title == "" {
		return nil, status.Error(codes.InvalidArgument, "Services Provider 'title' is not given")
	}

	client, ok := s.drivers[req.GetType()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "Driver type '%s' not registered", req.GetType())
	}

	tfc, ok := ctx.Value(nocloud.TestFromCreate).(bool)
	tfc = ok && tfc
	if !tfc {
		for ext, data := range req.GetExtentions() {
			client, ok := s.extention_servers[ext]
			if !ok {
				return nil, status.Errorf(codes.NotFound, "Extention Server type '%s' not registered", req.GetType())
			}
			res, err := client.Test(ctx, &sppb.ServicesProvidersExtentionData{
				Data: data,
			})
			if err != nil {
				return nil, err
			}
			if !res.Result {
				err := fmt.Sprintf("Extention '%s': %s", ext, res.Error)
				return &sppb.TestResponse{
					Result: res.Result, Error: err,
				}, nil
			}
		}
	}

	test_req := &driverpb.TestServiceProviderConfigRequest{ServicesProvider: req}
	if !tfc && len(req.GetExtentions()) > 0 {
		test_req.SyntaxOnly = true
	}

	return client.TestServiceProviderConfig(ctx, test_req)
}

func (s *ServicesProviderServer) Create(ctx context.Context, req *sppb.ServicesProvider) (res *sppb.ServicesProvider, err error) {
	log := s.log.Named("Create")
	log.Debug("Create request received", zap.Any("request", req))

	testRes, err := s.Test(ctx, req)
	if err != nil {
		return req, err
	}
	if !testRes.Result {
		return req, status.Error(codes.Internal, testRes.Error)
	}

	sp := &graph.ServicesProvider{ServicesProvider: req}

	for ext, data := range req.GetExtentions() {
		client, ok := s.extention_servers[ext]
		if !ok {
			s.UnregisterExtentions(ctx, log, sp)
			return nil, status.Errorf(codes.NotFound, "Extention Server type '%s' not registered", req.GetType())
		}
		res, err := client.Register(ctx, &sppb.ServicesProvidersExtentionData{
			Data: data,
		})
		if err != nil {
			s.UnregisterExtentions(ctx, log, sp)
			return nil, err
		}
		if !res.Result {
			s.UnregisterExtentions(ctx, log, sp)
			return req, status.Errorf(codes.Internal, "Extention '%s': %s", ext, res.Error)
		}
	}

	ctx = context.WithValue(ctx, nocloud.TestFromCreate, true)
	testRes, err = s.Test(ctx, req)
	if err != nil {
		s.UnregisterExtentions(ctx, log, sp)
		return req, err
	}
	if !testRes.Result {
		s.UnregisterExtentions(ctx, log, sp)
		return req, status.Error(codes.Internal, testRes.Error)
	}

	if sp.State == nil {
		sp.State = &proto.State{State: proto.NoCloudState_INIT}
	}

	err = s.ctrl.Create(ctx, sp)
	if err != nil {
		s.UnregisterExtentions(ctx, log, sp)
		s.log.Debug("Error allocating in DataBase", zap.Any("sp", sp), zap.Error(err))
		return req, status.Error(codes.Internal, "Error allocating in DataBase")
	}
	return sp.ServicesProvider, err
}

func (s *ServicesProviderServer) UnregisterExtentions(ctx context.Context, log *zap.Logger, sp *graph.ServicesProvider) {
	log.Debug("Unregistering ServicesProvider Extentions")
	for ext, data := range sp.GetExtentions() {
		client, ok := s.extention_servers[ext]
		if !ok {
			continue // TODO add Warnings
		}
		res, err := client.Unregister(ctx, &sppb.ServicesProvidersExtentionData{
			Data: data,
		})
		if err != nil {
			log.Error("Error unregistering extension", zap.Error(err))
			continue // TODO add Warnings
		}
		if !res.Result {
			log.Error("Error unregistering extension", zap.Any("result", res))
			continue // TODO add Warnings
		}
	}
}

func (s *ServicesProviderServer) Delete(ctx context.Context, req *sppb.DeleteRequest) (res *sppb.DeleteResponse, err error) {
	log := s.log.Named("Delete")
	log.Debug("Request received", zap.Any("request", req))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns, err := s.ns_ctrl.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	ok := s.ca.HasAccess(ctx, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	sp, err := s.ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		log.Error("Error getting ServicesProvider from DB", zap.Error(err))
		return nil, status.Error(codes.NotFound, "ServicesProvider not Found in DB")
	}

	services, err := s.ctrl.GetServices(ctx, sp)
	if err != nil {
		log.Error("Error getting provisioned Services from DB", zap.Error(err))
		return nil, status.Error(codes.Internal, "Couldn't get Provisioned Services")
	}

	if len(services) > 0 {
		res = &sppb.DeleteResponse{Result: false, Services: make([]string, len(services))}
		for i, service := range services {
			res.Services[i] = service.GetUuid()
		}
		return res, nil
	}

	err = s.ctrl.DeleteEdges(ctx, sp.DocumentMeta.ID.String())
	if err != nil {
		log.Error("Error deleting edges", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error deleting edges")
	}

	err = s.ctrl.Delete(ctx, sp.ServicesProvider)
	if err != nil {
		log.Error("Error deleting ServicesProvider", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error deleting ServicesProvider")
	}

	s.UnregisterExtentions(ctx, log, sp)
	return &sppb.DeleteResponse{Result: true}, nil
}

func (s *ServicesProviderServer) Update(ctx context.Context, req *sppb.ServicesProvider) (res *sppb.ServicesProvider, err error) {
	log := s.log.Named("Update")
	log.Debug("Update request received", zap.Any("request", req))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns, err := s.ns_ctrl.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	ok := s.ca.HasAccess(ctx, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	oldSp, err := s.ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		log.Error("Error getting ServicesProvider from DB", zap.Error(err))
		return nil, status.Error(codes.NotFound, "ServicesProvider not Found in DB")
	}

	if oldSp.GetStatus() == stpb.NoCloudStatus_DEL {
		log.Error("Error updating ServicesProvider. SP is Deleted")
		return nil, status.Error(codes.Canceled, "ServicesProvider is Deleted")
	}

	sp := &graph.ServicesProvider{ServicesProvider: oldSp.ServicesProvider}
	if newTitle := req.GetTitle(); newTitle != "" {
		sp.Title = newTitle
	}
	if newSecrets := req.GetSecrets(); newSecrets != nil {
		sp.Secrets = newSecrets
	}
	if newProxy := req.GetProxy(); newProxy != nil {
		sp.Proxy = newProxy
	}
	if newVars := req.GetVars(); newVars != nil {
		sp.Vars = newVars
	}
	if newLocations := req.GetLocations(); newLocations != nil {
		if len(newLocations) == 1 && newLocations[0].Id == "_nocloud.remove" {
			newLocations = []*sppb.LocationConf{}
		}
		sp.Locations = newLocations
	}
	if newPublicData := req.GetPublicData(); newPublicData != nil {
		sp.PublicData = newPublicData
	}
	sp.Meta = req.GetMeta()
	sp.Public = req.GetPublic()

	testRes, err := s.Test(ctx, sp.ServicesProvider)
	if err != nil {
		return req, err
	}
	if !testRes.Result {
		return req, status.Error(codes.Internal, testRes.Error)
	}

	log.Info("meta", zap.Any("meta", sp.ServicesProvider.GetMeta()))

	err = s.ctrl.Update(ctx, sp.ServicesProvider)
	if err != nil {
		s.log.Debug("Error updating in DataBase", zap.Any("sp", sp), zap.Error(err))
		return req, status.Error(codes.Internal, "Error updating in DataBase")
	}
	return sp.ServicesProvider, err
}

func (s *ServicesProviderServer) Get(ctx context.Context, request *sppb.GetRequest) (res *sppb.ServicesProvider, err error) {
	log := s.log.Named("Get")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	r, err := s.ctrl.Get(ctx, request.GetUuid())
	if err != nil {
		log.Debug("Error getting ServicesProvider from DB", zap.Error(err))
		return nil, status.Error(codes.NotFound, "ServicesProvider not Found in DB")
	}

	ns, err := s.ns_ctrl.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	acc := access.Access{
		Level: access.Level_NONE, Role: "",
	}

	if err != nil {
		log.Warn("Error retrieving Namespace", zap.Error(err))

	} else if ns.Access != nil {
		acc = *ns.Access
	}

	if acc.Level != access.Level_ROOT && r.GetStatus() == stpb.NoCloudStatus_DEL {
		log.Debug("Error getting ServicesProvider")
		return nil, status.Error(codes.PermissionDenied, "Not enough rights")
	}

	if acc.Level < access.Level_ROOT {
		r.Secrets = nil
	}
	if acc.Level < access.Level_ADMIN {
		r.Vars = nil
	}

	return r.ServicesProvider, nil
}

func (s *ServicesProviderServer) List(ctx context.Context, req *sppb.ListRequest) (res *sppb.ListResponse, err error) {
	log := s.log.Named("List")
	log.Debug("Request received", zap.Any("request", req), zap.Any("context", ctx))

	var requestor string
	if !req.Anonymously {
		requestor = ctx.Value(nocloud.NoCloudAccount).(string)
	}
	log.Debug("Requestor", zap.String("id", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	isRoot := s.ca.HasAccess(ctx, requestor, ns, access.Level_ROOT)

	r, err := s.ctrl.List(ctx, requestor, isRoot)
	if err != nil {
		log.Debug("Error reading ServicesProviders from DB", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error reading ServicesProviders from DB")
	}

	res = &sppb.ListResponse{Pool: []*sppb.ServicesProvider{}}
	for i, sp := range r {
		if sp.GetStatus() != stpb.NoCloudStatus_DEL || (sp.GetStatus() == stpb.NoCloudStatus_DEL && req.ShowDeleted && isRoot) {
			res.Pool = append(res.Pool, r[i].ServicesProvider)
		}
	}

	return res, nil
}

func (s *ServicesProviderServer) BindPlan(ctx context.Context, req *sppb.BindPlanRequest) (res *sppb.BindPlanResponse, err error) {
	log := s.log.Named("BindPlan")
	log.Debug("Request received", zap.Any("request", req))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns, err := s.ns_ctrl.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	ok := s.ca.HasAccess(ctx, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	for _, plan := range req.GetPlans() {
		err = s.ctrl.BindPlan(ctx, req.Uuid, plan)
		if err != nil {
			return nil, err
		}
	}

	sp, err := s.ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}

	if sp.GetMeta() == nil {
		sp.Meta = make(map[string]*structpb.Value)
	}

	plans, ok := sp.GetMeta()["plans"]
	var reqPlans = req.GetPlans()
	var plansInterface = make([]interface{}, len(reqPlans))
	for i, v := range reqPlans {
		plansInterface[i] = v
	}

	newPlansPb, _ := structpb.NewList(plansInterface)

	if !ok {
		plans = structpb.NewListValue(newPlansPb)
	} else {
		plans.GetListValue().Values = append(plans.GetListValue().GetValues(), newPlansPb.GetValues()...)
	}

	sp.Meta["plans"] = plans

	err = s.ctrl.Update(ctx, sp.ServicesProvider)
	if err != nil {
		return nil, err
	}

	return &sppb.BindPlanResponse{}, err
}

func (s *ServicesProviderServer) UnbindPlan(ctx context.Context, req *sppb.UnbindPlanRequest) (res *sppb.UnbindPlanResponse, err error) {
	log := s.log.Named("UnbindPlan")
	log.Debug("Request received", zap.Any("request", req), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns, err := s.ns_ctrl.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	ok := s.ca.HasAccess(ctx, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	for _, plan := range req.GetPlans() {
		err = s.ctrl.UnbindPlan(ctx, req.Uuid, plan)
		if err != nil {
			return nil, err
		}
	}

	sp, err := s.ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}

	if sp.GetMeta() == nil {
		return &sppb.UnbindPlanResponse{}, nil
	}

	plans, ok := sp.GetMeta()["plans"]

	if !ok {
		return &sppb.UnbindPlanResponse{}, nil
	}

	plansValues := plans.GetListValue().GetValues()

	var newPlansValues = make([]*structpb.Value, 0)

	var newPlansMap = make(map[string]struct{})
	reqPlans := req.GetPlans()

	for i := range reqPlans {
		newPlansMap[reqPlans[i]] = struct{}{}
	}

	for i := range plansValues {
		if _, ok := newPlansMap[plansValues[i].GetStringValue()]; ok {
			continue
		}
		newPlansValues = append(newPlansValues, plansValues[i])
	}

	plans.GetListValue().Values = newPlansValues

	sp.Meta["plans"] = plans

	err = s.ctrl.Update(ctx, sp.ServicesProvider)
	if err != nil {
		return nil, err
	}

	return &sppb.UnbindPlanResponse{}, nil
}

func (s *ServicesProviderServer) Invoke(ctx context.Context, req *sppb.InvokeRequest) (*sppb.InvokeResponse, error) {
	log := s.log.Named("Invoke")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	sp, err := s.ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to get ServicesProvider", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	client, ok := s.drivers[sp.Type]
	if !ok {
		log.Error("Failed to get driver", zap.String("type", sp.Type))
		return nil, status.Error(codes.NotFound, "Driver not found")
	}

	invoke, err := client.SpInvoke(ctx, &driverpb.SpInvokeRequest{
		ServicesProvider: sp.ServicesProvider,
		Method:           req.Method,
		Params:           req.Params,
	})

	var event = &elpb.Event{
		Entity:    schema.SERVICES_PROVIDERS_COL,
		Uuid:      req.GetUuid(),
		Scope:     "driver",
		Action:    req.Method,
		Rc:        0,
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	if err != nil {
		event.Rc = 1
		nocloud.Log(log, event)
		return invoke, err
	}

	nocloud.Log(log, event)
	return invoke, nil
}

func (s *ServicesProviderServer) Prep(ctx context.Context, req *sppb.PrepSP) (*sppb.PrepSP, error) {
	log := s.log.Named("Prep")

	ns, err := s.ns_ctrl.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	if ns.Access == nil || ns.Access.Level != access.Level_ROOT {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Preparation")
	}

	sp := req.GetSp()

	if sp == nil {
		return nil, status.Error(codes.InvalidArgument, "ServicesProvider base config is not present")
	}

	client, ok := s.drivers[sp.Type]
	if !ok {
		log.Error("Failed to get driver", zap.String("type", sp.Type))
		return nil, status.Error(codes.NotFound, "Driver not found")
	}

	return client.SpPrep(ctx, req)
}
