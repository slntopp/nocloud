package account_groups

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	rpb "github.com/slntopp/nocloud-proto/registry"
	accpb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountGroupsServer struct {
	log *zap.Logger

	rpb.UnimplementedAccountGroupsServiceServer

	ctrl    graph.AccountGroupsController
	ns_ctrl graph.NamespacesController
	ca      graph.CommonActionsController
	db      driver.Database
}

func NewAccountGroupsServer(log *zap.Logger, db driver.Database) *AccountGroupsServer {
	return &AccountGroupsServer{
		log:     log.Named("AccountGroupsServer"),
		ctrl:    graph.NewAccountGroupsController(log, db),
		ns_ctrl: graph.NewNamespacesController(log, db),
		ca:      graph.NewCommonActionsController(log, db),
		db:      db,
	}
}

func (s *AccountGroupsServer) Create(ctx context.Context, req *accpb.AccountGroup) (*accpb.AccountGroup, error) {
	log := s.log.Named("Create")
	log.Debug("Create request received", zap.Any("request", req))

	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requester", zap.String("id", requester))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Create")
	}

	cat, err := s.ctrl.Create(ctx, req)

	return cat, err
}

func (s *AccountGroupsServer) Update(ctx context.Context, req *accpb.AccountGroup) (*accpb.AccountGroup, error) {
	log := s.log.Named("Update")
	log.Debug("Update request received", zap.Any("request", req))

	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requester", zap.String("id", requester))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Update")
	}

	cat, err := s.ctrl.Update(ctx, req)

	return cat, err
}

func (s *AccountGroupsServer) Get(ctx context.Context, req *accpb.GetRequest) (*accpb.AccountGroup, error) {
	log := s.log.Named("Get")
	log.Debug("Get request received", zap.Any("request", req))
	_, _ = ctx.Value(nocloud.NoCloudAccount).(string)

	cat, err := s.ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		return cat, err
	}

	return cat, err
}

func (s *AccountGroupsServer) List(ctx context.Context, req *accpb.ListRequest) (*accpb.AccountGroupsListResponse, error) {
	log := s.log.Named("List")
	log.Debug("List request received")

	groups, err := s.ctrl.List(ctx, req)

	result := &accpb.AccountGroupsListResponse{Pool: groups}

	return result, err
}

func (s *AccountGroupsServer) Delete(ctx context.Context, req *accpb.DeleteRequest) (*accpb.DeleteResponse, error) {
	log := s.log.Named("Delete")
	log.Debug("Update request received", zap.Any("request", req))

	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requester", zap.String("id", requester))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Delete")
	}

	err := s.ctrl.Delete(ctx, req.GetUuid())
	result := &accpb.DeleteResponse{Result: true}
	return result, err
}
