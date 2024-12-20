package billing

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	bpb "github.com/slntopp/nocloud-proto/billing"
	apb "github.com/slntopp/nocloud-proto/billing/addons"
	pb "github.com/slntopp/nocloud-proto/billing/promocodes"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"slices"
	"strings"
	"time"
)

type PromocodesServer struct {
	log *zap.Logger

	db driver.Database

	promos graph.PromocodesController
	nss    graph.NamespacesController
	plans  graph.BillingPlansController
	addons graph.AddonsController
	curr   graph.CurrencyController
	ca     graph.CommonActionsController
}

func NewPromocodesServer(logger *zap.Logger, db driver.Database,
	promocodes graph.PromocodesController, nss graph.NamespacesController,
	plans graph.BillingPlansController, addons graph.AddonsController, curr graph.CurrencyController,
	ca graph.CommonActionsController) *PromocodesServer {
	log := logger.Named("PromocodesServer")
	return &PromocodesServer{
		log:    log,
		db:     db,
		promos: promocodes,
		nss:    nss,
		plans:  plans,
		addons: addons,
		curr:   curr,
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
	if parts[0] == "invoices" {
		res.Invoice = &parts[1]
	} else if parts[0] == "instances" {
		res.Instance = &parts[1]
	} else {
		return nil, fmt.Errorf("invalid resource type: %s", parts[0])
	}
	return res, nil
}

func (s *PromocodesServer) ApplySale(ctx context.Context, r *connect.Request[bpb.ApplySaleRequest]) (*connect.Response[bpb.ApplySaleResponse], error) {
	log := s.log.Named("ApplySale")
	req := r.Msg
	_, _ = ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("body", req))

	var (
		addons = make([]*apb.Addon, 0)
		plans  = make([]*bpb.Plan, 0)
		promos = make([]*pb.Promocode, 0)
	)
	stringsToAny := func(s []string) []any {
		result := make([]any, len(s))
		for i, v := range s {
			result[i] = v
		}
		return result
	}

	if len(req.GetPromocodes()) > 0 {
		l, err := structpb.NewValue(stringsToAny(req.GetPromocodes()))
		if err != nil {
			log.Error("Failed to construct structpb.Value", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to convert")
		}
		promos, err = s.promos.List(ctx, &pb.ListPromocodesRequest{Filters: map[string]*structpb.Value{
			"uuids": l,
		}})
	}
	if len(req.GetBillingPlan()) > 0 {
		_plans, err := s.plans.List(ctx, "", req.GetBillingPlan())
		if err != nil {
			log.Error("Failed to list billing plans", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to obtain plans")
		}
		for _, p := range _plans {
			if p != nil {
				plans = append(plans, p.Plan)
			}
		}
	}
	if len(req.GetAddons()) > 0 {
		l, err := structpb.NewValue(stringsToAny(req.GetAddons()))
		if err != nil {
			log.Error("Failed to construct structpb.Value", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to convert")
		}
		addons, err = s.addons.List(ctx, &apb.ListAddonsRequest{Filters: map[string]*structpb.Value{
			"uuids": l,
		}})
		if err != nil {
			log.Error("Failed to list addons", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to obtain addons")
		}
	}

	planAddons := map[string]string{}
	if len(plans) > 0 && plans[0] != nil {
		plan := plans[0]
		for _, a := range plan.GetAddons() {
			planAddons[a] = plan.GetUuid()
		}
		for _, prod := range plan.GetProducts() {
			for _, a := range prod.GetAddons() {
				planAddons[a] = plan.GetUuid()
			}
		}
	}

	// Promocode apply
	var disc float64

	for _, a := range addons {
		for k, v := range a.GetPeriods() {
			disc, _ = s.promos.CalculateResourceDiscount(promos, planAddons[a.GetUuid()], "addon", a.GetUuid(), v)
			a.GetPeriods()[k] = v - disc
		}
	}

	for _, p := range plans {
		for _, res := range p.GetResources() {
			if res == nil {
				continue
			}
			disc, _ = s.promos.CalculateResourceDiscount(promos, p.GetUuid(), "resource", res.GetKey(), res.GetPrice())
			res.Price -= disc
		}
		for k, prod := range p.GetProducts() {
			if prod == nil {
				continue
			}
			disc, _ = s.promos.CalculateResourceDiscount(promos, p.GetUuid(), "product", k, prod.GetPrice())
			prod.Price -= disc
		}
	}

	//

	conv := graph.NewConverter(r.Header(), s.curr)
	for _, val := range addons {
		conv.ConvertObjectPrices(val)
	}
	for _, val := range plans {
		conv.ConvertObjectPrices(val)
	}
	resp := connect.NewResponse(&bpb.ApplySaleResponse{
		BillingPlans: plans,
		Addons:       addons,
	})
	conv.SetResponseHeader(resp.Header())
	return graph.HandleConvertionError(resp, conv)
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
	requester, ok := ctx.Value(nocloud.NoCloudAccount).(string)
	var isAdmin bool
	if ok {
		isAdmin = s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN)
	}

	promo, err := s.promos.GetByCode(ctx, r.Msg.GetCode(), requester)
	if err != nil {
		log.Error("Failed to get promocode by code", zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("promocode not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get promocode"))
	}

	if !isAdmin {
		if promo.Status != pb.PromocodeStatus_ACTIVE && promo.Status != pb.PromocodeStatus_STATUS_UNKNOWN {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("promocode not found"))
		}
		if promo.Condition == pb.PromocodeCondition_CONDITION_UNKNOWN {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("can't apply promocode"))
		}
		if promo.Condition == pb.PromocodeCondition_EXPIRED {
			return nil, connect.NewError(connect.CodeResourceExhausted, fmt.Errorf("promocode is expired"))
		}
		if promo.Condition == pb.PromocodeCondition_USED {
			return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("maximum activations per account"))
		}
		if promo.Condition == pb.PromocodeCondition_ALL_TAKEN {
			return nil, connect.NewError(connect.CodeOutOfRange, fmt.Errorf("global limit exceeded"))
		}
	}

	if r.Msg.BillingPlan != nil {
		if ok, err = s.promos.IsPlanAffected(ctx, promo, r.Msg.GetBillingPlan()); !ok || err != nil {
			log.Error("Requested promocode doesn't affect passed billing plan", zap.Error(err), zap.Bool("result", ok))
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("promocode has no effect"))
		}
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
	log.Debug("List promocodes request received", zap.Any("request", r))
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
