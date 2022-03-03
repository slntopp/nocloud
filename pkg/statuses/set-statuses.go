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
	"fmt"

	redis "github.com/go-redis/redis/v8"
	pb "github.com/slntopp/nocloud/pkg/instances/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const KEYS_PREFIX = "_st"

type StatusesServer struct {
	pb.UnimplementedStatesServiceServer

	log *zap.Logger
	rdb *redis.Client
}

func NewStatusesServer(log *zap.Logger, rdb *redis.Client) *StatusesServer {
	return &StatusesServer{
		log: log.Named("StatusesServer"), rdb: rdb,
	}
}

// Store Instance State to Redis key and channel
func (s *StatusesServer) PostInstanceState(
	ctx context.Context,
	req *pb.PostInstanceStateRequest,
) (*pb.PostInstanceStateResponse, error) {

	state := req.GetState()
	key := fmt.Sprintf("%s:%s", KEYS_PREFIX, req.GetUuid())
	json, err := json.Marshal(state.GetMeta())
	if err != nil {
		s.log.Error("Error Marshal JSON",
			zap.String("key", key), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error  Marshal JSON")
	}

	r := s.rdb.Set(ctx, key, json, 0)
	_, err = r.Result()
	if err != nil {
		s.log.Error("Error putting status to Redis",
			zap.String("key", key), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error putting status to Redis")
	}

	err = s.rdb.Publish(ctx, key, json).Err()
	if err != nil {
		s.log.Error("Error putting status to Redis channel",
			zap.String("key", key), zap.Error(err))
	}

	return &pb.PostInstanceStateResponse{Uuid: req.Uuid, Result: 0, Error: ""}, nil
}
