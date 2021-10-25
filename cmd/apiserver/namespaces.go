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

	"github.com/slntopp/nocloud/pkg/accounting/namespacespb"
	"github.com/slntopp/nocloud/pkg/api/apipb"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
)

type namespacesAPI struct {
	client namespacespb.NamespacesServiceClient
	apipb.UnimplementedNamespacesServiceServer
}

func (ns *namespacesAPI) mustEmbedUnimplementedNamespacesServiceServer() {
	log.Info("Method missing")
}

func (ns *namespacesAPI) Create(ctx context.Context, request *namespacespb.CreateRequest) (*namespacespb.CreateResponse, error) {
	log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	ctx = AddCrossServiceMetadata(ctx)
	return ns.client.Create(ctx, request)
}

func (ns *namespacesAPI) List(ctx context.Context, request *namespacespb.ListRequest) (*namespacespb.ListResponse, error) {
	log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	ctx = AddCrossServiceMetadata(ctx)
	return ns.client.List(ctx, request)
}

func (ns *namespacesAPI) Join(ctx context.Context, request *namespacespb.JoinRequest) (*namespacespb.JoinResponse, error) {
	log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	ctx = AddCrossServiceMetadata(ctx)
	return ns.client.Join(ctx, request)
}

func (ns *namespacesAPI) Link(ctx context.Context, request *namespacespb.LinkRequest) (*namespacespb.LinkResponse, error) {
	log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	ctx = AddCrossServiceMetadata(ctx)
	return ns.client.Link(ctx, request)
}