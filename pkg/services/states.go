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
package services

import (
	"context"

	"github.com/slntopp/nocloud/pkg/nocloud"
	pb "github.com/slntopp/nocloud/pkg/services/proto"
	stpb "github.com/slntopp/nocloud/pkg/states/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//Gets states Instanses of Servce from pkg/states
func (s *ServicesServer) GetStates(ctx context.Context, request *pb.GetStatesRequest) (*stpb.GetStatesResponse, error) {
	service, err := s.Get(ctx, &pb.GetRequest{
		Uuid: request.Uuid,
	})
	if err != nil {
		return nil, err
	}

	return s.GetStatesInternal(ctx, service)
}


func (s *ServicesServer) GetStatesInternal(ctx context.Context, service *pb.Service) (*stpb.GetStatesResponse, error) {
	var keys []string
	for _, igroup := range service.GetInstancesGroups() {
		for _, inst := range igroup.GetInstances() {
			keys = append(keys, inst.GetUuid())
		}
	}
	resp, err := s.states.GetStates(ctx, &stpb.GetStatesRequest{
		Uuids: keys,
	})
	return resp, err
}

func (s *ServicesServer) PostState(ctx context.Context, request *stpb.PostStateRequest) (*stpb.PostStateResponse, error) {
	inst, ok := ctx.Value(nocloud.NoCloudInstance).(string)
	if !ok || inst != request.GetUuid() {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Post State to Instance")
	}
	return s.states.PostState(ctx, request)
}