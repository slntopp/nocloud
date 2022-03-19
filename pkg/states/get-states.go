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
package states

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	pb "github.com/slntopp/nocloud/pkg/states/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//Gets states Instanses of Servce
func (s *StatesServer) GetStates(ctx context.Context, req *pb.GetStatesRequest) (resp *pb.GetStatesResponse, err error) {

	resp = &pb.GetStatesResponse{
		States:  make(map[string]*pb.State),
	}

	keys := req.GetUuids()	
	for i, uuid := range keys {
		keys[i] = fmt.Sprintf("%s:%s", KEYS_PREFIX, uuid)
	}

	r := s.rdb.MGet(ctx, keys...)
	states, err := r.Result()
	if err != nil {
		s.log.Error("Error getting states from Redis", zap.Strings("keys", keys), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting states from Redis")
	}

	for i, state := range states {
		var istate pb.State
		key := strings.Replace(keys[i], KEYS_PREFIX + ":", "", 1)

		switch state.(type) {
		case string:
			err = json.Unmarshal([]byte(state.(string)), &istate)
		case nil:
			states := make(map[string]*pb.State)
			states[key] = &pb.State{
				State: pb.NoCloudState_UNKNOWN,
			}
			return &pb.GetStatesResponse{
				States: states,
			}, nil
		}
		if err != nil {
			s.log.Error("Error Unmarshal JSON",
				zap.String("key", keys[i]), zap.Error(err))
			return nil, status.Error(codes.Internal, "Error Unmarshal JSON")
		}

		resp.States[key] = &istate
	}

	return resp, nil
}
