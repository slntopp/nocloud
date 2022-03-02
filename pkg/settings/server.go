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
	"fmt"
	"strings"

	redis "github.com/go-redis/redis/v8"
	"github.com/slntopp/nocloud/pkg/nocloud"
	pb "github.com/slntopp/nocloud/pkg/settings/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	strcase "github.com/stoewer/go-strcase"

	"go.uber.org/zap"
)

const KEYS_PREFIX = "_settings"
const KEYS_DESC_POSTFIX = "desc"
const KEYS_VISIBILITY_POSTFIX = "pub"

var KEY_NS_PATTERN = fmt.Sprintf("%s:*[^:%s|:%s]", KEYS_PREFIX, KEYS_VISIBILITY_POSTFIX, KEYS_DESC_POSTFIX)

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
		keys[i] = fmt.Sprintf("%s:%s", KEYS_PREFIX, key)
	}
	
	response, err := s.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		s.log.Error("Error getting data from Redis", zap.Strings("keys", keys), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting data from Redis")
	}
	
	result := make(map[string]interface{})
	for i, key := range req.GetKeys() {
		result[strcase.KebabCase(key)] = response[i]
	}

	res, err := structpb.NewStruct(result)
	if err != nil {
		s.log.Error("Error serializing map to Struct", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error serializing map to Struct")
	}
	return res, nil
}

func (s *SettingsServiceServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	access := ctx.Value(nocloud.NoCloudRootAccess).(int32)
	if access < 3 {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights")
	}

	key := fmt.Sprintf("%s:%s", KEYS_PREFIX, strcase.LowerCamelCase(req.GetKey()))
	request := map[string]interface{}{
		key: req.GetValue(),
		fmt.Sprintf("%s:%s", key, KEYS_VISIBILITY_POSTFIX): req.GetPublic(),
	}
	if req.GetDescription() != "" {
		request[fmt.Sprintf("%s:%s", key, KEYS_DESC_POSTFIX)] = req.GetDescription()
	}

	r := s.rdb.MSet(ctx, request)
	_, err := r.Result()
	if err != nil {
		s.log.Error("Error allocating keys in Redis", zap.String("key", key), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error allocating keys in Redis")
	}

	go func() {
		s.rdb.Publish(ctx, key, req.GetValue())
	}()

	return &pb.PutResponse{Key: req.GetKey()}, nil
}

func (s *SettingsServiceServer) Sub(req *pb.SubRequest, stream pb.SettingsService_SubServer) error {
	sub := s.rdb.Subscribe(context.TODO(), strcase.LowerCamelCase(req.GetKey()))
	defer sub.Close()
	for {
		select {
		case msg := <- sub.Channel():
			stream.Send(&pb.SubRequest{
				Key: msg.Channel,
				Value: msg.Payload,
			})
		case <- stream.Context().Done():
			return nil
		}
	}
}

func (s *SettingsServiceServer) Keys(ctx context.Context, _ *pb.KeysRequest) (*pb.KeysResponse, error) {
	r := s.rdb.Keys(ctx, KEY_NS_PATTERN)
	keys, err := r.Result()
	if err != nil {
		s.log.Error("Error getting keys from Redis", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting keys from Redis")
	}
	keys_l := len(keys)
	if keys_l == 0 {
		return &pb.KeysResponse{}, nil
	}
	keys_all := make([]string, keys_l * 2)
	for i, key := range keys {
		keys_all[i] = fmt.Sprintf("%s:%s", key, KEYS_VISIBILITY_POSTFIX)
		keys_all[i + keys_l] = fmt.Sprintf("%s:%s", key, KEYS_DESC_POSTFIX)
	}
	vals, err := s.rdb.MGet(ctx, keys_all...).Result()
	if err != nil {
		s.log.Error("Error getting values from Redis", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting values from Redis")
	}

	result := make([]*pb.KeysResponse_Key, keys_l)
	for i, key := range keys {
		key = strings.SplitN(key, ":", 2)[1]
		result[i] = &pb.KeysResponse_Key{
			Key: strcase.KebabCase(key),
		}
		if pub, ok := vals[i].(string); ok {
			result[i].Public = pub == "1"
		}
		if desc, ok := vals[i + keys_l].(string); ok {
			result[i].Description = desc
		}
	}
	return &pb.KeysResponse{Pool: result}, nil
}