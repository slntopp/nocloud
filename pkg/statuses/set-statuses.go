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
var rdbCtx context.Context = context.Background()

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
func (s *StatusesServer) PostState(
	ctx context.Context,
	req *pb.PostStateRequest,
) (*pb.PostStateResponse, error) {
	log := s.log.Named("PostState")
	log.Debug("Request received", zap.Any("request", req))

	state := req.GetState()
	key := fmt.Sprintf("%s:%s", KEYS_PREFIX, req.GetUuid())
	json, err := json.Marshal(state)
	if err != nil {
		s.log.Error("Error Marshal JSON",
			zap.String("key", key), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error  Marshal JSON")
	}

	log.Debug("Storing State in Redis", zap.String("key", key))
	r := s.rdb.Set(rdbCtx, key, json, 0)
	_, err = r.Result()
	if err != nil {
		s.log.Error("Error putting status to Redis",
			zap.String("key", key), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error putting status to Redis")
	}
	log.Debug("State stored in Redis")

	go func() {
		log.Debug("Storing State in Redis Channel", zap.String("key", key))
		err = s.rdb.Publish(rdbCtx, key, json).Err()
		if err != nil {
			s.log.Error("Error putting status to Redis channel", zap.String("key", key), zap.Error(err))
			return
		}
		log.Debug("State stored in Redis Channel")
	}()

	return &pb.PostStateResponse{Uuid: req.Uuid, Result: 0, Error: ""}, nil
}
