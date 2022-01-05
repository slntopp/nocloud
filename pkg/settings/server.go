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
package settings

import (
	"context"

	redis "github.com/go-redis/redis/v8"
	pb "github.com/slntopp/nocloud/pkg/settings/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"go.uber.org/zap"
)

const KEYS_PREFIX = "settings"
const KEYS_DESC_POSTFIX = "desc"
const KEYS_VISIBILITY_POSTFIX = "pub"

type SettingsServiceServer struct {
	pb.UnimplementedSettingsServiceServer

	log *zap.Logger
	rdb *redis.Client
}

func NewSettingsServer(log *zap.Logger, rdb *redis.Client) *SettingsServiceServer {
	return &SettingsServiceServer{
		log: log.Named("SettingsServer"), rdb: rdb,
	}
}

func (s *SettingsServiceServer) Get(ctx context.Context, req *pb.GetRequest) (*structpb.Struct, error) {
	keys := make([]string, len(req.GetKeys()))
	for i, key := range req.GetKeys() {
		keys[i] = KEYS_PREFIX + ":" + key
	}
	
	response, err := s.rdb.MGet(ctx, req.GetKeys()...).Result()
	if err != nil {
		s.log.Error("Error getting data from Redis", zap.Strings("keys", keys), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting data from Redis")
	}

	result := make(map[string]interface{})
	for i, key := range req.GetKeys() {
		result[key] = response[i]
	}

	res, err := structpb.NewStruct(result)
	if err != nil {
		s.log.Error("Error serializing map to Struct", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error serializing map to Struct")
	}
	return res, nil
}

func (s *SettingsServiceServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	key := KEYS_PREFIX + ":" + req.GetKey()
	request := map[string]interface{}{
		key: req.GetValue(),
		key + ":" + KEYS_VISIBILITY_POSTFIX: req.GetPublic(),
	}
	if req.GetDescription() != "" {
		request[key + ":" + KEYS_DESC_POSTFIX] = req.GetDescription()
	}

	r := s.rdb.MSet(ctx, request)
	_, err := r.Result()
	if err != nil {
		s.log.Error("Error allocating keys in Redis", zap.String("key", key), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error allocating keys in Redis")
	}
	return &pb.PutResponse{Key: req.GetKey()}, nil
}