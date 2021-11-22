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

	"github.com/arangodb/go-driver"
	driverpb "github.com/slntopp/nocloud/pkg/drivers/instance/vanilla"
	"github.com/slntopp/nocloud/pkg/graph"
	sppb "github.com/slntopp/nocloud/pkg/services_providers/proto"
	"go.uber.org/zap"
)

type ServicesProviderServer struct {
	sppb.UnimplementedServicesProvidersServiceServer

	drivers map[string]driverpb.DriverServiceClient
	db driver.Database
	ctrl graph.ServicesProvidersController

	log *zap.Logger
}

func NewServicesProviderServer(log *zap.Logger, db driver.Database) *ServicesProviderServer {
	return &ServicesProviderServer{log: log, db: db, ctrl: graph.NewServicesProvidersController(log, db)}
}

func (s *ServicesProviderServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

func (s *ServicesProviderServer) Test(ctx context.Context, req *sppb.ServicesProvider) (*sppb.TestResponse, error) {
	s.log.Debug("Test request received", zap.Any("request", req))

	client, ok := s.drivers[req.GetType()]
	if !ok {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Driver type '%s' not registered", req.GetType()))
	}

	return client.TestServiceProviderConfig(ctx, req)
}

func (s *ServicesProviderServer) Test(ctx context.Context, req sppb.ServicesProvider) (sppb.TestResponse, error) {
	s.log.Debug("Test request received", zap.Any("request", req))

}