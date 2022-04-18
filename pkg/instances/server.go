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
package instances

import (
	"context"

	"github.com/arangodb/go-driver"
	driverpb "github.com/slntopp/nocloud/pkg/drivers/instance/vanilla"
	"github.com/slntopp/nocloud/pkg/graph"
	pb "github.com/slntopp/nocloud/pkg/instances/proto"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/access"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InstancesServer struct {
	pb.UnimplementedInstancesServiceServer
	log *zap.Logger
	
	ctrl *graph.InstancesController
	ig_ctrl *graph.InstancesGroupsController
	
	drivers  map[string]driverpb.DriverServiceClient
	
	db driver.Database
}

func NewInstancesServiceServer(logger *zap.Logger, db driver.Database) *InstancesServer {
	ig_ctrl := graph.NewInstancesGroupsController(logger, db) 
	return &InstancesServer{
		db: db, log: logger.Named("instances"),
		ctrl: ig_ctrl.Instances(),
		ig_ctrl: ig_ctrl,
		drivers:  make(map[string]driverpb.DriverServiceClient),
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
	var instance pb.Instance
	err := graph.GetWithAccess(
		ctx, s.db,
		driver.NewDocumentID(schema.ACCOUNTS_COL, requestor),
		instance_id, &instance,
	)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	if instance.AccessLevel == nil || *instance.AccessLevel < access.MGMT {
		log.Error("Access denied")
		return nil, status.Error(codes.PermissionDenied, "Access denied")
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
		Instance: &instance,
		ServicesProvider: r.SP,
		Method: req.Method,
		Params: req.Params,
	})
}