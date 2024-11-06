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
	"slices"
	"strings"
	"time"
)

type PromocodesServer struct {
	log *zap.Logger

	db driver.Database

	promos graph.PromocodesController
	nss    graph.NamespacesController
	ca     graph.CommonActionsController
}

func NewPromocodesServer(logger *zap.Logger, db driver.Database,
	promocodes graph.PromocodesController, nss graph.NamespacesController, ca graph.CommonActionsController) *PromocodesServer {
	log := logger.Named("PromocodesServer")
	return &PromocodesServer{
		log:    log,
		db:     db,
		promos: promocodes,
		nss:    nss,
		ca:     ca,
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
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage promocodes")
	}

	r.Msg.Created = time.Now().Unix()
	r.Msg.Uses = make([]*pb.EntryResource, 0)
	promo, err := s.promos.Create(ctx, r.Msg)
	if err != nil {
		log.Error("Failed to create promocode", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(promo), nil
}

func (s *PromocodesServer) Update(ctx context.Context, r *connect.Request[pb.Promocode]) (*connect.Response[pb.Promocode], error) {
	log := s.log.Named("Update")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage promocodes")
	}

	r.Msg.Uses = nil
	promo, err := s.promos.Update(ctx, r.Msg)
	if err != nil {
		log.Error("Failed to update promocode", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(promo), nil
}

func (s *PromocodesServer) Get(ctx context.Context, r *connect.Request[pb.Promocode]) (*connect.Response[pb.Promocode], error) {
	log := s.log.Named("Get")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage promocodes")
	}

	promo, err := s.promos.Get(ctx, r.Msg.GetUuid())
	if err != nil {
		log.Error("Failed to get promocode", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(promo), nil
}

func (s *PromocodesServer) GetByCode(ctx context.Context, r *connect.Request[pb.GetPromocodeByCodeRequest]) (*connect.Response[pb.Promocode], error) {
	log := s.log.Named("GetByCode")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	isAdmin := s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN)

	promo, err := s.promos.GetByCode(ctx, r.Msg.GetCode())
	if err != nil {
		log.Error("Failed to get promocode by code", zap.Error(err))
		return nil, fmt.Errorf("promocode not found")
	}

	if promo.Status == pb.PromocodeStatus_DELETED && !isAdmin {
		return nil, fmt.Errorf("promocode not found")
	}

	if !isAdmin {
		promo.Uses = nil
		promo.Limit = 0
		promo.Created = 0
		promo.UsesPerUser = 0
	}

	return connect.NewResponse(promo), nil
}

func (s *PromocodesServer) Apply(ctx context.Context, r *connect.Request[pb.ApplyPromocodeRequest]) (*connect.Response[pb.ApplyPromocodeResponse], error) {
	log := s.log.Named("Apply")

	promo, err := s.promos.GetByCode(ctx, r.Msg.GetCode())
	if err != nil {
		log.Error("Failed to get promocode by code", zap.Error(err))
		return nil, fmt.Errorf("cannot apply promocode. Promocode not found")
	}

	if promo.Status == pb.PromocodeStatus_DELETED || promo.Status == pb.PromocodeStatus_SUSPENDED {
		return nil, fmt.Errorf("cannot apply promocode. Promocode is not exists or currently inactive")
	}

	entry, err := parseEntryResource(r.Msg.GetResource())
	if err != nil {
		log.Error("Failed to parse promocode resource", zap.Error(err))
		return nil, fmt.Errorf("cannot apply promocode. Invalid request body")
	}

	entry.Account = ""
	err = s.promos.AddEntry(ctx, promo.GetUuid(), entry)
	if err != nil {
		log.Error("Failed to apply promocode", zap.Error(err))
		return nil, fmt.Errorf("cannot apply promocode. %s", err.Error())
	}

	return connect.NewResponse(&pb.ApplyPromocodeResponse{Success: true}), nil
}

func (s *PromocodesServer) Detach(ctx context.Context, r *connect.Request[pb.DetachPromocodeRequest]) (*connect.Response[pb.DetachPromocodeResponse], error) {
	log := s.log.Named("Detach")
	req := r.Msg
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage promocodes")
	}

	promo, err := s.promos.Get(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to get promocode", zap.Error(err))
		return nil, err
	}

	entry, err := parseEntryResource(req.GetResource())
	if err != nil {
		log.Error("Failed to parse promocode resource", zap.Error(err))
		return nil, err
	}

	err = s.promos.RemoveEntry(ctx, promo.GetUuid(), entry)
	if err != nil {
		log.Error("Failed to detach promocode", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.DetachPromocodeResponse{Success: true}), nil
}

func (s *PromocodesServer) List(ctx context.Context, r *connect.Request[pb.ListPromocodesRequest]) (*connect.Response[pb.ListPromocodesResponse], error) {
	log := s.log.Named("List")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	// TODO: maybe refactor somehow
	entryRes := make([]*pb.EntryResource, 0)

	isAdmin := s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN)
	if !isAdmin {
		resources := r.Msg.GetFilters()["resources"].GetListValue().AsSlice()
		if len(resources) == 0 {
			return connect.NewResponse(&pb.ListPromocodesResponse{Promocodes: make([]*pb.Promocode, 0)}), nil
		}
		for _, res := range resources {
			resStr, ok := res.(string)
			if !ok {
				log.Error("Failed to parse promocode resource. Not a string")
				return nil, status.Error(codes.InvalidArgument, "Failed to parse promocode resource. Not a string")
			}
			entry, err := parseEntryResource(resStr)
			if err != nil {
				log.Error("Failed to parse promocode resource", zap.Error(err))
				return nil, err
			}
			if entry.Invoice != nil {
				if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.INVOICES_COL, entry.GetInvoice()), access.Level_ADMIN) {
					return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to get by requested resource")
				}
			}
			if entry.Instance != nil {
				if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.INSTANCES_COL, entry.GetInstance()), access.Level_ADMIN) {
					return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to get by requested resource")
				}
			}
			entryRes = append(entryRes, entry)
		}
	}

	promocodes, err := s.promos.List(ctx, r.Msg)
	if err != nil {
		log.Error("Failed to list promocodes", zap.Error(err))
		return nil, err
	}

	if !isAdmin {
		for _, promo := range promocodes {
			filtetedUses := make([]*pb.EntryResource, 0)
			promo.Limit = 0
			promo.Created = 0
			promo.UsesPerUser = 0
			// Filter to show only uses associated by request user resources
			for _, entry := range entryRes {
				if slices.ContainsFunc(promo.Uses, func(e *pb.EntryResource) bool {
					if e.Instance != nil && entry.Instance != nil && *e.Instance == *entry.Instance {
						return true
					}
					if e.Invoice != nil && entry.Invoice != nil && *e.Invoice == *entry.Invoice {
						return true
					}
					return false
				}) {
					filtetedUses = append(filtetedUses, entry)
				}
			}
			promo.Uses = filtetedUses
		}
	}

	return connect.NewResponse(&pb.ListPromocodesResponse{Promocodes: promocodes}), nil
}

func (s *PromocodesServer) Count(ctx context.Context, r *connect.Request[pb.CountPromocodesRequest]) (*connect.Response[pb.CountPromocodesResponse], error) {
	log := s.log.Named("Count")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	isAdmin := s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN)
	if !isAdmin {
		resources := r.Msg.GetFilters()["resources"].GetListValue().AsSlice()
		if len(resources) == 0 {
			return connect.NewResponse(&pb.CountPromocodesResponse{Total: 0}), nil
		}
		for _, res := range r.Msg.GetFilters()["resources"].GetListValue().AsSlice() {
			resStr, ok := res.(string)
			if !ok {
				log.Error("Failed to parse promocode resource. Resource is not a string")
				return nil, status.Error(codes.InvalidArgument, "Failed to parse promocode resource. Not a string")
			}
			entry, err := parseEntryResource(resStr)
			if err != nil {
				log.Error("Failed to parse promocode resource", zap.Error(err))
				return nil, err
			}
			if entry.Invoice != nil {
				if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.INVOICES_COL, entry.GetInvoice()), access.Level_ADMIN) {
					return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to get by requested resource")
				}
			}
			if entry.Instance != nil {
				if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.INSTANCES_COL, entry.GetInstance()), access.Level_ADMIN) {
					return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to get by requested resource")
				}
			}
		}
	}

	count, err := s.promos.Count(ctx, r.Msg)
	if err != nil {
		log.Error("Failed to count promocodes", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.CountPromocodesResponse{Total: count}), nil
}

func (s *PromocodesServer) Delete(ctx context.Context, r *connect.Request[pb.Promocode]) (*connect.Response[pb.Promocode], error) {
	log := s.log.Named("Delete")
	requester := ctx.Value(nocloud.NoCloudAccount).(string)

	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage promocodes")
	}

	err := s.promos.Delete(ctx, r.Msg.GetUuid())
	if err != nil {
		log.Error("Failed to delete promocode", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(r.Msg), nil
}
