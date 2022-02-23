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

	spb "github.com/slntopp/nocloud/pkg/services/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

//Gets statuses Instanses of Servce
func (s *StatusesServer) StateGet(ctx context.Context, req *spb.Service) (resp *spb.GetStatesResponse, err error) {

	resp = &spb.GetStatesResponse{
		States:  make(map[string]*spb.State),
	}

	for _, ig := range req.InstancesGroups {
		for in := range ig.Instances {

			instance := ig.Instances[in]
			instance_uuid := string(instance.Uuid)

			r := s.rdb.Get(ctx, KEYS_PREFIX+":"+instance_uuid)
			st, err := r.Result()
			if err != nil {
				s.log.Error("Error getting status from Redis",
					zap.String("zone", KEYS_PREFIX+":"+instance_uuid), zap.Error(err))
				return nil, status.Error(codes.Internal, "Error getting status from Redis")
			}

			var stpb structpb.Value
			err = stpb.UnmarshalJSON([]byte(st))
			if err != nil {
				s.log.Error("Error Unmarshal JSON",
					zap.String("zone", KEYS_PREFIX+":"+string(req.Uuid)), zap.Error(err))
				return nil, status.Error(codes.Internal, "Error  Unmarshal JSON")
			}

			resp.States[instance_uuid] = &spb.State{
				State: int32(stpb.GetStructValue().GetFields()["state"].GetNumberValue()),
				Meta: stpb.GetStructValue().Fields,
			}
		}

	}

	return resp, nil
}
