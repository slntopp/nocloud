package main

import (
	"context"

	"github.com/slntopp/nocloud/pkg/accounts/accountspb"
	"github.com/slntopp/nocloud/pkg/api/apipb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountsAPI struct {
	client accountspb.AccountsServiceClient
	apipb.UnimplementedAccountsServiceServer
}

func (acc *accountsAPI) mustEmbedUnimplementedAccountsServiceServer() {
	log.Info("Method missing")
}

func (acc *accountsAPI) Token(ctx context.Context, request *accountspb.TokenRequest) (*accountspb.TokenResponse, error) {
	res, err := acc.client.Token(ctx, request)
	if err != nil {
		log.Error("Generating Token failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "Wrong credentials given")
	}
	return res, nil
}