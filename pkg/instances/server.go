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
package instances

import (
	"context"
	"github.com/slntopp/nocloud-proto/states"

	"github.com/arangodb/go-driver"
	amqp "github.com/rabbitmq/amqp091-go"
	accesspb "github.com/slntopp/nocloud-proto/access"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	pb "github.com/slntopp/nocloud-proto/instances"
	spb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	s "github.com/slntopp/nocloud/pkg/states"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InstancesServer struct {
	pb.UnimplementedInstancesServiceServer
	log *zap.Logger

	ctrl    *graph.InstancesController
	ig_ctrl *graph.InstancesGroupsController

	drivers map[string]driverpb.DriverServiceClient

	db driver.Database
}

func NewInstancesServiceServer(logger *zap.Logger, db driver.Database, rbmq *amqp.Connection) *InstancesServer {
	log := logger.Named("instances")
	log.Debug("New Instances Server Creating")
	ig_ctrl := graph.NewInstancesGroupsController(logger, db)

	log.Debug("Setting up StatesPubSub")
	s := s.NewStatesPubSub(log, &db, rbmq)
	ch := s.Channel()
	log.Debug("initializing Exchange with name \"states\" of type \"topic\"")
	s.TopicExchange(ch, "states") // init Exchange with name "states" of type "topic"
	log.Debug("initializing Consumer queue of topic \"states.instances\"")
	s.StatesConsumerInit(ch, "states", "instances", schema.INSTANCES_COL) // init Consumer queue of topic "states.instances"

	log.Debug("Setting up PubSub")
	d := NewPubSub(log, &db, rbmq)
	ch = d.Channel()
	log.Debug("initializing Exchange with name \"datas\" of type \"topic\"")
	d.TopicExchange(ch, "datas") // init Exchange with name "datas" of type "topic"
	log.Debug("initializing Consumer queue of topic \"datas.instances\"")
	d.ConsumerInit(ch, "datas", "instances", schema.INSTANCES_COL) // init Consumer queue of topic "datas.instances"
	log.Debug("initializing Consumer queue of topic \"datas.instances-groups\"")
	d.ConsumerInit(ch, "datas", "instances-groups", schema.INSTANCES_GROUPS_COL) // init Consumer queue of topic "datas.instances-groups"

	return &InstancesServer{
		db: db, log: log,
		ctrl:    ig_ctrl.Instances(),
		ig_ctrl: ig_ctrl,
		drivers: make(map[string]driverpb.DriverServiceClient),
	}
}

func (s *InstancesServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

func (s *InstancesServer) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
	log := s.log.Named("invoke")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	instance_id := driver.NewDocumentID(schema.INSTANCES_COL, req.Uuid)
	var instance graph.Instance
	instance, err := graph.GetWithAccess[graph.Instance](
		ctx, s.db,
		instance_id,
	)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	if instance.GetState().GetState() == states.NoCloudState_SUSPENDED {
		log.Error("Machine is suspended. Functionality is limited", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.Unavailable, "Machine is suspended. Functionality is limited")
	}

	r, err := s.ctrl.GetGroup(ctx, instance_id.String())
	if err != nil {
		log.Error("Failed to get Group and ServicesProvider", zap.Error(err))
		return nil, err
	}

	client, ok := s.drivers[r.SP.Type]
	if !ok {
		log.Error("Failed to get driver", zap.String("type", r.SP.Type))
		return nil, status.Error(codes.NotFound, "Driver not found")
	}
	return client.Invoke(ctx, &driverpb.InvokeRequest{
		Instance:         instance.Instance,
		ServicesProvider: r.SP,
		Method:           req.Method,
		Params:           req.Params,
	})
}

func (s *InstancesServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	log := s.log.Named("delete")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	instance_id := driver.NewDocumentID(schema.INSTANCES_COL, req.Uuid)
	var instance graph.Instance
	instance, err := graph.GetWithAccess[graph.Instance](
		ctx, s.db,
		instance_id,
	)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	err = s.ctrl.SetStatus(ctx, instance.Instance, spb.NoCloudStatus_DEL)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteResponse{
		Result: true,
	}, nil
}

func (s *InstancesServer) TransferIG(ctx context.Context, req *pb.TransferIGRequest) (*pb.TransferIGResponse, error) {
	log := s.log.Named("transfer")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	igId := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, req.GetUuid())
	newSrvId := driver.NewDocumentID(schema.SERVICES_COL, req.GetService())
	ig, err := graph.GetWithAccess[graph.InstancesGroup](ctx, s.db, igId)

	if err != nil {
		log.Error("Failed to get instances group", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	srv, err := graph.GetWithAccess[graph.Service](ctx, s.db, newSrvId)

	if err != nil {
		log.Error("Failed to get service", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if ig.GetAccess().GetLevel() < accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	if srv.GetAccess().GetLevel() < accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	srvEdge, err := s.ig_ctrl.GetEdge(ctx, igId.String(), schema.SERVICES_COL)

	if err != nil {
		log.Error("Failed to get Service", zap.Error(err))
		return nil, err
	}

	err = s.ig_ctrl.TransferIG(ctx, srvEdge, newSrvId, igId)
	if err != nil {
		return nil, err
	}

	return &pb.TransferIGResponse{
		Result: true,
		Meta:   nil,
	}, nil
}

func (s *InstancesServer) TransferInstance(ctx context.Context, req *pb.TransferInstanceRequest) (*pb.TransferInstanceResponse, error) {
	log := s.log.Named("transfer")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	instanceId := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, req.Uuid)
	newIGId := driver.NewDocumentID(schema.SERVICES_COL, req.GetIg())
	ig, err := graph.GetWithAccess[graph.InstancesGroup](ctx, s.db, instanceId)

	if err != nil {
		log.Error("Failed to get instances group", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	srv, err := graph.GetWithAccess[graph.Service](ctx, s.db, newIGId)

	if err != nil {
		log.Error("Failed to get service", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if ig.GetAccess().GetLevel() < accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	if srv.GetAccess().GetLevel() < accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	igEdge, err := s.ctrl.GetEdge(ctx, instanceId.String(), schema.INSTANCES_GROUPS_COL)

	if err != nil {
		log.Error("Failed to get Service", zap.Error(err))
		return nil, err
	}

	err = s.ctrl.TransferInst(ctx, igEdge, newIGId, instanceId)
	if err != nil {
		return nil, err
	}

	return &pb.TransferInstanceResponse{
		Result: true,
		Meta:   nil,
	}, nil
}
