package showcase_categories

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
	"google.golang.org/protobuf/types/known/structpb"
)

type CategoriesServer struct {
	log *zap.Logger

	sppb.UnimplementedShowcaseCategoriesServiceServer

	ctrl    graph.CategoriesController
	ns_ctrl graph.NamespacesController
	ca      graph.CommonActionsController
	db      driver.Database
}

func NewCategoriesServer(log *zap.Logger, db driver.Database) *CategoriesServer {
	return &CategoriesServer{
		log:     log.Named("ShowcaseCategoriesServer"),
		ctrl:    graph.NewCategoriesController(log, db),
		ns_ctrl: graph.NewNamespacesController(log, db),
		ca:      graph.NewCommonActionsController(log, db),
		db:      db,
	}
}

func (s *CategoriesServer) Create(ctx context.Context, req *sppb.ShowcaseCategory) (*sppb.ShowcaseCategory, error) {
	log := s.log.Named("Create")
	log.Debug("Create request received", zap.Any("request", req))

	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requester", zap.String("id", requester))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Invoke")
	}

	cat, err := s.ctrl.Create(ctx, req)

	return cat, err
}

func (s *CategoriesServer) Update(ctx context.Context, req *sppb.ShowcaseCategory) (*sppb.ShowcaseCategory, error) {
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

func (s *CategoriesServer) Get(ctx context.Context, req *sppb.GetRequest) (*sppb.ShowcaseCategory, error) {
	log := s.log.Named("Get")
	log.Debug("Get request received", zap.Any("request", req))
	requester, _ := ctx.Value(nocloud.NoCloudAccount).(string)

	cat, err := s.ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		return cat, err
	}

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	if !cat.GetPublic() && !s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT) {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Get")
	}

	return cat, err
}

func (s *CategoriesServer) List(ctx context.Context, req *sppb.ListRequest) (*sppb.ShowcaseCategories, error) {
	log := s.log.Named("List")
	log.Debug("List request received")

	var requester string
	if !req.Anonymously {
		requester = ctx.Value(nocloud.NoCloudAccount).(string)
	}

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	isRoot := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT)

	categories, err := s.ctrl.List(ctx, requester, isRoot, req)
	if req.GetOmitPromos() {
		for _, sc := range categories {
			sc.Promo = map[string]*structpb.Value{}
		}
	}

	result := &sppb.ShowcaseCategories{Categories: categories}

	return result, err
}

func (s *CategoriesServer) Delete(ctx context.Context, req *sppb.DeleteRequest) (*sppb.DeleteResponse, error) {
	log := s.log.Named("Update")
	log.Debug("Update request received", zap.Any("request", req))

	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requester", zap.String("id", requester))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := s.ca.HasAccess(ctx, requester, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to perform Delete")
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
