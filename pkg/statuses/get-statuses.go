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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	pb "github.com/slntopp/nocloud/pkg/instances/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//Gets statuses Instanses of Servce
func (s *StatusesServer) GetInstancesStates(ctx context.Context, req *pb.GetInstancesStatesRequest) (resp *pb.GetInstancesStatesResponse, err error) {

	resp = &pb.GetInstancesStatesResponse{
		States:  make(map[string]*pb.InstanceState),
	}

	keys := req.GetInstances()	
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
		var istate pb.InstanceState
		switch state.(type) {
		case string:
			err = json.Unmarshal([]byte(state.(string)), &istate)
		case nil:
			err = errors.New("No Data")
		}
		if err != nil {
			s.log.Error("Error Unmarshal JSON",
				zap.String("key", keys[i]), zap.Error(err))
			return nil, status.Error(codes.Internal, "Error Unmarshal JSON")
		}

		key := strings.Replace(keys[i], KEYS_PREFIX + ":", "", 1)
		resp.States[key] = &istate
	}

	return resp, nil
}
