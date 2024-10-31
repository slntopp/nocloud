package billing

import (
	"connectrpc.com/connect"
	"context"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing/descriptions"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DescriptionsServer struct {
	log *zap.Logger

	db driver.Database

	descriptions graph.DescriptionsController
	nss          graph.NamespacesController
}

func NewDescriptionsServer(logger *zap.Logger, db driver.Database) *DescriptionsServer {
	log := logger.Named("DescriptionsServer")
	return &DescriptionsServer{
		log:          log,
		db:           db,
		descriptions: graph.NewDescriptionsController(log, db),
		nss:          graph.NewNamespacesController(log.Named("DescriptionsCtrl"), db),
	}
}

func (s *DescriptionsServer) Create(ctx context.Context, r *connect.Request[pb.Description]) (*connect.Response[pb.Description], error) {
	log := s.log.Named("Create")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Descriptions")
	}

	description, err := s.descriptions.Create(ctx, r.Msg)
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(description), nil
}

func (s *DescriptionsServer) Update(ctx context.Context, r *connect.Request[pb.Description]) (*connect.Response[pb.Description], error) {
	log := s.log.Named("Update")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Descriptions")
	}

	description, err := s.descriptions.Update(ctx, r.Msg)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(description), nil
}

func (s *DescriptionsServer) Get(ctx context.Context, r *connect.Request[pb.Description]) (*connect.Response[pb.Description], error) {
	log := s.log.Named("Get")

	description, err := s.descriptions.Get(ctx, r.Msg.GetUuid())
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(description), nil
}

func (s *DescriptionsServer) List(ctx context.Context, r *connect.Request[pb.ListDescriptionsRequest]) (*connect.Response[pb.ListDescriptionsResponse], error) {
	log := s.log.Named("List")

	descriptions, err := s.descriptions.List(ctx, r.Msg)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.ListDescriptionsResponse{Descriptions: descriptions}), nil
}

func (s *DescriptionsServer) Count(ctx context.Context, r *connect.Request[pb.CountDescriptionsRequest]) (*connect.Response[pb.CountDescriptionsResponse], error) {
	log := s.log.Named("List")

	descriptions, err := s.descriptions.Count(ctx)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.CountDescriptionsResponse{Total: int64(len(descriptions))}), nil
}

func (s *DescriptionsServer) Delete(ctx context.Context, r *connect.Request[pb.Description]) (*connect.Response[pb.Description], error) {
	log := s.log.Named("Create")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Descriptions")
	}

	err := s.descriptions.Delete(ctx, r.Msg.GetUuid())
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(r.Msg), nil
}
