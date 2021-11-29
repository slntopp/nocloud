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
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	servicespb "github.com/slntopp/nocloud/pkg/services/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServicesServiceServer struct {
	servicespb.UnimplementedServicesServiceServer
	db driver.Database
	ctrl graph.ServicesController
	sp_ctrl graph.ServicesProvidersController
	ns_ctrl graph.NamespacesController

	drivers map[string]driverpb.DriverServiceClient

	log *zap.Logger
}

func NewServicesServer(log *zap.Logger, db driver.Database) *ServicesServiceServer {
	return &ServicesServiceServer{
		log: log, db: db, ctrl: graph.NewServicesController(log, db),
		sp_ctrl: graph.NewServicesProvidersController(log, db),
		ns_ctrl: graph.NewNamespacesController(log, db),
		drivers: make(map[string]driverpb.DriverServiceClient),
	}
}

type InstancesGroupDriverContext struct {
	sp *graph.ServicesProvider
	client *driverpb.DriverServiceClient
}

func (s *ServicesServiceServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

func (s *ServicesServiceServer) DoTestServiceConfig(ctx context.Context, log *zap.Logger, request *servicespb.CreateRequest) (*servicespb.TestConfigResponse, *graph.Namespace, error) {
	ctx, err := nocloud.ValidateMetadata(ctx, log)
	if err != nil {
		return nil, nil, err
	}
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	response := &servicespb.TestConfigResponse{Result: true, Errors: make([]*servicespb.TestConfigError, 0)}

	namespace, err := s.ns_ctrl.Get(ctx, request.GetNamespace())
	if err != nil {
		s.log.Debug("Error getting namespace", zap.Error(err))
		return nil, nil, status.Error(codes.NotFound, "Namespace not found")
	}
	// Checking if requestor has access to Namespace Service going to be put in
	ok := graph.HasAccess(ctx, s.db, requestor, namespace.ID.String(), access.ADMIN)
	if !ok {
		return nil, nil, status.Error(codes.PermissionDenied, "Not enough access rights to Namespace")
	}

	service := request.GetService()
	groups  := service.GetInstancesGroups()

	log.Debug("Init validation", zap.Any("groups", groups), zap.Int("amount", len(groups)))
	for name, group := range service.GetInstancesGroups() {
		log.Debug("Validating Instances Group", zap.String("group", name))
		groupType := group.GetType()

		config_err := servicespb.TestConfigError{
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
			errors := make([]*servicespb.TestConfigError, 0)
			for _, confErr := range res.Errors {
				errors = append(errors, &servicespb.TestConfigError{
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

	return response, &namespace, nil
}

func (s *ServicesServiceServer) TestConfig(ctx context.Context, request *servicespb.CreateRequest) (*servicespb.TestConfigResponse, error) {
	log := s.log.Named("TestServiceConfig")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))
	response, _, err := s.DoTestServiceConfig(ctx, log, request)
	return response, err
}

func (s *ServicesServiceServer) Create(ctx context.Context, request *servicespb.CreateRequest) (*servicespb.Service, error) {
	log := s.log.Named("CreateService")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))
	testResult, namespace, err := s.DoTestServiceConfig(ctx, log, request)

	if err != nil {
		return nil, err
	} else if !testResult.Result {
		return nil, status.Error(codes.InvalidArgument, "Config didn't pass test")
	}

	service := request.GetService()
	doc, err := s.ctrl.Create(ctx, service)
	if err != nil {
		log.Error("Error while creating service", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while creating Service")
	}

	err = s.ctrl.Join(ctx, doc, namespace, access.ADMIN, roles.OWNER)	
	if err != nil {
		log.Error("Error while joining service to namespace", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while joining service to namespace")
	}
	return service, nil
}

func (s *ServicesServiceServer) Up(ctx context.Context, request *servicespb.UpRequest) (*servicespb.UpResponse, error) {
	log := s.log.Named("Up")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	service, err := s.ctrl.Get(ctx, request.GetId())
	if err != nil {
		log.Debug("Error getting Service", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not found")
	}
	log.Debug("Found Service", zap.Any("service", service))

	deploy_policies := request.GetDeployPolicies()
	contexts := make(map[string]*InstancesGroupDriverContext)

	for _, group := range service.GetInstancesGroups() {
		sp_id := deploy_policies[group.GetUuid()]
		sp, err := s.sp_ctrl.Get(ctx, sp_id)
		if err != nil {
			s.log.Error("Error getting ServiceProvider", zap.Error(err), zap.String("id", sp_id))
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Error getting ServiceProvider(%s)", sp_id))
		}
		
		groupType := group.GetType()
		client, ok := s.drivers[groupType]
		if !ok {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Driver of type '%s' not registered", groupType))
		}
		contexts[group.GetUuid()] = &InstancesGroupDriverContext{sp, &client}
	}

	service.Status = "starting"
	s.log.Debug("Updated Service", zap.Any("service", service))
	err = s.ctrl.Update(ctx, service.Service)
	if err != nil {
		s.log.Error("Error updating Service", zap.Error(err), zap.Any("service", service))
		return nil, status.Error(codes.Internal, "Error storing updates")
	}

	for _, group := range service.GetInstancesGroups() {
		c, ok := contexts[group.GetUuid()]
		if !ok {
			log.Debug("Instance Group has no context", zap.String("group", group.GetUuid()), zap.String("service", service.GetUuid()))
			continue
		}
		client := *c.client
		sp := c.sp

		response, err := client.Up(ctx, &driverpb.UpRequest{Group: group, ServicesProvider: sp.ServicesProvider})
		if err != nil {
			s.log.Error("Error deploying group", zap.Any("service_provider", sp), zap.Any("group", group), zap.Error(err))
			continue
		}
		s.log.Debug("Up Request Result", zap.Any("response", response))

		// TODO: Change to Hash comparation
		// TODO: Add cleanups
		if len(group.Instances) != len(response.GetGroup().GetInstances()) {
			s.log.Error("Instances config changed by Driver")
			continue
		}
		for i, instance := range response.GetGroup().GetInstances() {
			group.Instances[i].Data = instance.GetData()
		}
		
		group.Data = response.GetGroup().GetData()
		err = s.ctrl.Provide(ctx, sp.ID, service.ID, group.GetUuid())
		if err != nil {
			s.log.Error("Error linking group to ServiceProvider", zap.Any("service_provider", sp.GetUuid()), zap.Any("group", group), zap.Error(err))
			continue
		}
		s.log.Debug("Updated Group", zap.Any("group", group))
	}
	
	service.Status = "up"
	s.log.Debug("Updated Service", zap.Any("service", service))
	err = s.ctrl.Update(ctx, service.Service)
	if err != nil {
		s.log.Error("Error updating Service", zap.Error(err), zap.Any("service", service))
		return nil, status.Error(codes.Internal, "Error storing updates")
	}

	return &servicespb.UpResponse{}, nil
}

func (s *ServicesServiceServer) Down(ctx context.Context, request *servicespb.DownRequest) (*servicespb.DownResponse, error) {
	log := s.log.Named("Down")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	service, err := s.ctrl.Get(ctx, request.GetId())
	if err != nil {
		log.Debug("Error getting Service", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not found")
	}
	log.Debug("Found Service", zap.Any("service", service))

	provisions, err := s.ctrl.GetProvisions(ctx, service.ID.String())
	if err != nil {
		s.log.Debug("Can't get provisions for Service", zap.Any("service", service), zap.Error(err))
		return nil, status.Error(codes.Internal, "Can't gather Service provisions")
	}
	contexts := make(map[string]*InstancesGroupDriverContext)

	for _, group := range service.GetInstancesGroups() {
		sp_id, ok := provisions[group.GetUuid()]
		if !ok {
			log.Debug("Instance Group has not provision", zap.String("group", group.GetUuid()), zap.String("service", service.GetUuid()))
			continue
		}
		sp, err := s.sp_ctrl.Get(ctx, sp_id)
		if err != nil {
			s.log.Error("Error getting ServiceProvider", zap.Error(err), zap.String("id", sp_id))
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Error getting ServiceProvider(%s)", sp_id))
		}

		groupType := group.GetType()
		client, ok := s.drivers[groupType]
		if !ok {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Driver of type '%s' not registered", groupType))
		}

		contexts[group.GetUuid()] = &InstancesGroupDriverContext{sp, &client}
	}

	service.Status = "stopping"
	s.log.Debug("Updated Service", zap.Any("service", service))
	err = s.ctrl.Update(ctx, service.Service)
	if err != nil {
		s.log.Error("Error updating Service", zap.Error(err), zap.Any("service", service))
		return nil, status.Error(codes.Internal, "Error storing updates")
	}

	for _, group := range service.GetInstancesGroups() {		
		c, ok := contexts[group.GetUuid()]
		if !ok {
			log.Debug("Instance Group has no context, i.e. provision", zap.String("group", group.GetUuid()), zap.String("service", service.GetUuid()))
			continue
		}
		client := *c.client
		sp := c.sp

		_, err := client.Down(ctx, &driverpb.DownRequest{Group: group, ServicesProvider: sp.ServicesProvider})
		if err != nil {
			s.log.Error("Error undeploying group", zap.Any("service_provider", sp), zap.Any("group", group), zap.Error(err))
			continue
		}
		group.Data = nil
		err = s.ctrl.Unprovide(ctx, group.GetUuid())
		if err != nil {
			s.log.Error("Error unlinking group from ServiceProvider", zap.Any("service_provider", sp.GetUuid()), zap.Any("group", group), zap.Error(err))
			continue
		}
	}

	service.Status = "down"
	s.log.Debug("Updated Service", zap.Any("service", service))
	err = s.ctrl.Update(ctx, service.Service)
	if err != nil {
		s.log.Error("Error updating Service", zap.Error(err), zap.Any("service", service))
		return nil, status.Error(codes.Internal, "Error storing updates")
	}

	return &servicespb.DownResponse{}, nil
}

func (s *ServicesServiceServer) Get(ctx context.Context, request *servicespb.GetRequest) (res *servicespb.Service, err error) {
	log := s.log.Named("Get")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	ctx, err = nocloud.ValidateMetadata(ctx, log)
	if err != nil {
		return nil, err
	}
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	r, err := s.ctrl.Get(ctx, request.GetId())
	if err != nil {
		log.Debug("Error getting Service from DB", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not Found in DB")
	}

	ok := graph.HasAccess(ctx, s.db, requestor, r.ID.String(), access.READ)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights")
	}

	return r.Service, nil
}

func (s *ServicesServiceServer) List(ctx context.Context, request *servicespb.ListRequest) (response *servicespb.ListResponse, err error) {
	log := s.log.Named("List")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	ctx, err = nocloud.ValidateMetadata(ctx, log)
	if err != nil {
		return nil, err
	}
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	r, err := s.ctrl.List(ctx, requestor, request.Depth)
	if err != nil {
		log.Debug("Error reading Services from DB", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error reading Services from DB")
	}

	response = &servicespb.ListResponse{Services: make([]*servicespb.Service, len(r))}
	for i, service := range r {
		response.Services[i] = service.Service
	}

	return response, nil
}