package accounts

import (
	"context"

	"github.com/arangodb/go-driver"

	"github.com/slntopp/nocloud/pkg/accounts/accountspb"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"

	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jwt "github.com/dgrijalva/jwt-go"
)

type AccountsServiceServer struct {
	accountspb.UnimplementedAccountsServiceServer
	db driver.Database
	ctrl graph.AccountsController
	ns_ctrl graph.NamespacesController

	log *zap.Logger
	SIGNING_KEY []byte
}

func NewServer(log *zap.Logger, db driver.Database) *AccountsServiceServer {
	accountsCol, _ := db.Collection(nil, graph.ACCOUNTS_COL)
	credCol, _ := db.Collection(nil, graph.CREDENTIALS_COL)
	nsCol, _ := db.Collection(nil, graph.NAMESPACES_COL)

	return &AccountsServiceServer{
		log: log, db: db, 
		ctrl: graph.NewAccountsController(
			log.Named("AccountsController"), accountsCol, credCol,
		),
		ns_ctrl: graph.NewNamespacesController(
			log.Named("NamespacesController"), nsCol,
		),
	}
}
}

func (s *AccountsServiceServer) Token(ctx context.Context, request *accountspb.TokenRequest) (*accountspb.TokenResponse, error) {
	log := s.log.Named("Token")

	log.Debug("Token request received", zap.Any("request", request))
	account, ok := s.ctrl.Authorize(ctx, request.Auth.Type, request.Auth.Data...)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Wrong credentials given")
	}
	log.Debug("Authorized user", zap.String("ID", account.ID.String()))

	claims := jwt.MapClaims{}
	claims[nocloud.NOCLOUD_ACCOUNT_CLAIM] = account.Key
	claims["exp"] = request.Exp

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token_string, err := token.SignedString(s.SIGNING_KEY)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to issue token")
	}

	return &accountspb.TokenResponse{Token: token_string}, nil
}