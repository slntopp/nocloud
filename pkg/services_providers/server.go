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
package services_providers

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	driverpb "github.com/slntopp/nocloud/pkg/drivers/instance/vanilla"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	sppb "github.com/slntopp/nocloud/pkg/services_providers/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type ServicesProviderServer struct {
	sppb.UnimplementedServicesProvidersServiceServer

	drivers map[string]driverpb.DriverServiceClient
	db driver.Database
	ctrl graph.ServicesProvidersController
	ns_ctrl graph.NamespacesController

	log *zap.Logger
}

func NewServicesProviderServer(log *zap.Logger, db driver.Database) *ServicesProviderServer {
	return &ServicesProviderServer{
		log: log, db: db, ctrl: graph.NewServicesProvidersController(log, db),
		ns_ctrl: graph.NewNamespacesController(log, db),
		drivers: make(map[string]driverpb.DriverServiceClient),
	}
}

func (s *ServicesProviderServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

func (s *ServicesProviderServer) Test(ctx context.Context, req *sppb.ServicesProvider) (*sppb.TestResponse, error) {
	s.log.Debug("Test request received", zap.Any("request", req))

	title := req.GetTitle()
	if title == "" {
		return nil, status.Error(codes.InvalidArgument, "Services Provider 'title' is not given")
	}

	client, ok := s.drivers[req.GetType()]
	if !ok {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Driver type '%s' not registered", req.GetType()))
	}

	return client.TestServiceProviderConfig(ctx, req)
}

func (s *ServicesProviderServer) Create(ctx context.Context, req *sppb.ServicesProvider) (res *sppb.ServicesProvider, err error) {
	s.log.Debug("Create request received", zap.Any("request", req))

	testRes, err := s.Test(ctx, req)
	if err != nil {
		return req, err
	}
	if !testRes.Result {
		return req, status.Error(codes.Internal, testRes.Error)
	}

	sp := &graph.ServicesProvider{ServicesProvider: req}
	err = s.ctrl.Create(ctx, sp)
	return sp.ServicesProvider, err
}

func (s *ServicesProviderServer) Get(ctx context.Context, request *sppb.GetRequest) (res *sppb.ServicesProvider, err error) {
	log := s.log.Named("Get")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	ctx, err = nocloud.ValidateMetadata(ctx, log)
	if err != nil {
		return nil, err
	}
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	r, err := s.ctrl.Get(ctx, request.GetUuid())
	if err != nil {
		log.Debug("Error getting Service from DB", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not Found in DB")
	}

	return r.ServicesProvider, nil
}

func (s *ServicesProviderServer) List(ctx context.Context, req *sppb.ListRequest) (res *sppb.ListResponse, err error) {
	log := s.log.Named("List")
	log.Debug("Request received", zap.Any("request", req), zap.Any("context", ctx))

	ctx, err = nocloud.ValidateMetadata(ctx, log)
	if err != nil {
		return nil, err
	}
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	r, err := s.ctrl.List(ctx, requestor)
	if err != nil {
		log.Debug("Error reading ServicesProviders from DB", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error reading ServicesProviders from DB")
	}

	res = &sppb.ListResponse{Pool: make([]*sppb.ServicesProvider, len(r))}
	for i, sp := range r {
		res.Pool[i] = sp.ServicesProvider
	}

	return res, nil
}

func (s *ServicesProviderServer) Invoke(ctx context.Context, req *sppb.ActionRequest) (res *structpb.Struct, err error) {
	log := s.log.Named("Invoke")
	log.Debug("Request received", zap.Any("request", req), zap.Any("context", ctx))

	ctx, err = nocloud.ValidateMetadata(ctx, log)
	if err != nil {
		return nil, err
	}
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns, err := s.ns_ctrl.Get(ctx, "0")
	if err != nil {
		return nil, err
	}
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID.String(), 3)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	sp, err := s.ctrl.Get(ctx, req.GetServicesProvider().GetUuid())
	if err != nil {
		log.Error("Error getting services provider",
			zap.String("services_provider", req.GetServicesProvider().GetUuid()),
			zap.Error(err),
		)
		return nil, status.Error(codes.NotFound, "ServicesProvider not found")
	}

	req.ServicesProvider = sp.ServicesProvider

	client, ok := s.drivers[sp.GetType()]
	if !ok {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Driver type '%s' not registered", sp.GetType()))
	}

	return client.Invoke(ctx, req)
}