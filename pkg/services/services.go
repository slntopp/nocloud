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
package services

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	driverpb "github.com/slntopp/nocloud/pkg/drivers/instance/proto"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/access"
	servicespb "github.com/slntopp/nocloud/pkg/services/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServicesServiceServer struct {
	servicespb.UnimplementedServicesServiceServer
	db driver.Database
	ctrl graph.ServicesController

	drivers map[string]driverpb.DriverServiceClient

	log *zap.Logger
}

func NewServicesServer(log *zap.Logger, db driver.Database) *ServicesServiceServer {
	servicesCol, _ := db.Collection(nil, graph.INSTANCES_COL)

	return &ServicesServiceServer{
		log: log, db: db, ctrl: graph.NewServicesController(log, servicesCol),
	}
}

func (s *ServicesServiceServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

func (s *ServicesServiceServer) ValidateServiceConfig(ctx context.Context, request *servicespb.ValidateServiceConfigRequest) (*servicespb.ValidateServiceConfigResponse, error) {
	log := s.log.Named("ValidateServiceConfig")
	log.Debug("Get request received", zap.Any("request", request), zap.Any("context", ctx))
	ctx, err := nocloud.ValidateMetadata(ctx, log)
	if err != nil {
		return nil, err
	}
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	response := servicespb.ValidateServiceConfigResponse{Result: false, Error: make([]*servicespb.ValidateConfigSyntaxResponse, 0)}

	service := request.GetConfig()
	
	// Checking if requestor has access to Namespace Service going to be put in
	ok := graph.HasAccess(ctx, s.db, requestor, request.Namespace, access.ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Account")
	}

	log.Debug("Init validation")
	for name, group := range service.InstancesGroup {
		log.Debug("Validating Instances Group", zap.String("group", name))
		groupType := group.GetType()

		config_err := servicespb.ValidateConfigSyntaxResponse{
			Group: groupType,
		}

		client, ok := s.drivers[groupType]
		if !ok {
			response.Result = false
			msg := fmt.Sprintf("Driver Type '%w' not registered", groupType)
			config_err.Error = &driverpb.ValidateConfigSyntaxResponse{
				Result: false, Error: &msg,
			}
			response.Error = append(
				response.Error, &config_err,
			)
			continue
		}

		for _, instance := range group.Instances {
			log.Debug("Validating Service Instance", zap.String("instance", instance.GetTitle()))
			config := instance.GetConfig()
			res, err := client.ValidateConfigSyntax(ctx, &driverpb.ValidateConfigSyntaxRequest{Config: config})
			if err != nil {
				response.Result = false
				msg := fmt.Sprintf("Error Validating Config for Instance '%w' of type '%w'", instance.GetTitle(), groupType)
				config_err.Error = &driverpb.ValidateConfigSyntaxResponse{
					Result: false, Error: &msg,
				}
				response.Error = append(
					response.Error, &config_err,
				)
				continue
			}

			if !res.Result {
				response.Result = false
				response.Error = append(response.Error, &servicespb.ValidateConfigSyntaxResponse{
					Instance: instance.GetTitle(),
					Error: res,
				})
			}
		}
	}

	return &response, nil
}