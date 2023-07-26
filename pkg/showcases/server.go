package showcases

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShowcasesServer struct {
	log *zap.Logger

	sppb.UnimplementedShowcasesServiceServer

	ctrl    *graph.ShowcasesController
	ns_ctrl graph.NamespacesController
	db      driver.Database
}

func NewShowcasesServer(log *zap.Logger, db driver.Database) *ShowcasesServer {
	return &ShowcasesServer{
		log:     log.Named("ShowcasesServer"),
		ctrl:    graph.NewShowcasesController(log, db),
		ns_ctrl: graph.NewNamespacesController(log, db),
		db:      db,
	}
}

func (s *ShowcasesServer) Create(ctx context.Context, req *sppb.Showcase) (*sppb.Showcase, error) {
	log := s.log.Named("Create")
	log.Debug("Create request received", zap.Any("request", req))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns, err := s.ns_ctrl.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	showcase, err := s.ctrl.Create(ctx, req)

	return showcase, err
}

func (s *ShowcasesServer) Update(ctx context.Context, req *sppb.Showcase) (*sppb.Showcase, error) {
	log := s.log.Named("Update")
	log.Debug("Update request received", zap.Any("request", req))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns, err := s.ns_ctrl.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	showcase, err := s.ctrl.Update(ctx, req)

	return showcase, err
}

func (s *ShowcasesServer) Get(ctx context.Context, req *sppb.GetRequest) (*sppb.Showcase, error) {
	log := s.log.Named("Get")
	log.Debug("Create request received", zap.Any("request", req))

	showcase, err := s.ctrl.Get(ctx, req.GetUuid())

	return showcase, err
}

func (s *ShowcasesServer) List(ctx context.Context, req *sppb.ListRequest) (*sppb.Showcases, error) {
	log := s.log.Named("List")
	log.Debug("List request received")

	showcases, err := s.ctrl.List(ctx)

	result := &sppb.Showcases{Showcases: showcases}

	return result, err
}

func (s *ShowcasesServer) Delete(ctx context.Context, req *sppb.DeleteRequest) (*sppb.DeleteResponse, error) {
	log := s.log.Named("Update")
	log.Debug("Update request received", zap.Any("request", req))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns, err := s.ns_ctrl.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	err = s.ctrl.Delete(ctx, req.GetUuid())

	result := &sppb.DeleteResponse{
		Result: true,
	}

	if err != nil {
		result.Result = false
	}

	return result, err
}
