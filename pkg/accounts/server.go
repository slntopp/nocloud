package accounts

import (
	"context"

	"github.com/slntopp/nocloud/pkg/accounts/accountspb"
	"go.uber.org/zap"
)

type AccountsServiceServer struct {
	accountspb.UnimplementedAccountsServiceServer

	log *zap.Logger
}

func NewServer(log *zap.Logger) *AccountsServiceServer {
	return &AccountsServiceServer{log: log}
}

func (s *AccountsServiceServer) Token(ctx context.Context, request *accountspb.TokenRequest) (*accountspb.TokenResponse, error) {
	log := s.log.Named("Token")

	log.Debug("Token request received", zap.Any("request", request))

	return &accountspb.TokenResponse{Token: "sometoken"}, nil
}