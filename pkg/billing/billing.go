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
	"context"
	"fmt"
	epb "github.com/slntopp/nocloud-proto/events"
	ccinstances "github.com/slntopp/nocloud-proto/instances/instancesconnect"
	registrypb "github.com/slntopp/nocloud-proto/registry"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/whmcs_gateway"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"slices"
	"strings"

	"encoding/json"
	"time"

	"connectrpc.com/connect"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	dpb "github.com/slntopp/nocloud-proto/billing/descriptions"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	elpb "github.com/slntopp/nocloud-proto/events_logging"
	healthpb "github.com/slntopp/nocloud-proto/health"
	statuspb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/wI2L/jsondiff"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

const internalAccessKey = nocloud.ContextKey("billing-internal-access")

func ctxWithRoot(ctx context.Context) context.Context {
	return context.WithValue(ctx, nocloud.NoCloudAccount, schema.ROOT_NAMESPACE_KEY)
}

func ctxWithInternalAccess(ctx context.Context) context.Context {
	return context.WithValue(ctx, internalAccessKey, true)
}

func hasInternalAccess(ctx context.Context) bool {
	v, _ := ctx.Value(internalAccessKey).(bool)
	return v
}

type BillingServiceServer struct {
	log *zap.Logger

	rbmq                    rabbitmq.Connection
	ConsumerStatus          *healthpb.RoutineStatus
	InstancesConsumerStatus *healthpb.RoutineStatus

	nss          graph.NamespacesController
	plans        graph.BillingPlansController
	transactions graph.TransactionsController
	invoices     graph.InvoicesController
	records      graph.RecordsController
	currencies   graph.CurrencyController
	accounts     graph.AccountsController
	descriptions graph.DescriptionsController
	instances    graph.InstancesController
	services     graph.ServicesController
	sp           graph.ServicesProvidersController
	addons       graph.AddonsController
	promocodes   graph.PromocodesController
	ca           graph.CommonActionsController

	db  driver.Database
	rdb redisdb.Client

	settingsClient  settingspb.SettingsServiceClient
	accClient       registrypb.AccountsServiceClient
	eventsClient    epb.EventsServiceClient
	instancesClient ccinstances.InstancesServiceClient

	cron *healthpb.RoutineStatus

	gen  *healthpb.RoutineStatus
	proc *healthpb.RoutineStatus
	sus  *healthpb.RoutineStatus

	drivers map[string]driverpb.DriverServiceClient

	rootToken string

	whmcsGateway *whmcs_gateway.WhmcsGateway

	invoicesPublisher func(event *epb.Event) error
}

func NewBillingServiceServer(logger *zap.Logger, db driver.Database, conn rabbitmq.Connection, rdb redisdb.Client, drivers map[string]driverpb.DriverServiceClient, token string,
	settingsClient settingspb.SettingsServiceClient, accClient registrypb.AccountsServiceClient, eventsClient epb.EventsServiceClient, instClient ccinstances.InstancesServiceClient,
	nss graph.NamespacesController, plans graph.BillingPlansController, transactions graph.TransactionsController, invoices graph.InvoicesController,
	records graph.RecordsController, currencies graph.CurrencyController, accounts graph.AccountsController, descriptions graph.DescriptionsController,
	instances graph.InstancesController, sp graph.ServicesProvidersController, services graph.ServicesController, addons graph.AddonsController,
	ca graph.CommonActionsController, promocodes graph.PromocodesController, whmcsGateway *whmcs_gateway.WhmcsGateway, invPub func(event *epb.Event) error) *BillingServiceServer {
	log := logger.Named("BillingService")
	s := &BillingServiceServer{
		rbmq:              conn,
		log:               log,
		nss:               nss,
		plans:             plans,
		transactions:      transactions,
		records:           records,
		currencies:        currencies,
		accounts:          accounts,
		invoices:          invoices,
		services:          services,
		descriptions:      descriptions,
		instances:         instances,
		sp:                sp,
		addons:            addons,
		promocodes:        promocodes,
		ca:                ca,
		db:                db,
		rdb:               rdb,
		drivers:           drivers,
		settingsClient:    settingsClient,
		accClient:         accClient,
		eventsClient:      eventsClient,
		instancesClient:   instClient,
		rootToken:         token,
		whmcsGateway:      whmcsGateway,
		invoicesPublisher: invPub,
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
		cron: &healthpb.RoutineStatus{
			Routine: "Daily Cron Job (UTC)",
			Status: &healthpb.ServingStatus{
				Service: "Billing Machine",
				Status:  healthpb.Status_STOPPED,
			},
		},
		ConsumerStatus: &healthpb.RoutineStatus{
			Routine: "Billing Records Consumer",
			Status: &healthpb.ServingStatus{
				Service: "Billing Machine",
				Status:  healthpb.Status_STOPPED,
			},
		},
		InstancesConsumerStatus: &healthpb.RoutineStatus{
			Routine: "Invoices Issuer Consumer",
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
	ok := s.ca.HasAccess(ctx, requestor, ns.ID, access.Level_ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage BillingPlans")
	}

	// Check title unique
	plans, err := s.plans.List(ctx, "")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error creating plan: %v", err))
	}
	for _, p := range plans {
		if p.Status == statuspb.NoCloudStatus_DEL {
			continue
		}
		if plan.Title == p.Title {
			return nil, status.Error(codes.AlreadyExists, "plan with the same title already exists")
		}
	}

	res, err := s.plans.Create(ctx, plan)
	var event = &elpb.Event{
		Entity:    schema.BILLING_PLANS_COL,
		Scope:     "database",
		Action:    "create",
		Rc:        0,
		Requestor: ctx.Value(nocloud.NoCloudAccount).(string),
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}
	if err != nil {
		event.Rc = 1
		nocloud.Log(log, event)
		log.Error("Error creating plan", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("error creating plan: %v", err))
	}

	resp := connect.NewResponse(res.Plan)
	nocloud.Log(log, event)

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
	ok := s.ca.HasAccess(ctx, requestor, ns.ID, access.Level_ADMIN)
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

	// Check title unique
	plans, err := s.plans.List(ctx, "")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error creating plan: %v", err))
	}
	for _, p := range plans {
		if p.GetUuid() == plan.GetUuid() {
			continue
		}
		if p.Status == statuspb.NoCloudStatus_DEL {
			continue
		}
		if plan.Title == p.Title {
			return nil, status.Error(codes.AlreadyExists, "plan with the same title already exists")
		}
	}

	res, err := s.plans.Update(ctx, plan)
	oldPlanMarshal, _ := json.Marshal(plan)
	newPlanMarshal, _ := json.Marshal(res)
	var event = &elpb.Event{
		Entity:    schema.BILLING_PLANS_COL,
		Uuid:      plan.GetUuid(),
		Scope:     "database",
		Action:    "update",
		Rc:        0,
		Requestor: ctx.Value(nocloud.NoCloudAccount).(string),
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}
	if err != nil {
		event.Rc = 0
		nocloud.Log(log, event)
		log.Error("Error updating plan", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("error updating plan: %v", err))
	}

	diff, err := jsondiff.CompareJSON(oldPlanMarshal, newPlanMarshal)
	if err != nil {
		event.Rc = 0
		nocloud.Log(log, event)
	} else {
		event.Snapshot.Diff = diff.String()
		nocloud.Log(log, event)
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
	ok := s.ca.HasAccess(ctx, requestor, ns.ID, access.Level_ADMIN)
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

	var event = &elpb.Event{
		Entity:    schema.BILLING_PLANS_COL,
		Uuid:      plan.Uuid,
		Scope:     "database",
		Action:    "delete",
		Rc:        0,
		Requestor: ctx.Value(nocloud.NoCloudAccount).(string),
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	err = s.plans.Delete(ctx, plan)
	if err != nil {
		event.Rc = 1
		nocloud.Log(log, event)
		log.Error("Error deleting plan", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error deleting plan")
	}

	resp := connect.NewResponse(plan)
	nocloud.Log(log, event)

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
	ok := s.ca.HasAccess(ctx, requestor, namespaceId, access.Level_ROOT)

	if ok {
		return resp, nil
	}

	ok = s.ca.HasAccess(ctx, requestor, p.ID, access.Level_READ)

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
    FILTER (i.to.id == 0 || i.from.id == 0) && i.rate == 1
        RETURN i
)

RETURN cur.to.id == 0 ? cur.from : cur.to
`

func buildPlansListQuery(req *pb.ListRequest, hasAccess bool) (string, map[string]interface{}) {
	var query string
	var vars = map[string]interface{}{}
	if req.GetSpUuid() == "" {
		query = `FOR p IN @@plans`
		vars["@plans"] = schema.BILLING_PLANS_COL
	} else {
		query = `FOR p, edge IN 1 OUTBOUND @sp GRAPH Billing`
		query += ` FILTER IS_SAME_COLLECTION(p, @@plans)`
		spDocId := driver.NewDocumentID(schema.SERVICES_PROVIDERS_COL, req.GetSpUuid())
		vars["sp"] = spDocId
		vars["@plans"] = schema.BILLING_PLANS_COL
	}

	if len(req.GetExcludeUuids()) > 0 {
		query += ` FILTER p._key NOT IN @excludeUuids`
		vars["excludeUuids"] = req.GetExcludeUuids()
	}

	if req.GetAnonymously() {
		query += ` FILTER p.public == true`
		query += ` FILTER COUNT(
    FOR sp, edge2 IN 1 INBOUND p GRAPH Billing
        FILTER IS_SAME_COLLECTION(@@services_providers, sp)
        RETURN 1
) > 0`
		vars["@services_providers"] = schema.SERVICES_PROVIDERS_COL
	}

	if !req.GetShowDeleted() {
		query += ` FILTER p.status != @status`
		vars["status"] = statuspb.NoCloudStatus_DEL
	}

	if !hasAccess {
		query += ` FILTER p.public == true`
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "kind" || key == "type" {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER p.%s in @%s`, key, key)
				vars[key] = values
			} else if key == "meta.isIndividual" {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER TO_BOOL(p.meta.isIndividual) in @individual`)
				vars["individual"] = values
			} else if key == "search_param" {
				query += fmt.Sprintf(` FILTER LOWER(p["title"]) LIKE LOWER("%s")
|| p._key LIKE "%s"`,
					"%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%")
			} else if key == "public" {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER TO_BOOL(p.%s) in @%s`, key, key)
				vars[key] = values
			} else if key == "status" {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				varsKey := "filteredStatuses"
				query += fmt.Sprintf(` FILTER TO_NUMBER(p.%s) in @%s`, key, varsKey)
				vars[varsKey] = values
			} else {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER t["%s"] in @%s`, key, key)
				vars[key] = values
			}
		}
	}

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT p.%s %s`
		field, sort := req.GetField(), req.GetSort()

		if field == "total" {
			if strings.ToLower(sort) == "asc" {
				sort = "desc"
			} else {
				sort = "asc"
			}
		}

		query += fmt.Sprintf(subQuery, field, sort)
	}

	if req.Page != nil && req.Limit != nil {
		if req.GetLimit() != 0 {
			limit, page := req.GetLimit(), req.GetPage()
			offset := (page - 1) * limit

			query += ` LIMIT @offset, @count`
			vars["offset"] = offset
			vars["count"] = limit
		}
	}

	query += ` RETURN merge(p, {uuid: p._key})`
	return query, vars
}

func (s *BillingServiceServer) ListPlans(ctx context.Context, r *connect.Request[pb.ListRequest]) (*connect.Response[pb.ListResponse], error) {
	log := s.log.Named("ListPlans")
	req := r.Msg

	var requestor string
	if !req.Anonymously {
		requestor = ctx.Value(nocloud.NoCloudAccount).(string)
	}
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))
	ok := s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT)

	if req.GetUuid() != "" {
		return s._HandleGetSinglePlan(ctx, r)
	}

	query, vars := buildPlansListQuery(req, ok)
	log.Debug("Ready to retrieve plans", zap.String("query", query), zap.Any("vars", vars))
	req.Limit = nil
	req.Page = nil
	countQuery, countVars := buildPlansListQuery(req, ok)
	countQuery = fmt.Sprintf(`RETURN COUNT(%s)`, countQuery)

	// Getting plans
	cursor, err := s.db.Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to retrieve plans", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve plans")
	}
	defer cursor.Close()

	var plans []*pb.Plan
	for {
		plan := &pb.Plan{}
		meta, err := cursor.ReadDocument(ctx, plan)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			log.Error("Failed to retrieve plans", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to retrieve plans")
		}
		plan.Uuid = meta.Key
		plans = append(plans, plan)
	}

	// Getting plans count
	cursor, err = s.db.Query(ctx, countQuery, countVars)
	if err != nil {
		log.Error("Failed to retrieve plans count", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve plans count")
	}
	defer cursor.Close()
	var count int
	if _, err = cursor.ReadDocument(ctx, &count); err != nil {
		log.Error("Failed to retrieve plans count", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve plans count")
	}
	log.Debug("Plans count", zap.Int("count", count))

	// Apply account prices
	if !req.Anonymously {
		acc, err := s.accounts.Get(ctx, requestor)
		if err != nil {
			log.Error("Error getting account", zap.Error(err))
			return nil, status.Error(codes.Internal, "Error getting account")
		}

		cur := acc.Account.GetCurrency()
		if cur == nil {
			cur = &pb.Currency{
				Id:    schema.DEFAULT_CURRENCY_ID,
				Title: schema.DEFAULT_CURRENCY_NAME,
			}
		}

		dbCur := struct {
			Id    int32  `json:"id"`
			Title string `json:"title"`
		}{}
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

		defaultCur := &pb.Currency{
			Id:    dbCur.Id,
			Title: dbCur.Title,
		}

		var (
			rate float64
		)

		if cur.GetId() == defaultCur.GetId() {
			rate = 1
		} else {
			rate, _, err = s.currencies.GetExchangeRate(ctx, defaultCur, cur)
			if err != nil {
				log.Error("Error getting rate", zap.Error(err))
				return nil, status.Error(codes.Internal, "Error getting rate")
			}
		}

		for planIndex := range plans {
			plan := plans[planIndex]
			graph.ConvertPlan(plan, rate, cur.Precision, cur.Rounding)
		}

		log.Debug("Plans retrieved", zap.Any("plans", plans), zap.Int("count", len(plans)), zap.Int("total", count))
		resp := connect.NewResponse(&pb.ListResponse{Pool: plans, Total: uint64(count)})
		graph.ExplicitSetPrimaryCurrencyHeader(resp.Header(), cur.Code)
		return resp, nil
	}

	conv := graph.NewConverter(r.Header(), s.currencies)
	for _, plan := range plans {
		conv.ConvertObjectPrices(plan)
	}

	log.Debug("Plans retrieved", zap.Any("plans", plans), zap.Int("count", len(plans)), zap.Int("total", count))
	resp := connect.NewResponse(&pb.ListResponse{Pool: plans, Total: uint64(count)})
	conv.SetResponseHeader(resp.Header())
	return graph.HandleConvertionError(resp, conv)
}

const countPlansQuery = `
LET plans = (
    %s
)

let type = (
 FOR p IN plans
 FILTER p.type
 FILTER p.type != ""
 RETURN DISTINCT p.type
)

return { 
	unique: {
        type: type,
	},
	total: LENGTH(plans)
}
`

func (s *BillingServiceServer) PlansUnique(ctx context.Context, _req *connect.Request[pb.PlansUniqueRequest]) (*connect.Response[pb.PlansUniqueResponse], error) {
	log := s.log.Named("PlansUnique")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))
	ok := s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT)

	type Response struct {
		Total  int                    `json:"total"`
		Unique map[string]interface{} `json:"unique"`
	}

	query, vars := buildPlansListQuery(&pb.ListRequest{
		SpUuid:      req.GetSpUuid(),
		Anonymously: req.GetAnonymously(),
		ShowDeleted: req.GetShowDeleted(),
		Filters:     req.GetFilters(),
	}, ok)

	query = fmt.Sprintf(countPlansQuery, query)

	s.log.Debug("Query", zap.Any("q", query))
	s.log.Debug("Ready to build query", zap.Any("bindVars", vars))

	c, err := s.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var resp Response
	_, err = c.ReadDocument(ctx, &resp)
	if err != nil {
		return nil, err
	}
	log.Debug("Response", zap.Any("resp", resp))

	var result pb.PlansUniqueResponse
	obj, err := structpb.NewStruct(resp.Unique)
	if err != nil {
		return nil, err
	}
	result.Unique = structpb.NewStructValue(obj)
	result.Total = uint64(resp.Total)

	return connect.NewResponse(&result), nil
}

// ListPlansInstances TODO: Optimize to fetch only provided plans in uuids array, not all plans by query
func (s *BillingServiceServer) ListPlansInstances(ctx context.Context, r *connect.Request[pb.ListPlansInstancesRequest]) (*connect.Response[pb.ListPlansInstancesResponse], error) {
	log := s.log.Named("ListPlans")

	req := r.Msg

	var requestor string
	if !req.Anonymously {
		requestor = ctx.Value(nocloud.NoCloudAccount).(string)
	}
	log.Debug("Requestor", zap.String("id", requestor))

	var useUuidsArray = false
	if len(req.GetUuids()) > 0 {
		useUuidsArray = true
	}

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
	ok := s.ca.HasAccess(ctx, requestor, namespaceId, access.Level_ROOT)

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
		if useUuidsArray && !slices.Contains(req.GetUuids(), plan.GetUuid()) {
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

func (s *BillingServiceServer) _HandleGetSinglePlan(ctx context.Context, r *connect.Request[pb.ListRequest]) (*connect.Response[pb.ListResponse], error) {
	var requester string
	if !r.Msg.GetAnonymously() {
		requester = ctx.Value(nocloud.NoCloudAccount).(string)
	}

	p, err := s.plans.Get(ctx, &pb.Plan{Uuid: r.Msg.GetUuid()})
	if err != nil {
		return nil, status.Error(codes.NotFound, "Plan doesn't exist")
	}

	ok := s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT)

	conv := graph.NewConverter(r.Header(), s.currencies)
	conv.ConvertObjectPrices(p)
	if r.Msg.GetAnonymously() && p.Public {
		resp := connect.NewResponse(&pb.ListResponse{Pool: []*pb.Plan{p.Plan}, Total: 1})
		conv.SetResponseHeader(resp.Header())
		return graph.HandleConvertionError(resp, conv)
	}

	if !ok && !p.Public {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	resp := connect.NewResponse(&pb.ListResponse{Pool: []*pb.Plan{p.Plan}, Total: 1})
	conv.SetResponseHeader(resp.Header())
	return graph.HandleConvertionError(resp, conv)
}

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
