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
	sppb "github.com/slntopp/nocloud/pkg/services_providers/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type spRegistryAPI struct {
	client sppb.ServicesProvidersServiceClient
	apipb.UnimplementedServicesProvidersServiceServer

	log *zap.Logger
}

func (sp *spRegistryAPI) mustEmbedUnimplementedServicesProvidersServiceServer() {}

func (sp *spRegistryAPI) Test(ctx context.Context, req *sppb.ServicesProvider) (*sppb.TestResponse, error) {
	return sp.client.Test(ctx, req)
}

func (sp *spRegistryAPI) Create(ctx context.Context, req *sppb.ServicesProvider) (*sppb.ServicesProvider, error) {
	return sp.client.Create(ctx, req)
}

func (sp *spRegistryAPI) Get(ctx context.Context, req *sppb.GetRequest) (*sppb.ServicesProvider, error) {
	return sp.client.Get(ctx, req)
}

func (sp *spRegistryAPI) List(ctx context.Context, req *sppb.ListRequest) (*sppb.ListResponse, error) {
	return sp.client.List(ctx, req)
}

func (sp *spRegistryAPI) Invoke(ctx context.Context, req *sppb.ActionRequest) (*structpb.Struct, error) {
	return sp.client.Invoke(ctx, req)
}