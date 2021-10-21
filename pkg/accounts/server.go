package accounts

import (
	"context"

	"github.com/arangodb/go-driver"

	"github.com/slntopp/nocloud/pkg/accounts/accountspb"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"

	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

func ValidateMetadata(ctx context.Context, log *zap.Logger) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Error("Failed to get metadata from context")
		return ctx, status.Error(codes.Aborted, "Failed to get metadata from context")
	}

	//Check for Authentication
	requestor := md.Get(nocloud.NOCLOUD_ACCOUNT_CLAIM)
	if requestor == nil {
		log.Error("Failed to authenticate account")
		return ctx, status.Error(codes.Unauthenticated, "Not authenticated")
	}
	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, requestor[0])

	return ctx, nil
}
