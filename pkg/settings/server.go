/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

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
	pb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	strcase "github.com/stoewer/go-strcase"

	"go.uber.org/zap"
)

const KEYS_PREFIX = "_settings"

var KEY_NS_PATTERN = fmt.Sprintf("%s:*", KEYS_PREFIX)

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

func (s *SettingsServiceServer) Get(ctx context.Context, req *pb.GetRequest) (res *structpb.Struct, err error) {
	log := s.log.Named("Get")
	log.Debug("Request received", zap.Strings("keys", req.GetKeys()))

	result := make(map[string]interface{})

	for _, key := range req.GetKeys() {
		dbKey := fmt.Sprintf("%s:%s", KEYS_PREFIX, strcase.LowerCamelCase(key))
		log.Debug("Reading hash", zap.String("key", dbKey))
		r := s.rdb.HGet(ctx, dbKey, "value")
		result[key], err = r.Result()
		log.Debug("Result", zap.Any("value", result[key]), zap.Error(err))
	}

	res, err = structpb.NewStruct(result)
	if err != nil {
		log.Error("Error serializing map to Struct", zap.Error(err))
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
	r := s.rdb.HSet(ctx, key, "value", req.GetValue(),
		"desc", req.GetDescription(), "pub", req.GetPublic())
	_, err := r.Result()
	if err != nil {
		s.log.Error("Error allocating keys in Redis", zap.String("key", key), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error allocating keys in Redis")
	}

	return &pb.PutResponse{Key: req.GetKey()}, nil
}

func (s *SettingsServiceServer) Keys(ctx context.Context, _ *pb.KeysRequest) (*pb.KeysResponse, error) {
	r := s.rdb.Keys(ctx, KEY_NS_PATTERN)
	keys, err := r.Result()
	if err != nil {
		s.log.Error("Error getting keys from Redis", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting keys from Redis")
	}

	result := make([]*pb.KeysResponse_Key, len(keys))
	for i, key := range keys {
		r := s.rdb.HGetAll(ctx, key)
		data, err := r.Result()

		key = strings.Replace(key, KEYS_PREFIX+":", "", 1)
		key = strcase.KebabCase(key)

		if err != nil {
			result[i] = &pb.KeysResponse_Key{Key: key, Description: "Unresolved"}
			continue
		}
		result[i] = &pb.KeysResponse_Key{
			Key:         key,
			Description: data["desc"],
			Public:      data["pub"] == "1",
		}
	}

	return &pb.KeysResponse{Pool: result}, nil
}

func (s *SettingsServiceServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	key := fmt.Sprintf("%s:%s", KEYS_PREFIX, strcase.LowerCamelCase(req.GetKey()))

	_, err := s.rdb.Del(ctx, key).Result()
	if err != nil {
		s.log.Error("Error deleting key from Redis", zap.String("key", key), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error deleting key from Redis")
	}

	return &pb.DeleteResponse{Key: req.GetKey()}, nil
}

func (s *SettingsServiceServer) Sub(req *pb.GetRequest, srv pb.SettingsService_SubServer) error {
	log := s.log.Named("Sub")
	log.Debug("Request received", zap.Strings("keys", req.GetKeys()))

	keys := make([]string, len(req.GetKeys()))

	tmpl := fmt.Sprintf("__keyspace@%d__:%s:", s.rdb.Options().DB, KEYS_PREFIX) + "%s"

	for i, key := range req.GetKeys() {
		keys[i] = fmt.Sprintf(tmpl, strcase.LowerCamelCase(key))
	}

	log.Debug("Subscribing to", zap.Strings("keys", keys))
	r := s.rdb.Subscribe(srv.Context(), keys...)
	defer r.Close()

	ch := r.Channel()

	for msg := range ch {
		log.Info("Message Received", zap.String("channel", msg.Channel), zap.String("payload", msg.Payload))

		var key string
		_, err := fmt.Sscanf(msg.Channel, tmpl, &key)
		if err != nil {
			log.Warn("Couldn't Sscanf setting Key", zap.Error(err))
			continue
		}

		key = strcase.KebabCase(key)

		err = srv.Send(&pb.KeyEvent{
			Key: key, Event: msg.Payload,
		})
		if err != nil {
			log.Warn("Couldn't send Event, closing stream", zap.Error(err))
			return nil
		}
	}

	log.Debug("Stream closed")
	return nil
}
