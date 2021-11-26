/*
Copyright © 2021 Nikita Ivanovski info@slnt-opp.xyz

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
package main

import (
	"context"

	"github.com/slntopp/nocloud/pkg/api/apipb"
	"github.com/slntopp/nocloud/pkg/nocloud"
	pb "github.com/slntopp/nocloud/pkg/services/proto"
	"go.uber.org/zap"
)

type servicesAPI struct {
	client pb.ServicesServiceClient
	apipb.UnimplementedServicesServiceServer

	log *zap.Logger
}

func (sp *servicesAPI) mustEmbedUnimplementedServicesServiceServer() {
	sp.log.Info("Method missing")
}

func (sp *servicesAPI) TestConfig(ctx context.Context, req *pb.CreateRequest) (*pb.TestConfigResponse, error) {
	sp.log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	return sp.client.TestConfig(ctx, req)
}

func (sp *servicesAPI) Create(ctx context.Context, req *pb.CreateRequest) (*pb.Service, error) {
	sp.log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	return sp.client.Create(ctx, req)
}

func (sp *servicesAPI) Up(ctx context.Context, req *pb.UpRequest) (*pb.UpResponse, error) {
	sp.log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	return sp.client.Up(ctx, req)
}

func (sp *servicesAPI) Get(ctx context.Context, req *pb.GetRequest) (*pb.Service, error) {
	sp.log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	return sp.client.Get(ctx, req)
}

func (sp *servicesAPI) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	sp.log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	return sp.client.List(ctx, req)
}