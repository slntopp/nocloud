/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package billing

import (
	"connectrpc.com/connect"
	"context"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	dpb "github.com/slntopp/nocloud-proto/billing/descriptions"
	healthpb "github.com/slntopp/nocloud-proto/health"
	statuspb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"strconv"
)

type Routine struct {
	Name     string
	LastExec string
	Running  bool
}

type BillingServiceServer struct {
	log *zap.Logger

	nss          graph.NamespacesController
	plans        graph.BillingPlansController
	transactions graph.TransactionsController
	invoices     graph.InvoicesController
	records      graph.RecordsController
	currencies   graph.CurrencyController
	accounts     graph.AccountsController
	descriptions *graph.DescriptionsController

	db driver.Database

	gen  *healthpb.RoutineStatus
	proc *healthpb.RoutineStatus
	sus  *healthpb.RoutineStatus
}

func NewBillingServiceServer(logger *zap.Logger, db driver.Database) *BillingServiceServer {
	log := logger.Named("BillingService")
	s := &BillingServiceServer{
		log:          log,
		nss:          graph.NewNamespacesController(log, db),
		plans:        graph.NewBillingPlansController(log.Named("PlansController"), db),
		transactions: graph.NewTransactionsController(log.Named("TransactionsController"), db),
		records:      graph.NewRecordsController(log.Named("RecordsController"), db),
		currencies:   graph.NewCurrencyController(log.Named("CurrenciesController"), db),
		accounts:     graph.NewAccountsController(log.Named("AccountsController"), db),
		invoices:     graph.NewInvoicesController(log.Named("InvoicesController"), db),
		descriptions: graph.NewDescriptionsController(log, db),
		db:           db,
		gen: &healthpb.RoutineStatus{
			Routine: "Generate Transactions",
			Status: &healthpb.ServingStatus{
				Service: "Billing Machine",
				Status:  healthpb.Status_STOPPED,
			},
		}, proc: &healthpb.RoutineStatus{
			Routine: "Process Transactions",
			Status: &healthpb.ServingStatus{
				Service: "Billing Machine",
				Status:  healthpb.Status_STOPPED,
			},
		},
		sus: &healthpb.RoutineStatus{
			Routine: "Suspend Monitoring",
			Status: &healthpb.ServingStatus{
				Service: "Billing Machine",
				Status:  healthpb.Status_STOPPED,
			},
		},
	}

	s.migrate()

	return s
}

func (s *BillingServiceServer) migrate() {
	ctx := context.Background()
	log := s.log.Named("migrate")
	plans, err := s.plans.List(ctx, "")
	if err != nil {
		log.Error("failed to list plans", zap.Error(err))
		return
	}

	for _, plan := range plans {
		shouldUpdate := false

		for _, res := range plan.GetResources() {
			if res.GetMeta() == nil {
				continue
			}

			desc, ok := res.GetMeta()["description"]

			if res.GetDescriptionId() == "" && ok {
				create, err := s.descriptions.Create(ctx, &dpb.Description{
					Text: desc.GetStringValue(),
				})
				if err != nil {
					log.Error("failed to create description", zap.Error(err))
					return
				}
				res.DescriptionId = create.GetUuid()
				shouldUpdate = true
			}
		}

		for _, prod := range plan.GetProducts() {
			if prod.GetMeta() == nil {
				continue
			}

			desc, ok := prod.GetMeta()["description"]

			if prod.GetDescriptionId() == "" && ok {
				create, err := s.descriptions.Create(ctx, &dpb.Description{
					Text: desc.GetStringValue(),
				})
				if err != nil {
					log.Error("failed to create description", zap.Error(err))
					return
				}
				prod.DescriptionId = create.GetUuid()
				shouldUpdate = true
			}
		}

		if shouldUpdate {
			_, err := s.plans.Update(ctx, plan.Plan)
			if err != nil {
				log.Error("Failed to update plan")
				return
			}
		}
	}
	log.Info("Finished migration")
}

func (s *BillingServiceServer) CreatePlan(ctx context.Context, req *connect.Request[pb.Plan]) (*connect.Response[pb.Plan], error) {
	log := s.log.Named("CreatePlan")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	plan := req.Msg
	log.Debug("request", zap.Any("plan", plan), zap.String("requestor", requestor))

	ns, err := s.nss.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage BillingPlans")
	}

	res, err := s.plans.Create(ctx, plan)
	if err != nil {
		log.Error("Error creating plan", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error creating plan")
	}

	resp := connect.NewResponse(res.Plan)

	return resp, nil
}

func (s *BillingServiceServer) UpdatePlan(ctx context.Context, req *connect.Request[pb.Plan]) (*connect.Response[pb.Plan], error) {
	log := s.log.Named("UpdatePlan")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	plan := req.Msg
	log.Debug("request", zap.Any("plan", plan), zap.String("requestor", requestor))

	ns, err := s.nss.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage BillingPlans")
	}

	pbStatus, err := s.plans.CheckStatus(ctx, plan)
	if err != nil {
		return nil, err
	}

	if pbStatus == statuspb.NoCloudStatus_DEL {
		return nil, status.Error(codes.Canceled, "Billing plan deleted")
	}

	res, err := s.plans.Update(ctx, plan)
	if err != nil {
		log.Error("Error updating plan", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error updating plan")
	}

	resp := connect.NewResponse(res.Plan)

	return resp, nil
}

func (s *BillingServiceServer) DeletePlan(ctx context.Context, req *connect.Request[pb.Plan]) (*connect.Response[pb.Plan], error) {
	log := s.log.Named("DeletePlan")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	plan := req.Msg
	log.Debug("request", zap.Any("plan", plan), zap.String("requestor", requestor))

	ns, err := s.nss.Get(ctx, schema.ROOT_NAMESPACE_KEY)
	if err != nil {
		return nil, err
	}
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage BillingPlans")
	}

	/*planId := driver.NewDocumentID(schema.BILLING_PLANS_COL, plan.GetUuid())

	cursor, err := s.db.Query(ctx, getPlanInstances, map[string]interface{}{
		"permissions":       schema.PERMISSIONS_GRAPH.Name,
		"plan":              planId,
		"@instances_groups": schema.INSTANCES_GROUPS_COL,
		"@instances":        schema.INSTANCES_COL,
		"status":            statuspb.NoCloudStatus_DEL,
	})
	if err != nil {
		log.Error("Error getting instances", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting instances")
	}

	if cursor.HasMore() {
		return nil, status.Error(codes.DataLoss, "Сan't delete plan due to related instances")
	}*/

	err = s.plans.Delete(ctx, plan)
	if err != nil {
		log.Error("Error deleting plan", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error deleting plan")
	}

	resp := connect.NewResponse(plan)

	return resp, nil
}

func (s *BillingServiceServer) GetPlan(ctx context.Context, req *connect.Request[pb.Plan]) (*connect.Response[pb.Plan], error) {
	log := s.log.Named("GetPlan")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	plan := req.Msg
	log.Debug("request", zap.Any("plan", plan), zap.String("requestor", requestor))

	p, err := s.plans.Get(ctx, plan)
	if err != nil {
		log.Error("Error getting plan", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting plan")
	}

	resp := connect.NewResponse(p.Plan)

	if p.Public && p.GetStatus() != statuspb.NoCloudStatus_DEL {
		return resp, nil
	}

	namespaceId := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, namespaceId, access.Level_ROOT)

	if ok {
		return resp, nil
	}

	ok = graph.HasAccess(ctx, s.db, requestor, p.ID, access.Level_READ)

	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage BillingPlans")
	}

	if p.GetStatus() == statuspb.NoCloudStatus_DEL {
		return nil, status.Error(codes.NotFound, "Plan was deleted")
	}

	return resp, nil
}

var getDefaultCurrencyQuery = `
LET cur = LAST(
    FOR i IN Currencies2Currencies
    FILTER (i.to == 0 || i.from == 0) && i.rate == 1
        RETURN i
)

RETURN cur.to == 0 ? cur.from : cur.to
`

func (s *BillingServiceServer) ListPlans(ctx context.Context, r *connect.Request[pb.ListRequest]) (*connect.Response[pb.ListResponse], error) {
	log := s.log.Named("ListPlans")

	req := r.Msg

	var requestor string
	if !req.Anonymously {
		requestor = ctx.Value(nocloud.NoCloudAccount).(string)
	}
	log.Debug("Requestor", zap.String("id", requestor))

	plans, err := s.plans.List(ctx, req.SpUuid)
	if err != nil {
		log.Error("Error listing plans", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error listing plans")
	}

	namespaceId := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, namespaceId, access.Level_ROOT)

	result := make([]*pb.Plan, 0)
	for _, plan := range plans {
		if plan.Public && (plan.GetStatus() != statuspb.NoCloudStatus_DEL || req.GetShowDeleted()) {
			result = append(result, plan.Plan)
			continue
		}
		if req.Anonymously {
			continue
		}
		if !ok {
			continue
		}
		if !(plan.GetStatus() == statuspb.NoCloudStatus_DEL && req.GetShowDeleted()) {
			continue
		}

		result = append(result, plan.Plan)
	}

	if !req.Anonymously {
		acc, err := s.accounts.Get(ctx, requestor)
		if err != nil {
			log.Error("Error getting account", zap.Error(err))
			return nil, status.Error(codes.Internal, "Error getting account")
		}

		cur := acc.Account.GetCurrency()

		dbCur := graph.Currency{}
		queryContext := driver.WithQueryCount(ctx)
		res, err := s.db.Query(queryContext, getDefaultCurrencyQuery, map[string]interface{}{})
		if err != nil {
			return nil, err
		}
		if res.Count() != 0 {
			_, err = res.ReadDocument(ctx, &dbCur)
			if err != nil {
				log.Error("Failed to get default cur", zap.Error(err))
				return nil, status.Error(codes.Internal, "Failed to get default cur")
			}
		}

		id, err := strconv.ParseInt(dbCur.Key, 10, 32)
		if err != nil {
			log.Error("Failed to parse int", zap.Error(err))
		}
		defaultCur := &pb.Currency{
			Id:   id,
			Name: dbCur.Name,
		}

		var rate float64

		if cur.GetId() == defaultCur.GetId() {
			rate = 1
		} else {
			rate, err = s.currencies.GetExchangeRateDirect(ctx, defaultCur, cur)
			if err != nil {
				log.Error("Error getting rate", zap.Error(err))
				return nil, status.Error(codes.Internal, "Error getting rate")
			}
		}

		for planIndex := range result {
			plan := result[planIndex]

			products := plan.GetProducts()
			for key := range products {
				products[key].Price *= rate
			}
			plan.Products = products

			resources := plan.GetResources()
			for index := range resources {
				resources[index].Price *= rate
			}
			plan.Resources = resources

			result[planIndex] = plan
		}
	}

	resp := connect.NewResponse(&pb.ListResponse{Pool: result})

	return resp, nil
}

func (s *BillingServiceServer) ListPlansInstances(ctx context.Context, r *connect.Request[pb.ListPlansInstancesRequest]) (*connect.Response[pb.ListPlansInstancesResponse], error) {
	log := s.log.Named("ListPlans")

	req := r.Msg

	var requestor string
	if !req.Anonymously {
		requestor = ctx.Value(nocloud.NoCloudAccount).(string)
	}
	log.Debug("Requestor", zap.String("id", requestor))

	plans, err := s.plans.List(ctx, "")
	if err != nil {
		log.Error("Error listing plans", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error listing plans")
	}

	cursor, err := s.db.Query(ctx, getInstancesBillingPlans, map[string]interface{}{
		"status": statuspb.NoCloudStatus_DEL,
	})
	if err != nil {
		return nil, err
	}

	var planInstancesCount = make(map[string]int)

	for cursor.HasMore() {
		planUuid := ""
		_, err := cursor.ReadDocument(ctx, &planUuid)
		if err != nil {
			return nil, err
		}
		if _, ok := planInstancesCount[planUuid]; !ok {
			planInstancesCount[planUuid] = 1
		} else {
			planInstancesCount[planUuid] += 1
		}
	}

	namespaceId := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, namespaceId, access.Level_ROOT)

	result := make(map[string]*structpb.Value)

	for _, plan := range plans {
		if plan.Public {
			result[plan.GetUuid()] = structpb.NewNumberValue(float64(planInstancesCount[plan.GetUuid()]))
			continue
		}
		if req.Anonymously {
			continue
		}
		if !ok {
			continue
		}

		result[plan.GetUuid()] = structpb.NewNumberValue(float64(planInstancesCount[plan.GetUuid()]))
	}

	resp := connect.NewResponse(&pb.ListPlansInstancesResponse{Plans: result})

	return resp, nil
}

const getInstancesBillingPlans = `
FOR inst in Instances
	FILTER inst.status != @status
	RETURN inst.billing_plan.uuid
`

/*
const getPlanInstances = `
LET igs = (
    FOR node IN 2 INBOUND @plan GRAPH @permissions
    FILTER IS_SAME_COLLECTION(node, @@instances_groups)
    RETURN node._id
)

LET instances = (
	FOR ig in igs
    	FOR node, edge IN 1 OUTBOUND ig GRAPH @permissions
    	FILTER IS_SAME_COLLECTION(node, @@instances)
    	FILTER edge.role == "owner"
    	RETURN node
)

LET plan = DOCUMENT(@plan)

FOR inst in instances
	FILTER inst.billing_plan.uuid == plan._key
	FILTER inst.status != @status
	RETURN inst
`
*/
