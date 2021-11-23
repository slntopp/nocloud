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
	driverpb "github.com/slntopp/nocloud/pkg/drivers/instance/vanilla"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/instances/proto"
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
	return &ServicesServiceServer{
		log: log, db: db, ctrl: graph.NewServicesController(log, db), drivers: make(map[string]driverpb.DriverServiceClient),
	}
}

func (s *ServicesServiceServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

func (s *ServicesServiceServer) DoTestServiceConfig(ctx context.Context, log *zap.Logger, request *servicespb.CreateServiceRequest) (*servicespb.TestServiceConfigResponse, string, error) {
	ctx, err := nocloud.ValidateMetadata(ctx, log)
	if err != nil {
		return nil, "", err
	}
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	response := &servicespb.TestServiceConfigResponse{Result: true, Errors: make([]*servicespb.TestServiceConfigError, 0)}

	// Checking if requestor has access to Namespace Service going to be put in
	ok := graph.HasAccess(ctx, s.db, requestor, fmt.Sprintf("%s/%s", graph.NAMESPACES_COL, request.Namespace), access.ADMIN)
	if !ok {
		return nil, requestor, status.Error(codes.PermissionDenied, "Not enough access rights to Namespace")
	}

	service := request.GetService()
	groups  := service.GetInstancesGroups()

	log.Debug("Init validation", zap.Any("groups", groups), zap.Int("amount", len(groups)))
	for name, group := range service.GetInstancesGroups() {
		log.Debug("Validating Instances Group", zap.String("group", name))
		groupType := group.GetType()

		config_err := servicespb.TestServiceConfigError{
			InstanceGroup: name,
		}

		client, ok := s.drivers[groupType]
		if !ok {
			response.Result = false
			config_err.Error = fmt.Sprintf("Driver Type '%s' not registered", groupType)
			response.Errors = append(
				response.Errors, &config_err,
			)
			continue
		}

		res, err := client.TestInstancesGroupConfig(ctx, &proto.TestInstancesGroupConfigRequest{Group: group})
		if err != nil {
			response.Result = false
			config_err.Error = fmt.Sprintf("Error validating group '%s'", name)
			response.Errors = append(
				response.Errors, &config_err,
			)
			continue
		}
		if !res.GetResult() {
			response.Result = false
			errors := make([]*servicespb.TestServiceConfigError, 0)
			for _, confErr := range res.Errors {
				errors = append(errors, &servicespb.TestServiceConfigError{
					Error: confErr.Error,
					Instance: confErr.Instance,
					InstanceGroup: name,
				})
			}
			response.Errors = append(response.Errors, errors...)
			continue
		}
		log.Debug("Validated Instances Group", zap.String("group", name))
	}

	return response, requestor, nil
}

func (s *ServicesServiceServer) TestServiceConfig(ctx context.Context, request *servicespb.CreateServiceRequest) (*servicespb.TestServiceConfigResponse, error) {
	log := s.log.Named("TestServiceConfig")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))
	response, _, err := s.DoTestServiceConfig(ctx, log, request)
	return response, err
}

}