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
package dns

import (
	"context"
	"encoding/json"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"strings"

	pb "github.com/slntopp/nocloud-proto/dns"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const KEYS_PREFIX = "_dns"

type DNSServer struct {
	pb.UnimplementedDNSServer

	log *zap.Logger
	rdb redisdb.Client
}

func NewDNSServer(log *zap.Logger, rdb redisdb.Client) *DNSServer {
	return &DNSServer{
		log: log.Named("DNSServer"), rdb: rdb,
	}
}

func (s *DNSServer) Get(ctx context.Context, req *pb.Zone) (*pb.Zone, error) {
	domain := req.GetName()
	r := s.rdb.HGetAll(ctx, KEYS_PREFIX+":"+domain)
	records, err := r.Result()
	if err != nil {
		s.log.Error("Error getting records from Redis", zap.String("zone", KEYS_PREFIX+":"+req.GetName()), zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "Error getting records from Redis")
	}
	locations := make(map[string]*pb.Record)
	for loc, data := range records {
		var record pb.Record
		err := json.Unmarshal([]byte(data), &record)
		if err != nil {
			continue
		}
		locations[loc] = &record
	}

	req.Locations = locations
	return req, nil
}

func (s *DNSServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	r := s.rdb.Keys(ctx, KEYS_PREFIX+":*")
	keys, err := r.Result()
	if err != nil {
		s.log.Error("Error getting keys from Redis", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting keys from Redis")
	}
	for i, key := range keys {
		keys[i] = strings.Split(key, ":")[1]
	}
	return &pb.ListResponse{Zones: keys}, nil
}

func (s *DNSServer) Put(ctx context.Context, req *pb.Zone) (*pb.Result, error) {
	access := ctx.Value(nocloud.NoCloudRootAccess).(int)
	if access < 3 {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights")
	}

	data := make(map[string]interface{})
	for key, record := range req.GetLocations() {
		s.log.Debug("record", zap.String("string", record.String()))
		r, err := json.Marshal(record)
		if err != nil {
			s.log.Error("Error Marshaling Record", zap.Any("record", record), zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "Error Marshaling Record")
		}
		data[key] = r
		s.log.Debug("record result", zap.ByteString("string", r))
	}
	r := s.rdb.HSet(ctx, KEYS_PREFIX+":"+req.GetName(), data)
	upd, err := r.Result()
	if err != nil {
		s.log.Error("Error putting hash to Redis", zap.String("zone", KEYS_PREFIX+":"+req.GetName()), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error putting hash to Redis")
	}
	return &pb.Result{Result: upd}, nil
}

func (s *DNSServer) Delete(ctx context.Context, req *pb.Zone) (*pb.Result, error) {
	access := ctx.Value(nocloud.NoCloudRootAccess).(int)
	if access < 3 {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights")
	}

	r := s.rdb.Del(ctx, KEYS_PREFIX+":"+req.GetName())
	res, err := r.Result()
	if err != nil {
		s.log.Error("Error deleting zone in Redis", zap.String("zone", KEYS_PREFIX+":"+req.GetName()), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error deleting zone in Redis")
	}
	return &pb.Result{Result: res}, nil
}
