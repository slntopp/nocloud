package billing

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing/addons"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type AddonsServer struct {
	log *zap.Logger

	db driver.Database

	addons graph.AddonsController
	nss    graph.NamespacesController
	ca     graph.CommonActionsController
}

func NewAddonsServer(logger *zap.Logger, db driver.Database) *AddonsServer {
	log := logger.Named("AddonsServer")
	return &AddonsServer{
		log:    log,
		db:     db,
		addons: graph.NewAddonsController(log, db),
		nss:    graph.NewNamespacesController(log.Named("NamespacesController"), db),
		ca:     graph.NewCommonActionsController(log, db),
	}
}

func (s *AddonsServer) Create(ctx context.Context, r *connect.Request[pb.Addon]) (*connect.Response[pb.Addon], error) {
	log := s.log.Named("Create")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	if req.GetKind() == pb.Kind_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "kind is required")
	}

	if !s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Addons")
	}

	req.Created = time.Now().Unix()

	addon, err := s.addons.Create(ctx, req)
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	resp := connect.NewResponse(addon)

	return resp, nil
}

func (s *AddonsServer) Update(ctx context.Context, r *connect.Request[pb.Addon]) (*connect.Response[pb.Addon], error) {
	log := s.log.Named("Update")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	if req.GetKind() == pb.Kind_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "kind is required")
	}

	if !s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Addons")
	}

	addon, err := s.addons.Update(ctx, req)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return nil, err
	}

	resp := connect.NewResponse(addon)

	return resp, nil
}

func (s *AddonsServer) CreateBulk(ctx context.Context, r *connect.Request[pb.BulkAddons]) (*connect.Response[pb.BulkAddons], error) {
	log := s.log.Named("Create")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	if !s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Addons")
	}

	for _, addon := range req.GetAddons() {
		if addon.GetKind() == pb.Kind_UNSPECIFIED {
			return nil, status.Error(codes.InvalidArgument, "kind is required")
		}

		addon.Created = time.Now().Unix()
	}

	addons, err := s.addons.CreateBulk(ctx, req.GetAddons())
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	resp := connect.NewResponse(&pb.BulkAddons{Addons: addons})

	return resp, nil
}

func (s *AddonsServer) UpdateBulk(ctx context.Context, r *connect.Request[pb.BulkAddons]) (*connect.Response[pb.BulkAddons], error) {
	log := s.log.Named("Update")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	if !s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Addons")
	}

	for _, addon := range req.GetAddons() {
		if addon.GetKind() == pb.Kind_UNSPECIFIED {
			return nil, status.Error(codes.InvalidArgument, "kind is required")
		}
	}

	addons, err := s.addons.UpdateBulk(ctx, req.GetAddons())
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return nil, err
	}

	resp := connect.NewResponse(&pb.BulkAddons{Addons: addons})

	return resp, nil
}

func (s *AddonsServer) Get(ctx context.Context, r *connect.Request[pb.Addon]) (*connect.Response[pb.Addon], error) {
	log := s.log.Named("Get")

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	addon, err := s.addons.Get(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	if !s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) && !addon.GetPublic() {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Addons")
	}

	resp := connect.NewResponse(addon)

	return resp, nil
}

func (s *AddonsServer) List(ctx context.Context, r *connect.Request[pb.ListAddonsRequest]) (*connect.Response[pb.ListAddonsResponse], error) {
	log := s.log.Named("List")

	requestor, _ := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	addons, err := s.addons.List(ctx, req)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}
	if s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		resp := connect.NewResponse(&pb.ListAddonsResponse{Addons: addons})
		return resp, nil
	}

	var filteredAddons []*pb.Addon

	for _, val := range addons {
		if val.GetPublic() {
			filteredAddons = append(filteredAddons, val)
		}
	}
	resp := connect.NewResponse(&pb.ListAddonsResponse{Addons: filteredAddons})
	return resp, nil
}

func (s *AddonsServer) Count(ctx context.Context, r *connect.Request[pb.CountAddonsRequest]) (*connect.Response[pb.CountAddonsResponse], error) {
	log := s.log.Named("List")

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	addons, err := s.addons.Count(ctx, req)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	unique, err := s.addons.GetUnique(ctx)
	if err != nil {
		return nil, err
	}

	value, err := structpb.NewValue(unique)
	if err != nil {
		return nil, err
	}

	if s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		resp := connect.NewResponse(&pb.CountAddonsResponse{Total: int64(len(addons)), Unique: value})
		return resp, nil
	}

	var filteredAddons []*pb.Addon

	for _, val := range addons {
		if val.GetPublic() {
			filteredAddons = append(filteredAddons, val)
		}
	}

	resp := connect.NewResponse(&pb.CountAddonsResponse{Total: int64(len(filteredAddons)), Unique: value})
	return resp, nil
}

func (s *AddonsServer) Delete(ctx context.Context, r *connect.Request[pb.Addon]) (*connect.Response[pb.Addon], error) {
	log := s.log.Named("Create")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	if !s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Addons")
	}

	err := s.addons.Delete(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(req), nil
}
