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
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"strings"

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
	rdb redisdb.Client
}

func NewSettingsServer(log *zap.Logger, rdb redisdb.Client) *SettingsServiceServer {
	return &SettingsServiceServer{
		log: log.Named("SettingsServer"), rdb: rdb,
	}
}

func _GetRootClaim(ctx context.Context) int {
	lvl, ok := ctx.Value(nocloud.NoCloudRootAccess).(int)
	if !ok {
		return 0
	}

	return lvl
}

func (s *SettingsServiceServer) Get(ctx context.Context, req *pb.GetRequest) (res *structpb.Struct, err error) {
	level := _GetRootClaim(ctx)

	log := s.log.Named("Get")
	log.Debug("Request received", zap.Strings("keys", req.GetKeys()), zap.Int("level", level))

	result := make(map[string]interface{})

	for _, key := range req.GetKeys() {
		dbKey := fmt.Sprintf("%s:%s", KEYS_PREFIX, strcase.LowerCamelCase(key))

		r := s.rdb.HGetAll(ctx, dbKey)
		data, err := r.Result()
		if err != nil {
			log.Warn("Coudln't get Key", zap.Error(err))
			continue
		}

		public := false
		if data["public"] == "true" || data["public"] == "1" {
			public = true
		}
		if !public && level < 3 {
			continue
		}

		result[key] = data["value"]

		log.Debug("Result", zap.Any("value", result[key]))
	}

	res, err = structpb.NewStruct(result)
	if err != nil {
		log.Error("Error serializing map to Struct", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error serializing map to Struct")
	}
	return res, nil
}

func (s *SettingsServiceServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	log := s.log.Named("Put")

	access := _GetRootClaim(ctx)
	if access < 3 {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights")
	}

	key := fmt.Sprintf("%s:%s", KEYS_PREFIX, strcase.LowerCamelCase(req.GetKey()))
	log.Debug("Put request received", zap.String("key", key))

	r := s.rdb.HSet(ctx, key, "value", req.GetValue(),
		"desc", req.GetDescription(), "public", req.GetPublic())
	_, err := r.Result()
	if err != nil {
		log.Error("Error allocating keys in Redis", zap.String("key", key), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error allocating keys in Redis")
	}

	return &pb.PutResponse{Key: req.GetKey()}, nil
}

func (s *SettingsServiceServer) Keys(ctx context.Context, _ *pb.KeysRequest) (*pb.KeysResponse, error) {
	level := _GetRootClaim(ctx)

	r := s.rdb.Keys(ctx, KEY_NS_PATTERN)
	keys, err := r.Result()
	if err != nil {
		s.log.Error("Error getting keys from Redis", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting keys from Redis")
	}

	var result []*pb.KeysResponse_Key
	for _, key := range keys {
		r := s.rdb.HGetAll(ctx, key)
		data, err := r.Result()

		key = strings.Replace(key, KEYS_PREFIX+":", "", 1)
		key = strcase.KebabCase(key)

		if err != nil {
			result = append(result, &pb.KeysResponse_Key{
				Key: key, Description: "Unresolved",
			})
			continue
		}

		public := false
		if data["public"] == "true" || data["public"] == "1" {
			public = true
		}

		if !public && level < 3 {
			result = append(result, &pb.KeysResponse_Key{
				Key: key, Description: "Unresolved",
			})
			continue
		}

		result = append(result, &pb.KeysResponse_Key{
			Key:         key,
			Description: data["desc"],
			Public:      public,
		})
	}

	return &pb.KeysResponse{Pool: result}, nil
}

func (s *SettingsServiceServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {

	if _GetRootClaim(ctx) < 3 {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights")
	}

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

	res, err := s.Get(srv.Context(), req)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(res.AsMap()))
	for k := range res.AsMap() {
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		return status.Error(codes.InvalidArgument, "No keys available")
	}

	log.Debug("Filtered keys", zap.Strings("keys", keys))

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
