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
package services

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/cskr/pubsub"
	"github.com/slntopp/nocloud-proto/access"
	bpb "github.com/slntopp/nocloud-proto/billing"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	elpb "github.com/slntopp/nocloud-proto/events_logging"
	proto "github.com/slntopp/nocloud-proto/instances"
	pb "github.com/slntopp/nocloud-proto/services"
	stpb "github.com/slntopp/nocloud-proto/settings"
	spb "github.com/slntopp/nocloud-proto/states"
	statuspb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	sc "github.com/slntopp/nocloud/pkg/settings/client"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ServicesServer struct {
	pb.UnimplementedServicesServiceServer
	db      driver.Database
	ctrl    graph.ServicesController
	sp_ctrl graph.ServicesProvidersController
	ns_ctrl graph.NamespacesController

	drivers map[string]driverpb.DriverServiceClient

	billing bpb.BillingServiceClient

	ps *pubsub.PubSub

	log *zap.Logger
}

func NewServicesServer(_log *zap.Logger, db driver.Database, ps *pubsub.PubSub) *ServicesServer {
	log := _log.Named("ServicesServer")
	log.Debug("New Services Server Creating")

	return &ServicesServer{
		log: log, db: db, ctrl: graph.NewServicesController(log, db),
		sp_ctrl: graph.NewServicesProvidersController(log, db),
		ns_ctrl: graph.NewNamespacesController(log, db),
		drivers: make(map[string]driverpb.DriverServiceClient),
		ps:      ps,
	}
}

type InstancesGroupDriverContext struct {
	sp     *graph.ServicesProvider
	client *driverpb.DriverServiceClient
}

func (s *ServicesServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

func (s *ServicesServer) SetupSettingsClient(stC stpb.SettingsServiceClient, internal_token string) {
	sc.Setup(
		s.log, metadata.AppendToOutgoingContext(
			context.Background(), "authorization", "bearer "+internal_token,
		), &stC,
	)

	var ibps InstanceBillingPlanSettings
	err := sc.Fetch(IBPSKey, &ibps, defaultBillingPlanSettings)
	if err != nil {
		s.log.Error("Error fetching instance billing plan settings", zap.Error(err))
	}
}

func (s *ServicesServer) SetupBillingClient(bC bpb.BillingServiceClient) {
	s.billing = bC
}

type InstanceBillingPlanSettings struct {
	Required bool `json:"required"` // each instance must have it
}

const IBPSKey = "instance-billing-plan-settings"

var (
	defaultBillingPlanSettings = &sc.Setting[InstanceBillingPlanSettings]{
		Value: InstanceBillingPlanSettings{
			Required: true,
		},
		Description: "Instances Billing Plans Settings",
		Level:       access.Level_ADMIN,
	}
)

var getStatusQuey = `
	LET acc = DOCUMENT(@account)
	return acc.suspended
`

func (s *ServicesServer) CheckRequestorStatus(ctx context.Context, id driver.DocumentID) bool {
	cur, _ := s.db.Query(ctx, getStatusQuey, map[string]interface{}{
		"account": id,
	})

	isSuspended := false

	for cur.HasMore() {
		cur.ReadDocument(ctx, &isSuspended)
	}

	return isSuspended
}

func (s *ServicesServer) DoTestServiceConfig(ctx context.Context, log *zap.Logger, service *pb.Service) (*pb.TestConfigResponse, error) {
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	response := &pb.TestConfigResponse{Result: true, Errors: make([]*pb.TestConfigError, 0)}

	var ibps InstanceBillingPlanSettings
	err := sc.Fetch(IBPSKey, &ibps, defaultBillingPlanSettings)
	if err != nil {
		log.Error("Error fetching instance billing plan settings", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while fetching Instance Billing Plan Settings")
	}
	log.Debug("Instance billing plan settings", zap.Any("settings", ibps))

	groups := service.GetInstancesGroups()

	bp_cache := make(map[string]*bpb.Plan)

	log.Debug("Init validation", zap.Int("amount", len(groups)))
	for _, group := range service.GetInstancesGroups() {
		log.Debug("Validating Instances Group", zap.Any("group", group))
		groupType := group.GetType()

		for _, instance := range group.GetInstances() {
			log.Debug("Instance BillingPlan is", zap.Bool("provided", instance.BillingPlan != nil))
			if ibps.Required && instance.BillingPlan == nil {
				response.Result = false
				log.Error("ibps req", zap.Bool("ibps", ibps.Required))
				log.Error("instance.BillingPlan", zap.Any("bill plan", instance.BillingPlan))
				terr := pb.TestConfigError{
					Error:         "Instance has no billing plan and no default is set",
					Instance:      instance.Title,
					InstanceGroup: group.Title,
				}
				response.Errors = append(response.Errors, &terr)
				continue
			}
			if instance.BillingPlan != nil && instance.BillingPlan.Uuid != "" {
				_, ok := bp_cache[instance.BillingPlan.Uuid]
				if !ok {
					_, err := s.billing.GetPlan(ctx, instance.BillingPlan)
					if err != nil {
						log.Error("Error fetching BillingPlan", zap.Error(err))
						return nil, err
					}
				}

				inst_ctrl := s.ctrl.IGController().Instances()
				old_instance, _ := inst_ctrl.Get(ctx, instance.GetUuid())

				var equal = false

				if old_instance != nil {
					oldRes := old_instance.GetResources()
					newRes := instance.GetResources()

					log.Debug("res", zap.Any("oldRes", oldRes))
					log.Debug("res", zap.Any("newRes", newRes))

					equal = reflect.DeepEqual(oldRes, newRes)
					log.Debug("equal", zap.Bool("equal", equal))
				}

				if !equal {
					err := inst_ctrl.CheckEdgeExist(ctx, group.GetSp(), instance)

					if err != nil {
						response.Result = false
						log.Error("IGCONTROLLER err", zap.String("err", err.Error()))
						terr := pb.TestConfigError{
							Error:         err.Error(),
							Instance:      instance.Title,
							InstanceGroup: group.Title,
						}
						response.Errors = append(response.Errors, &terr)
						continue
					}
				}

				err = inst_ctrl.ValidateBillingPlan(ctx, group.GetSp(), instance)
				if err != nil {
					response.Result = false
					log.Error("IGCONTROLLER err", zap.String("err", err.Error()))
					terr := pb.TestConfigError{
						Error:         err.Error(),
						Instance:      instance.Title,
						InstanceGroup: group.Title,
					}
					response.Errors = append(response.Errors, &terr)
					continue
				}
			}
		}

		config_err := pb.TestConfigError{
			InstanceGroup: group.Title,
		}

		client, ok := s.drivers[groupType]
		if !ok {
			response.Result = false
			log.Error("Not such driver", zap.Any("any", client))
			config_err.Error = fmt.Sprintf("Driver Type '%s' not registered", groupType)
			response.Errors = append(
				response.Errors, &config_err,
			)
			continue
		}

		sp, err := s.sp_ctrl.Get(ctx, group.GetSp())
		if err != nil {
			response.Result = false
			log.Error("Sp ctrl get", zap.String("any", err.Error()))
			config_err.Error = fmt.Sprintf("Error getting sp %s while validating group '%s': %v",
				group.GetSp(), group.Title, err)
			response.Errors = append(response.Errors, &config_err)
			continue
		}
		res, err := client.TestInstancesGroupConfig(ctx, &proto.TestInstancesGroupConfigRequest{
			Group: group,
			Sp:    sp.ServicesProvider,
		})
		if err != nil {
			log.Error("Test ig config err", zap.String("any", err.Error()))
			response.Result = false
			config_err.Error = fmt.Sprintf("Error validating group '%s': %v", group.Title, err)
			response.Errors = append(response.Errors, &config_err)
			continue
		}
		if !res.GetResult() {
			response.Result = false
			log.Error("No result", zap.Bool("asdqwe", res.GetResult()))
			errors := make([]*pb.TestConfigError, 0)
			for _, confErr := range res.Errors {
				errors = append(errors, &pb.TestConfigError{
					Error:         confErr.Error,
					Instance:      confErr.Instance,
					InstanceGroup: group.Title,
				})
			}
			response.Errors = append(response.Errors, errors...)
			continue
		}
		log.Debug("Validated Instances Group", zap.String("group", group.Title))
	}

	return response, nil
}

func (s *ServicesServer) TestConfig(ctx context.Context, request *pb.CreateRequest) (*pb.TestConfigResponse, error) {
	log := s.log.Named("TestServiceConfig")
	log.Debug("Request received", zap.Any("request", request))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	ns, err := s.ns_ctrl.Get(ctx, request.GetNamespace())
	if err != nil {
		s.log.Debug("Error getting namespace", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Namespace not found")
	}
	// Checking if requestor has access to Namespace Service going to be put in
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Namespace")
	}

	response, err := s.DoTestServiceConfig(ctx, log, request.GetService())
	return response, err
}

func (s *ServicesServer) Create(ctx context.Context, request *pb.CreateRequest) (*pb.Service, error) {
	log := s.log.Named("CreateService")
	log.Debug("Request received", zap.Any("request", request))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	requestorDoc := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	isSuspended := s.CheckRequestorStatus(ctx, requestorDoc)

	if isSuspended && requestor != schema.ROOT_ACCOUNT_KEY {
		return nil, status.Error(codes.Unavailable, "Requestor account is suspended")
	}

	ns, err := s.ns_ctrl.Get(ctx, request.GetNamespace())
	if err != nil {
		s.log.Debug("Error getting namespace", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Namespace not found")
	}
	// Checking if requestor has access to Namespace Service going to be put in
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Namespace")
	}

	service := request.GetService()
	contexts := make(map[string]*InstancesGroupDriverContext)

	testResult, err := s.DoTestServiceConfig(ctx, log, service)
	if err != nil {
		return nil, err
	} else if !testResult.Result {
		return nil, status.Error(codes.InvalidArgument, "Config didn't pass test")
	}

	doc, err := s.ctrl.Create(ctx, service)
	if err != nil {
		log.Error("Error while creating service", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while creating Service")
	}

	err = s.ctrl.Join(ctx, doc, &ns, access.Level_ADMIN, roles.OWNER)
	if err != nil {
		log.Error("Error while joining service to namespace", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while joining service to namespace")
	}

	for i, group := range service.GetInstancesGroups() {
		if group.Title == "" {
			return nil, status.Errorf(codes.InvalidArgument, "InstancesGroup #%d has no title", i)
		}

		if group.Sp == nil || group.GetSp() == "" {
			log.Error("Error no ServicesProvider UUID given")
			return nil, status.Errorf(codes.InvalidArgument, "Error no ServicesProvider UUID given for InstancesGroup(%s)", group.GetTitle())
		}
		sp_id := group.GetSp()
		sp, err := s.sp_ctrl.Get(ctx, sp_id)
		if err != nil {
			log.Error("Error getting ServiceProvider", zap.Error(err), zap.String("id", sp_id))
			return nil, status.Errorf(codes.InvalidArgument, "Error getting ServiceProvider(%s)", sp_id)
		}

		groupType := group.GetType()
		client, ok := s.drivers[groupType]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "Driver of type '%s' not registered", groupType)
		}
		contexts[group.GetUuid()] = &InstancesGroupDriverContext{sp, &client}
	}

	for _, group := range service.GetInstancesGroups() {
		c, ok := contexts[group.GetUuid()]
		if !ok {
			log.Debug("Instance Group has no context", zap.String("group", group.GetUuid()), zap.String("service", service.GetUuid()))
			continue
		}
		sp := c.sp

		err = s.ctrl.IGController().Provide(ctx, group.Uuid, sp.Uuid)
		if err != nil {
			log.Error("Error linking group to ServiceProvider", zap.Any("service_provider", sp.GetUuid()), zap.Any("group", group), zap.Error(err))
			continue
		}
		log.Debug("Created Group", zap.Any("group", group))
	}

	return service, nil
}

func (s *ServicesServer) Update(ctx context.Context, service *pb.Service) (*pb.Service, error) {
	log := s.log.Named("UpdateService")
	log.Debug("Request received", zap.Any("service", service))

	testResult, err := s.DoTestServiceConfig(ctx, log, service)

	if err != nil {
		return nil, err
	} else if !testResult.Result {
		return nil, status.Error(codes.InvalidArgument, "Config didn't pass test")
	}

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	docID := driver.NewDocumentID(schema.SERVICES_COL, service.Uuid)
	okAdmin := graph.HasAccess(ctx, s.db, requestor, docID, access.Level_ADMIN)
	okRoot := graph.HasAccess(ctx, s.db, requestor, docID, access.Level_ROOT)

	requestorDoc := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	isSuspended := s.CheckRequestorStatus(ctx, requestorDoc)

	if !okAdmin {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Update")
	}

	if isSuspended && !okRoot {
		return nil, status.Error(codes.Unavailable, "Requestor account is suspended")
	}

	err = s.ctrl.Update(ctx, service, true)
	if err != nil {
		log.Error("Error while updating service", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while updating Service")
	}

	service, err = s.ctrl.Get(ctx, requestor, service.GetUuid())
	if err != nil {
		log.Debug("Error getting Service from DB after Patch", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not Found in DB after Patch")
	}
	return service, nil
}

func (s *ServicesServer) Up(ctx context.Context, request *pb.UpRequest) (*pb.UpResponse, error) {
	log := s.log.Named("Up")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("request", request), zap.String("requestor", requestor))

	service, err := s.ctrl.Get(ctx, requestor, request.GetUuid())
	if err != nil {
		log.Debug("Error getting Service", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not found")
	}
	log.Debug("Found Service", zap.Any("service", service))

	if service.GetAccess().GetLevel() < access.Level_ADMIN {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Service")
	}

	if service.GetStatus() == statuspb.NoCloudStatus_SUS {
		return nil, status.Error(codes.PermissionDenied, "Can't deploy suspended service")
	}

	contexts := make(map[string]*InstancesGroupDriverContext)

	for _, group := range service.GetInstancesGroups() {
		sp_id := group.GetSp()
		sp, err := s.sp_ctrl.Get(ctx, sp_id)
		if err != nil {
			log.Error("Error getting ServiceProvider", zap.Error(err), zap.String("id", sp_id))
			return nil, status.Errorf(codes.InvalidArgument, "Error getting ServiceProvider(%s)", sp_id)
		}

		groupType := group.GetType()
		client, ok := s.drivers[groupType]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "Driver of type '%s' not registered", groupType)
		}
		contexts[group.GetUuid()] = &InstancesGroupDriverContext{sp, &client}
	}

	err = s.ctrl.SetStatus(ctx, service, statuspb.NoCloudStatus_PROC)
	if err != nil {
		log.Error("Error updating Service", zap.Error(err), zap.Any("service", service))
		return nil, status.Error(codes.Internal, "Error storing updates")
	}

	result := &pb.UpResponse{Errors: make([]*pb.UpError, 0)}
	for _, group := range service.GetInstancesGroups() {
		c, ok := contexts[group.GetUuid()]
		if !ok {
			log.Debug("Instance Group has no context", zap.String("group", group.GetUuid()), zap.String("service", service.GetUuid()))
			continue
		}
		client := *c.client
		sp := c.sp

		group.Status = statuspb.NoCloudStatus_UP

		response, err := client.Up(ctx, &driverpb.UpRequest{Group: group, ServicesProvider: sp.ServicesProvider})
		if err != nil {
			log.Error("Error deploying group", zap.Any("service_provider", sp), zap.Any("group", group), zap.Error(err))
			result.Errors = append(result.Errors, &pb.UpError{
				Data: map[string]string{
					"group": group.GetUuid(),
					"error": err.Error(),
				},
			})
			continue
		}

		err = s.ctrl.IGController().SetStatus(ctx, group, statuspb.NoCloudStatus_UP)
		if err != nil {
			log.Error("Error updating InstancesGroup", zap.Error(err), zap.Any("IG", group))
		}

		log.Debug("Up Request Result", zap.Any("response", response))

		if len(group.Instances) != len(response.GetGroup().GetInstances()) {
			log.Error("Instances config changed by Driver")
			result.Errors = append(result.Errors, &pb.UpError{
				Data: map[string]string{
					"group": group.GetUuid(),
					"error": "Instances config changed by Driver",
				},
			})
			continue
		}

		log.Debug("Updated Group", zap.Any("group", group))
	}

	service.Status = statuspb.NoCloudStatus_UP
	log.Debug("Updated Service", zap.Any("service", service))

	event := elpb.Event{
		Entity:    schema.SERVICES_COL,
		Uuid:      service.GetUuid(),
		Scope:     "database",
		Action:    "up",
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	err = s.ctrl.SetStatus(ctx, service, statuspb.NoCloudStatus_UP)
	if err != nil {
		log.Error("Error updating Service", zap.Error(err), zap.Any("service", service))
		event.Rc = 1
		nocloud.Log(log, &event)
		return nil, status.Error(codes.Internal, "Error storing updates")
	}

	event.Rc = 0
	nocloud.Log(log, &event)

	return &pb.UpResponse{}, nil
}

func (s *ServicesServer) Suspend(ctx context.Context, request *pb.SuspendRequest) (*pb.SuspendResponse, error) {
	log := s.log.Named("Suspend")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("request", request), zap.String("requestor", requestor))

	service, err := s.ctrl.Get(ctx, requestor, request.GetUuid())
	if err != nil {
		log.Debug("Error getting Service", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not found")
	}
	log.Debug("Found Service", zap.Any("service", service))

	if service.GetAccess().GetLevel() < access.Level_ROOT {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Service")
	}

	for _, group := range service.GetInstancesGroups() {
		groupController := s.ctrl.IGController()
		instancesController := groupController.Instances()

		if err := groupController.SetStatus(ctx, group, statuspb.NoCloudStatus_SUS); err != nil {
			return nil, err
		}
		for _, inst := range group.GetInstances() {
			if inst.GetStatus() == statuspb.NoCloudStatus_DEL {
				continue
			}
			if err := instancesController.SetStatus(ctx, inst, statuspb.NoCloudStatus_SUS); err != nil {
				return nil, err
			}
		}
	}

	event := elpb.Event{
		Entity:    schema.SERVICES_COL,
		Uuid:      service.GetUuid(),
		Scope:     "database",
		Action:    "suspend",
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	if err := s.ctrl.SetStatus(ctx, service, statuspb.NoCloudStatus_SUS); err != nil {
		log.Error("Error updating Service", zap.Error(err), zap.Any("service", service))
		event.Rc = 1
		nocloud.Log(log, &event)
		return nil, status.Error(codes.Internal, "Error storing updates")
	}

	event.Rc = 0
	nocloud.Log(log, &event)

	return &pb.SuspendResponse{}, nil
}

func (s *ServicesServer) Unsuspend(ctx context.Context, request *pb.UnsuspendRequest) (*pb.UnsuspendResponse, error) {
	log := s.log.Named("Unsuspend")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("request", request), zap.String("requestor", requestor))

	service, err := s.ctrl.Get(ctx, requestor, request.GetUuid())
	if err != nil {
		log.Debug("Error getting Service", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not found")
	}
	log.Debug("Found Service", zap.Any("service", service))

	if service.GetAccess().GetLevel() < access.Level_ROOT {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Service")
	}

	for _, group := range service.GetInstancesGroups() {
		groupController := s.ctrl.IGController()
		instancesController := groupController.Instances()
		if err := groupController.SetStatus(ctx, group, statuspb.NoCloudStatus_UP); err != nil {
			return nil, err
		}
		for _, inst := range group.GetInstances() {
			if inst.GetStatus() == statuspb.NoCloudStatus_DEL {
				continue
			}
			if err := instancesController.SetStatus(ctx, inst, statuspb.NoCloudStatus_UP); err != nil {
				return nil, err
			}
		}
	}

	event := elpb.Event{
		Entity:    schema.SERVICES_COL,
		Uuid:      service.GetUuid(),
		Scope:     "database",
		Action:    "unsuspend",
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	if err = s.ctrl.SetStatus(ctx, service, statuspb.NoCloudStatus_UP); err != nil {
		log.Error("Error updating Service", zap.Error(err), zap.Any("service", service))
		event.Rc = 1
		nocloud.Log(log, &event)
		return nil, status.Error(codes.Internal, "Error storing updates")
	}

	event.Rc = 0
	nocloud.Log(log, &event)

	return &pb.UnsuspendResponse{}, nil
}

func (s *ServicesServer) Down(ctx context.Context, request *pb.DownRequest) (*pb.DownResponse, error) {
	log := s.log.Named("Down")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("request", request), zap.String("requestor", requestor))

	service, err := s.ctrl.Get(ctx, requestor, request.GetUuid())
	if err != nil {
		log.Debug("Error getting Service", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not found")
	}
	log.Debug("Found Service", zap.Any("service", service))

	if service.GetAccess().GetLevel() < access.Level_MGMT {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Service")
	}

	if service.GetStatus() == statuspb.NoCloudStatus_SUS {
		return nil, status.Error(codes.PermissionDenied, "Can't undeploy suspended service")
	}

	contexts := make(map[string]*InstancesGroupDriverContext)

	for _, group := range service.GetInstancesGroups() {
		if group.Sp == nil || group.GetSp() == "" {
			log.Debug("Group is unprovisioned, skipping", zap.String("group", group.GetUuid()))
			continue
		}

		sp, err := s.sp_ctrl.Get(ctx, group.GetSp())
		if err != nil {
			log.Error("Error getting ServiceProvider", zap.Error(err), zap.String("id", group.GetSp()))
			return nil, status.Errorf(codes.InvalidArgument, "Error getting ServiceProvider(%s)", group.GetSp())
		}

		groupType := group.GetType()
		client, ok := s.drivers[groupType]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "Driver of type '%s' not registered", groupType)
		}

		contexts[group.GetUuid()] = &InstancesGroupDriverContext{sp, &client}
	}

	err = s.ctrl.SetStatus(ctx, service, statuspb.NoCloudStatus_PROC)
	if err != nil {
		log.Error("Error updating Service", zap.Error(err), zap.Any("service", service))
		return nil, status.Error(codes.Internal, "Error storing updates")
	}

	for i, group := range service.GetInstancesGroups() {
		c, ok := contexts[group.GetUuid()]
		if !ok {
			log.Debug("Instance Group has no context, i.e. provision", zap.String("group", group.GetUuid()), zap.String("service", service.GetUuid()))
			continue
		}
		client := *c.client
		sp := c.sp

		group.Status = statuspb.NoCloudStatus_INIT

		res, err := client.Down(ctx, &driverpb.DownRequest{Group: group, ServicesProvider: sp.ServicesProvider})
		if err != nil {
			log.Error("Error undeploying group", zap.Any("service_provider", sp), zap.Any("group", group), zap.Error(err))
			continue
		}

		err = s.ctrl.IGController().SetStatus(ctx, group, statuspb.NoCloudStatus_INIT)
		if err != nil {
			log.Error("Error updating InstancesGroup", zap.Error(err), zap.Any("IG", group))
		}

		group := res.GetGroup()
		// err = s.ctrl.Unprovide(ctx, group.GetUuid())
		// if err != nil {
		// 	log.Error("Error unlinking group from ServiceProvider", zap.Any("service_provider", sp.GetUuid()), zap.Any("group", group), zap.Error(err))
		// 	continue
		// }
		service.InstancesGroups[i] = group
	}

	event := elpb.Event{
		Entity:    schema.SERVICES_COL,
		Uuid:      service.GetUuid(),
		Scope:     "database",
		Action:    "down",
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	err = s.ctrl.SetStatus(ctx, service, statuspb.NoCloudStatus_INIT)
	if err != nil {
		log.Error("Error updating Service", zap.Error(err), zap.Any("service", service))
		event.Rc = 1
		nocloud.Log(log, &event)
		return nil, status.Error(codes.Internal, "Error storing updates")
	}

	event.Rc = 0
	nocloud.Log(log, &event)

	return &pb.DownResponse{}, nil
}

func (s *ServicesServer) Get(ctx context.Context, request *pb.GetRequest) (res *pb.Service, err error) {
	log := s.log.Named("Get")
	log.Debug("Request received", zap.Any("request", request))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	service, err := s.ctrl.Get(ctx, requestor, request.GetUuid())
	if err != nil {
		log.Debug("Error getting Service from DB", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not Found in DB")
	}

	if service.GetAccess().GetLevel() < access.Level_MGMT {
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	return service, nil
}

func (s *ServicesServer) List(ctx context.Context, request *pb.ListRequest) (response *pb.Services, err error) {
	log := s.log.Named("List")
	log.Debug("Request received", zap.String("namespace", request.GetNamespace()), zap.Bool("show_deleted", request.GetShowDeleted()))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	r, err := s.ctrl.List(ctx, requestor, request)
	if err != nil {
		log.Debug("Error reading Services from DB", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error reading Services from DB")
	}

	return &pb.Services{Pool: r.Result, Count: int64(r.Count)}, nil
}

func (s *ServicesServer) Delete(ctx context.Context, request *pb.DeleteRequest) (response *pb.DeleteResponse, err error) {
	log := s.log.Named("Delete")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("request", request), zap.String("requestor", requestor))

	service, err := s.ctrl.Get(ctx, requestor, request.GetUuid())
	if err != nil {
		log.Debug("Error getting Service from DB", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Service not Found in DB")
	}

	if service.GetAccess().GetLevel() < access.Level_ADMIN {
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	err = s.ctrl.Delete(ctx, service)
	if err != nil {
		log.Error("Error Deleting Service", zap.Error(err))
		return &pb.DeleteResponse{Result: false, Error: err.Error()}, nil
	}

	return &pb.DeleteResponse{Result: true}, nil
}

func (s *ServicesServer) Stream(req *pb.StreamRequest, srv pb.ServicesService_StreamServer) (err error) {
	log := s.log.Named("stream")
	log.Debug("Request received", zap.Any("req", req))
	requestor := srv.Context().Value(nocloud.NoCloudAccount).(string)

	messages := make(chan interface{}, 10)

	if service, err := s.ctrl.Get(srv.Context(), requestor, req.GetUuid()); err != nil || service.GetAccess().GetLevel() < access.Level_READ {
		log.Warn("Failed access check", zap.String("uuid", req.GetUuid()))
		return errors.New("failed access check")
	}

	uuids, err := s.ctrl.GetServiceInstancesUuids(req.GetUuid())
	if err != nil {
		log.Error("Couldn't find service", zap.Any("uuid", req.GetUuid()), zap.Error(err))
		return err
	}

	topics := make([]string, len(uuids))
	for i, id := range uuids {
		topics[i] = "instance/" + id
	}

	s.log.Debug("topics", zap.Any("topics", topics))

	s.ps.AddSub(messages, topics...)
	defer unsub(s.ps, messages)

	for msg := range messages {
		state := msg.(*spb.ObjectState)
		log.Debug("state", zap.Any("state", state))
		err := srv.Send(state)
		if err != nil {
			log.Warn("Unable to send message", zap.Error(err))
			break
		}
	}

	return nil
}

func unsub[T chan any](ps *pubsub.PubSub, ch chan any) {
	go ps.Unsub(ch)

	for range ch {
	}
}
