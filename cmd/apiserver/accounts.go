package main

import (
	"context"

	"github.com/slntopp/nocloud/pkg/accounting/accountspb"
	"github.com/slntopp/nocloud/pkg/api/apipb"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
)

type accountsAPI struct {
	client accountspb.AccountsServiceClient
	apipb.UnimplementedAccountsServiceServer
}

func (acc *accountsAPI) mustEmbedUnimplementedAccountsServiceServer() {
	log.Info("Method missing")
}

func (acc *accountsAPI) Token(ctx context.Context, request *accountspb.TokenRequest) (*accountspb.TokenResponse, error) {
	return acc.client.Token(ctx, request)
}

func (acc *accountsAPI) Create(ctx context.Context, request *accountspb.CreateRequest) (*accountspb.CreateResponse, error) {
	log.Debug("context", zap.Any("context", ctx), zap.String("account", ctx.Value(nocloud.NoCloudAccount).(string)))
	ctx = AddCrossServiceMetadata(ctx)
	return acc.client.Create(ctx, request)
}