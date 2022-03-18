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

package edge

import (
	"context"

	pb "github.com/slntopp/nocloud/pkg/edge/proto"
	"github.com/slntopp/nocloud/pkg/nocloud"
	stpb "github.com/slntopp/nocloud/pkg/states/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EdgeServiceServer struct {
	pb.UnimplementedEdgeServiceServer

	log *zap.Logger
	st stpb.StatesServiceClient
}

func NewEdgeServiceServer(log *zap.Logger, st stpb.StatesServiceClient) *EdgeServiceServer {
	return &EdgeServiceServer{
		log: log, st: st,
	}
}

func (s *EdgeServiceServer) Test(ctx context.Context, _ *pb.TestRequest) (*pb.TestResponse, error) {
	return &pb.TestResponse{Result: true}, nil
}

func (s *EdgeServiceServer) PostState(ctx context.Context, req *stpb.PostStateRequest) (*stpb.PostStateResponse, error) {
	inst, ok := ctx.Value(nocloud.NoCloudInstance).(string)
	if !ok || inst != req.GetUuid() {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Post State to Instance")
	}
	return s.st.PostState(ctx, req)
}