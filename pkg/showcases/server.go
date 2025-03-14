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

	ctrl    graph.ShowcasesController
	ns_ctrl graph.NamespacesController
	ca      graph.CommonActionsController
	db      driver.Database
}

func NewShowcasesServer(log *zap.Logger, db driver.Database) *ShowcasesServer {
	return &ShowcasesServer{
		log:     log.Named("ShowcasesServer"),
		ctrl:    graph.NewShowcasesController(log, db),
		ns_ctrl: graph.NewNamespacesController(log, db),
		ca:      graph.NewCommonActionsController(log, db),
		db:      db,
	}
}

func emptyPromo(promo *sppb.LanguagePromo) {
	promo.Icons = nil
	promo.Location = nil
	promo.Locations = nil
	promo.Offer = nil
	promo.Preview = ""
	promo.Rewards = nil
	promo.Service = nil
}

func (s *ShowcasesServer) Create(ctx context.Context, req *sppb.Showcase) (*sppb.Showcase, error) {
	log := s.log.Named("Create")
	log.Debug("Create request received", zap.Any("request", req))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requestor, ns, access.Level_ROOT)
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

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	showcase, err := s.ctrl.Update(ctx, req)

	return showcase, err
}

func (s *ShowcasesServer) Get(ctx context.Context, req *sppb.GetRequest) (*sppb.Showcase, error) {
	log := s.log.Named("Get")
	log.Debug("Create request received", zap.Any("request", req))
	requester, _ := ctx.Value(nocloud.NoCloudAccount).(string)

	showcase, err := s.ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		return showcase, err
	}

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	if !showcase.GetPublic() && !s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT) {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Get")
	}

	return showcase, err
}

func (s *ShowcasesServer) List(ctx context.Context, req *sppb.ListRequest) (*sppb.Showcases, error) {
	log := s.log.Named("List")
	log.Debug("List request received")

	var requestor string
	if !req.Anonymously {
		requestor = ctx.Value(nocloud.NoCloudAccount).(string)
	}

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	isRoot := s.ca.HasAccess(ctx, requestor, ns, access.Level_ROOT)

	showcases, err := s.ctrl.List(ctx, requestor, isRoot, req)
	if req.GetOmitPromos() {
		for _, sc := range showcases {
			for _, promo := range sc.GetPromo() {
				if promo != nil {
					emptyPromo(promo)
				}
			}
		}
	}

	result := &sppb.Showcases{Showcases: showcases}

	return result, err
}

func (s *ShowcasesServer) Delete(ctx context.Context, req *sppb.DeleteRequest) (*sppb.DeleteResponse, error) {
	log := s.log.Named("Update")
	log.Debug("Update request received", zap.Any("request", req))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	err := s.ctrl.Delete(ctx, req.GetUuid())

	result := &sppb.DeleteResponse{
		Result: true,
	}

	if err != nil {
		result.Result = false
	}

	return result, err
}
