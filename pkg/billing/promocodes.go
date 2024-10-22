package billing

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing/promocodes"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type PromocodesServer struct {
	log *zap.Logger

	db driver.Database

	promos *graph.PromocodesController
	nss    graph.NamespacesController
}

func NewPromocodesServer(logger *zap.Logger, db driver.Database) *PromocodesServer {
	log := logger.Named("PromocodesServer")
	return &PromocodesServer{
		log:    log,
		db:     db,
		promos: graph.NewPromocodesController(log, db),
		nss:    graph.NewNamespacesController(log.Named("PromocodesCtrl"), db),
	}
}

func parseEntryResource(resource string) (*pb.EntryResource, error) {
	res := &pb.EntryResource{}
	parts := strings.Split(resource, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid resource: %s", resource)
	}
	if parts[1] == "" {
		return nil, fmt.Errorf("resource id cannot be empty: %s", resource)
	}
	if strings.ToLower(parts[0]) == "invoices" {
		res.Invoice = &parts[1]
	}
	if strings.ToLower(parts[0]) == "instances" {
		res.Instance = &parts[1]
	}
	return res, nil
}

func (s *PromocodesServer) Create(ctx context.Context, r *connect.Request[pb.Promocode]) (*connect.Response[pb.Promocode], error) {
	log := s.log.Named("Create")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Descriptions")
	}

	promo, err := s.promos.Create(ctx, r.Msg)
	if err != nil {
		log.Error("Failed to create promocode", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(promo), nil
}

func (s *PromocodesServer) Update(ctx context.Context, r *connect.Request[pb.Promocode]) (*connect.Response[pb.Promocode], error) {
	log := s.log.Named("Update")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Descriptions")
	}

	promo, err := s.promos.Update(ctx, r.Msg)
	if err != nil {
		log.Error("Failed to update promocode", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(promo), nil
}

func (s *PromocodesServer) Get(ctx context.Context, r *connect.Request[pb.Promocode]) (*connect.Response[pb.Promocode], error) {
	log := s.log.Named("Get")

	promo, err := s.promos.Get(ctx, r.Msg.GetUuid())
	if err != nil {
		log.Error("Failed to get promocode", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(promo), nil
}

func (s *PromocodesServer) GetByCode(ctx context.Context, r *connect.Request[pb.GetPromocodeByCodeRequest]) (*connect.Response[pb.Promocode], error) {
	log := s.log.Named("GetByCode")

	promo, err := s.promos.GetByCode(ctx, r.Msg.GetCode())
	if err != nil {
		log.Error("Failed to get promocode by code", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(promo), nil
}

func (s *PromocodesServer) Apply(ctx context.Context, r *connect.Request[pb.ApplyPromocodeRequest]) (*connect.Response[pb.ApplyPromocodeResponse], error) {
	log := s.log.Named("Apply")

	promo, err := s.promos.GetByCode(ctx, r.Msg.GetCode())
	if err != nil {
		log.Error("Failed to get promocode by code", zap.Error(err))
		return nil, err
	}

	entry, err := parseEntryResource(r.Msg.GetResource())
	if err != nil {
		log.Error("Failed to parse promocode resource", zap.Error(err))
		return nil, err
	}

	err = s.promos.AddEntry(ctx, promo.GetUuid(), entry)
	if err != nil {
		log.Error("Failed to apply promocode", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.ApplyPromocodeResponse{Success: true}), nil
}

func (s *PromocodesServer) List(ctx context.Context, r *connect.Request[pb.ListPromocodesRequest]) (*connect.Response[pb.ListPromocodesResponse], error) {
	log := s.log.Named("List")

	promocodes, err := s.promos.List(ctx, r.Msg)
	if err != nil {
		log.Error("Failed to list promocodes", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.ListPromocodesResponse{Promocodes: promocodes}), nil
}

func (s *PromocodesServer) Count(ctx context.Context, r *connect.Request[pb.CountPromocodesRequest]) (*connect.Response[pb.CountPromocodesResponse], error) {
	log := s.log.Named("Count")

	promocodes, err := s.promos.Count(ctx)
	if err != nil {
		log.Error("Failed to count promocodes", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.CountPromocodesResponse{Total: int64(len(promocodes))}), nil
}

func (s *PromocodesServer) Delete(ctx context.Context, r *connect.Request[pb.Promocode]) (*connect.Response[pb.Promocode], error) {
	log := s.log.Named("Delete")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Descriptions")
	}

	err := s.promos.Delete(ctx, r.Msg.GetUuid())
	if err != nil {
		log.Error("Failed to delete promocode", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(r.Msg), nil
}
