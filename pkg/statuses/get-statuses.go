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
package statuses

import (
	"context"
	"fmt"

	pb "github.com/slntopp/nocloud/pkg/instances/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

//Gets statuses Instanses of Servce
func (s *StatusesServer) GetInstancesStates(ctx context.Context, req *pb.GetInstancesStatesRequest) (resp *pb.GetInstancesStatesResponse, err error) {

	resp = &pb.GetInstancesStatesResponse{
		States:  make(map[string]*pb.InstanceState),
	}

	for _, instance := range req.GetInstances() {

			instance_uuid := instance.GetUuid()
			key := fmt.Sprintf("%s:%s", KEYS_PREFIX, instance.GetUuid())

			r := s.rdb.Get(ctx, key)
			st, err := r.Result()
			if err != nil {
				s.log.Error("Error getting status from Redis",
					zap.String("key", key), zap.Error(err))
				return nil, status.Error(codes.Internal, "Error getting status from Redis")
			}

			var stpb structpb.Value
			err = stpb.UnmarshalJSON([]byte(st))
			if err != nil {
				s.log.Error("Error Unmarshal JSON",
					zap.String("key", key), zap.Error(err))
				return nil, status.Error(codes.Internal, "Error  Unmarshal JSON")
			}

			resp.States[instance_uuid] = &pb.InstanceState{
				State: int32(stpb.GetStructValue().GetFields()["state"].GetNumberValue()),
				Meta: stpb.GetStructValue().Fields,
			}
	}

	return resp, nil
}
