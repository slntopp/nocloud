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

package edge

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"

	pb "github.com/slntopp/nocloud-proto/edge"
	stpb "github.com/slntopp/nocloud-proto/states"
	"github.com/slntopp/nocloud/pkg/nocloud"
	s "github.com/slntopp/nocloud/pkg/states"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EdgeServiceServer struct {
	pb.UnimplementedEdgeServiceServer

	log *zap.Logger
	pub s.Pub
	db  driver.Database
}

func NewEdgeServiceServer(log *zap.Logger, db driver.Database, rbmq rabbitmq.Connection) *EdgeServiceServer {
	s := s.NewStatesPubSub(log, nil, rbmq)
	ch := s.Channel()
	s.TopicExchange(ch, "states")

	return &EdgeServiceServer{
		log: log, pub: s.Publisher(ch, "states", "instances"), db: db,
	}
}

func (s *EdgeServiceServer) Test(ctx context.Context, _ *pb.TestRequest) (*pb.TestResponse, error) {
	return &pb.TestResponse{Result: true}, nil
}

func (s *EdgeServiceServer) PostState(ctx context.Context, req *stpb.ObjectState) (*pb.Empty, error) {
	inst, ok := ctx.Value(nocloud.NoCloudInstance).(string)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Post State to Instance")
	}
	req.Uuid = inst

	if req.State == nil {
		return nil, status.Error(codes.InvalidArgument, "State is nil")
	}

	s.log.Debug("Publishing state", zap.String("instance", inst), zap.String("state", req.GetState().String()))
	_, err := s.pub(req)
	return &pb.Empty{}, err
}
