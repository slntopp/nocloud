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
	sppb "github.com/slntopp/nocloud/pkg/services_providers/proto"
	"go.uber.org/zap"
)

type spRegistryAPI struct {
	client sppb.ServicesProvidersServiceClient
	apipb.UnimplementedServicesProvidersServiceServer

	log *zap.Logger
}

func (sp *spRegistryAPI) mustEmbedUnimplementedServicesProviderServiceServer() {
	sp.log.Info("Method missing")
}

func (sp *spRegistryAPI) Test(ctx context.Context, req *sppb.ServicesProvider) (*sppb.TestResponse, error) {
	sp.log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	return sp.client.Test(ctx, req)
}