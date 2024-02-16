package billing

import (
	"connectrpc.com/connect"
	"context"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing/addons"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AddonsServer struct {
	log *zap.Logger

	db driver.Database

	addons *graph.AddonsController
	nss    graph.NamespacesController
}

func NewAddonsServer(logger *zap.Logger, db driver.Database) *AddonsServer {
	log := logger.Named("AddonsServer")
	return &AddonsServer{
		log:    log,
		db:     db,
		addons: graph.NewAddonsController(log, db),
		nss:    graph.NewNamespacesController(log.Named("NamespacesController"), db),
	}
}

func (s *AddonsServer) Create(ctx context.Context, r *connect.Request[pb.Addon]) (*connect.Response[pb.Addon], error) {
	log := s.log.Named("Create")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	if !graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Addons")
	}

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

	if !graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
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

func (s *AddonsServer) Get(ctx context.Context, r *connect.Request[pb.Addon]) (*connect.Response[pb.Addon], error) {
	log := s.log.Named("Get")

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	addon, err := s.addons.Get(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	if !graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) && !addon.GetPublic() {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Addons")
	}

	resp := connect.NewResponse(addon)

	return resp, nil
}

func (s *AddonsServer) List(ctx context.Context, r *connect.Request[pb.ListAddonsRequest]) (*connect.Response[pb.ListAddonsResponse], error) {
	log := s.log.Named("List")

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	addons, err := s.addons.List(ctx, req.GetGroup())
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}
	if graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		resp := connect.NewResponse(&pb.ListAddonsResponse{Addons: addons})
		return resp, nil
	}

	var filteredAddons []*pb.Addon

	for _, val := range filteredAddons {
		if val.GetPublic() {
			filteredAddons = append(filteredAddons, val)
		}
	}
	resp := connect.NewResponse(&pb.ListAddonsResponse{Addons: filteredAddons})
	return resp, nil
}

func (s *AddonsServer) Delete(ctx context.Context, r *connect.Request[pb.Addon]) (*connect.Response[pb.Addon], error) {
	log := s.log.Named("Create")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg

	if !graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Addons")
	}

	err := s.addons.Delete(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	return nil, nil
}