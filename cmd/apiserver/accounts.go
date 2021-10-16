package main

import (
	"context"

	"github.com/slntopp/nocloud/pkg/accounts/accountspb"
	"github.com/slntopp/nocloud/pkg/api/apipb"
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