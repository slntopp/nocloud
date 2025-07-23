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
package instances

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	billingpb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud-proto/health"
	rpb "github.com/slntopp/nocloud-proto/registry/accounts"
	servicespb "github.com/slntopp/nocloud-proto/services"
	"github.com/slntopp/nocloud/pkg/nocloud/payments"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	"github.com/slntopp/nocloud/pkg/nocloud/sync"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"slices"
	go_sync "sync"
	"time"

	elpb "github.com/slntopp/nocloud-proto/events_logging"
	"github.com/slntopp/nocloud-proto/notes"

	"github.com/arangodb/go-driver"
	accesspb "github.com/slntopp/nocloud-proto/access"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	pb "github.com/slntopp/nocloud-proto/instances"
	stpb "github.com/slntopp/nocloud-proto/states"
	spb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	s "github.com/slntopp/nocloud/pkg/states"
	st "github.com/slntopp/nocloud/pkg/statuses"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type InstancesServer struct {
	pb.UnimplementedInstancesServiceServer
	log *zap.Logger

	ctrl       graph.InstancesController
	ig_ctrl    graph.InstancesGroupsController
	sp_ctrl    graph.ServicesProvidersController
	srv_ctrl   graph.ServicesController
	promo_ctrl graph.PromocodesController
	acc_ctrl   graph.AccountsController
	inv_ctrl   graph.InvoicesController
	curr_ctrl  graph.CurrencyController
	ca         graph.CommonActionsController

	drivers map[string]driverpb.DriverServiceClient

	db driver.Database

	rdb redisdb.Client

	monitoring *health.RoutineStatus

	spSyncers map[string]*go_sync.Mutex
}

func NewInstancesServiceServer(logger *zap.Logger, db driver.Database, rbmq rabbitmq.Connection, rdb redisdb.Client) *InstancesServer {
	log := logger.Named("instances")
	log.Debug("New Instances Server Creating")
	ig_ctrl := graph.NewInstancesGroupsController(logger, db, rbmq)
	sp_ctrl := graph.NewServicesProvidersController(logger, db)
	srv_ctrl := graph.NewServicesController(logger, db, rbmq)
	promo_ctrl := graph.NewPromocodesController(logger, db, rbmq)
	acc_ctrl := graph.NewAccountsController(logger, db)
	inv_ctrl := graph.NewInvoicesController(logger, db)
	curr_ctrl := graph.NewCurrencyController(logger, db)
	ca := graph.NewCommonActionsController(logger, db)

	log.Debug("Setting up StatesPubSub")
	s := s.NewStatesPubSub(log, &db, rbmq)
	ch := s.Channel()
	log.Debug("initializing Exchange with name \"states\" of type \"topic\"")
	s.TopicExchange(ch, "states") // init Exchange with name "states" of type "topic"
	log.Debug("initializing Consumer queue of topic \"states.instances\"")
	s.StatesConsumerInit(ch, "states", "instances", schema.INSTANCES_COL) // init Consumer queue of topic "states.instances"

	log.Debug("Setting up PubSub")
	d := NewPubSub(log, &db, rbmq)
	ch = d.Channel()
	log.Debug("initializing Exchange with name \"datas\" of type \"topic\"")
	d.TopicExchange(ch, "datas") // init Exchange with name "datas" of type "topic"
	log.Debug("initializing Consumer queue of topic \"datas.instances\"")
	d.ConsumerInit(ch, "datas", "instances", schema.INSTANCES_COL) // init Consumer queue of topic "datas.instances"
	log.Debug("initializing Consumer queue of topic \"datas.instances-groups\"")
	d.ConsumerInit(ch, "datas", "instances-groups", schema.INSTANCES_GROUPS_COL) // init Consumer queue of topic "datas.instances-groups"

	log.Debug("Setting up StatusesPubSub")
	st := st.NewStatusesPubSub(log, &db, rbmq)
	ch = st.Channel()
	log.Debug("initializing Exchange with name \"statuses\" of type \"topic\"")
	st.TopicExchange(ch, "statuses")
	log.Debug("initializing Consumer queue of topic \"statuses.instances\"")
	st.StatusesConsumerInit(ch, "statuses", "instances", schema.INSTANCES_COL)

	return &InstancesServer{
		db: db, log: log,
		ctrl:       ig_ctrl.Instances(),
		ig_ctrl:    ig_ctrl,
		sp_ctrl:    sp_ctrl,
		srv_ctrl:   srv_ctrl,
		promo_ctrl: promo_ctrl,
		acc_ctrl:   acc_ctrl,
		inv_ctrl:   inv_ctrl,
		curr_ctrl:  curr_ctrl,
		ca:         ca,
		drivers:    make(map[string]driverpb.DriverServiceClient),
		rdb:        rdb,
		monitoring: &health.RoutineStatus{
			Routine: "Monitoring",
			Status: &health.ServingStatus{
				Service: "Services Registry",
				Status:  health.Status_STOPPED,
			},
		},
		spSyncers: make(map[string]*go_sync.Mutex),
	}
}

const nebulaIGType = "ione"

func (s *InstancesServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

var methodsToSync = []string{
	"manual_renew",
	"free_renew",
	"cancel_renew",
	"vpn",
}

func (s *InstancesServer) Invoke(ctx context.Context, _req *connect.Request[pb.InvokeRequest]) (*connect.Response[pb.InvokeResponse], error) {
	log := s.log.Named("invoke")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	instance_id := driver.NewDocumentID(schema.INSTANCES_COL, req.Uuid)
	r, err := s.ctrl.GetGroup(ctx, instance_id.String())
	if err != nil {
		log.Error("Failed to get Group and ServicesProvider", zap.Error(err))
		return nil, err
	}
	log = log.With(zap.String("sp", r.SP.GetUuid()))

	// Sync with driver's monitoring
	if slices.Contains(methodsToSync, req.Method) {
		syncer := sync.NewDataSyncer(log.With(zap.String("caller", "Invoke")), s.rdb, r.SP.GetUuid(), 50, 2000)
		defer syncer.Open()
		_ = syncer.WaitUntilOpenedAndCloseAfter()
	}

	var instance graph.Instance
	instance, err = s.ctrl.GetWithAccess(ctx, requestorId, instance_id.Key())
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	if instance.GetState().GetState() == stpb.NoCloudState_SUSPENDED && instance.GetAccess().GetLevel() < accesspb.Level_ROOT {
		log.Error("Machine is suspended. Functionality is limited", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.Unavailable, "Machine is suspended. Functionality is limited")
	}

	client, ok := s.drivers[r.SP.Type]
	if !ok {
		log.Error("Failed to get driver", zap.String("type", r.SP.Type))
		return nil, status.Error(codes.NotFound, "Driver not found")
	}
	invoke, err := client.Invoke(ctx, &driverpb.InvokeRequest{
		Instance:         instance.Instance,
		ServicesProvider: r.SP,
		Method:           req.Method,
		Params:           req.Params,
	})

	var event = &elpb.Event{
		Entity:    schema.INSTANCES_COL,
		Uuid:      req.GetUuid(),
		Scope:     "driver",
		Action:    req.Method,
		Rc:        0,
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	if err != nil {
		event.Rc = 1
		nocloud.Log(log, event)
		return connect.NewResponse(invoke), err
	}

	nocloud.Log(log, event)

	log.Debug("Request finished")
	return connect.NewResponse(invoke), nil
}

func (s *InstancesServer) Start(ctx context.Context, _req *connect.Request[pb.StartRequest]) (*connect.Response[pb.StartResponse], error) {
	log := s.log.Named("Start")
	req := _req.Msg
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	requesterId := driver.NewDocumentID(schema.ACCOUNTS_COL, requester)
	log.Debug("Requester", zap.String("id", requester))

	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), accesspb.Level_ADMIN) {
		log.Warn("No root access")
		return nil, status.Error(codes.PermissionDenied, "No access rights")
	}

	var instance graph.Instance
	instance, err := s.ctrl.GetWithAccess(ctx, requesterId, req.GetId())
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	if instance.Instance == nil {
		log.Error("Failed to get instance. No object")
		return nil, status.Error(codes.Internal, "Failed to obtain instance")
	}
	if instance.Uuid == "" {
		log.Error("Failed to get instance. No uuid")
		return nil, status.Error(codes.Internal, "Failed to obtain instance")
	}

	if instance.Config == nil {
		instance.Config = make(map[string]*structpb.Value)
	}
	old := proto.Clone(instance.Instance).(*pb.Instance)

	instance.Config["auto_start"] = structpb.NewBoolValue(true)

	if err = s.ctrl.Update(ctx, "", instance.Instance, old); err != nil {
		log.Error("Failed to update instance", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update instance")
	}

	return connect.NewResponse(&pb.StartResponse{
		Result: true,
	}), nil
}

func (s *InstancesServer) Delete(ctx context.Context, _req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	log := s.log.Named("delete")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	var instance graph.Instance
	instance, err := s.ctrl.GetWithAccess(ctx, requestorId, req.GetUuid())
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	err = s.ctrl.SetState(ctx, instance.Instance, stpb.NoCloudState_DELETED)

	if err != nil {
		log.Error("Failed to set state", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.ctrl.SetStatus(ctx, instance.Instance, spb.NoCloudStatus_DEL)

	if err != nil {
		log.Error("Failed to set status", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	var event = &elpb.Event{
		Entity:    schema.INSTANCES_COL,
		Uuid:      req.GetUuid(),
		Scope:     "database",
		Action:    "delete",
		Rc:        0,
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	nocloud.Log(log, event)

	return connect.NewResponse(&pb.DeleteResponse{
		Result: true,
	}), nil
}

func (s *InstancesServer) Create(ctx context.Context, _req *connect.Request[pb.CreateRequest]) (*connect.Response[pb.CreateResponse], error) {
	log := s.log.Named("Create")
	req := _req.Msg
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	requesterId := driver.NewDocumentID(schema.ACCOUNTS_COL, requester)
	log.Debug("Requester", zap.String("id", requester))

	if req.Promocode != nil && req.GetPromocode() != "" {
		ctx = context.WithValue(ctx, graph.CreationPromocodeKey, req.GetPromocode())
	}

	if req.AutoAssign {
		return s.createWithAutoAssign(ctx, req, requester)
	}

	igId := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, req.GetIg())
	var ig graph.InstancesGroup
	ig, err := s.ig_ctrl.GetWithAccess(ctx, requesterId, igId.Key())
	if err != nil {
		log.Error("Failed to get instance group", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	sp, err := s.ig_ctrl.GetSP(ctx, ig.GetUuid())
	if err != nil {
		log.Error("Failed to get sp", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if ig.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	isImported := ig.GetData()["imported"].GetBoolValue()
	if isImported {
		log.Error("Can't create instance with imported IG")
		return nil, status.Error(codes.InvalidArgument, "can't create instance with imported IG")
	}

	newId, err := s.ctrl.Create(ctx, igId, sp.GetUuid(), req.GetInstance())
	if err != nil {
		log.Error("Failed to create instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return connect.NewResponse(&pb.CreateResponse{
		Id:     newId,
		Result: true,
	}), nil
}

func (s *InstancesServer) createWithAutoAssign(ctx context.Context, req *pb.CreateRequest, requester string) (*connect.Response[pb.CreateResponse], error) {
	log := s.log.Named("createWithAutoAssign")
	log.Debug("Requester", zap.String("id", requester))

	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), accesspb.Level_ADMIN) {
		log.Warn("No root access")
		return nil, status.Error(codes.PermissionDenied, "No access rights")
	}
	account := req.Account

	acc, err := s.acc_ctrl.Get(ctx, account)
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	srvResp, err := s.srv_ctrl.List(ctx, requester, &servicespb.ListRequest{
		Filters: map[string]*structpb.Value{
			"account": structpb.NewStringValue(account),
		},
	})
	if err != nil {
		log.Error("Failed to retrieve services", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve services: %w", err)
	}
	services := srvResp.Result

	// Create service for user if it doesn't exist
	var srv *servicespb.Service
	srvCount := len(services)
	if srvCount > 1 {
		log.Info("Multiple services found for account. Transferring to first", zap.Int("count", srvCount))
	}
	if srvCount == 0 {
		log.Info("Account has no services, creating new")
		ns, err := s.acc_ctrl.GetNamespace(ctx, graph.Account{
			Account: &rpb.Account{
				Uuid: account,
			},
		})
		if err != nil {
			log.Error("Failed to get namespace", zap.Error(err))
			return nil, fmt.Errorf("failed to get namespace: %w", err)
		}
		if !s.ca.HasAccess(ctx, account, ns.ID, accesspb.Level_ADMIN) {
			log.Error("Destination account has no access to namespace")
			return nil, fmt.Errorf("destination account has no access to namespace")
		}
		if srv, err = s.srv_ctrl.Create(ctx, &servicespb.Service{
			Version: "1",
			Title:   "SRV_" + acc.GetTitle(),
		}); err != nil {
			log.Error("Failed to create service", zap.Error(err))
			return nil, fmt.Errorf("failed to create new service: %w", err)
		}
		if err = s.srv_ctrl.Join(ctx, srv, &ns, accesspb.Level_ADMIN, roles.OWNER); err != nil {
			log.Error("Error while creating access to service", zap.Error(err))
			return nil, fmt.Errorf("failed to create access to new service: %w", err)
		}
		if err = s.srv_ctrl.SetStatus(ctx, srv, spb.NoCloudStatus_UP); err != nil {
			log.Error("Failed to up service", zap.Error(err))
			return nil, fmt.Errorf("failed to up new service: %w", err)
		}
	} else {
		srv = services[0]
	}

	// Find instance group or create
	sp, err := s.sp_ctrl.Get(ctx, req.GetSp())
	if err != nil {
		log.Error("Failed to get service provider", zap.Error(err))
		return nil, fmt.Errorf("failed to obtain service provider: %w", err)
	}
	existingIGs := srv.GetInstancesGroups()
	for _, ig := range existingIGs {
		igSp, err := s.ig_ctrl.GetSP(ctx, ig.GetUuid())
		if err != nil {
			log.Error("Failed to get IG service provider", zap.Error(err))
			return nil, fmt.Errorf("failed to obtain IG's service provider: %w", err)
		}
		ig.Sp = &igSp.Uuid
	}

	var destIG *pb.InstancesGroup
	for _, ig := range existingIGs {
		ig.Instances = nil
		if ig.Data == nil {
			ig.Data = make(map[string]*structpb.Value)
		}
		if ig.GetSp() == sp.GetUuid() && !ig.Data["imported"].GetBoolValue() {
			destIG = ig
			break
		}
	}

	if destIG == nil {
		log.Info("Destination instances group not found, creating new")
		destIG = &pb.InstancesGroup{
			Type:  sp.GetType(),
			Title: acc.Title + " - " + sp.GetType(),
		}
		if err = s.ig_ctrl.Create(ctx, driver.NewDocumentID(schema.SERVICES_COL, srv.GetUuid()), destIG); err != nil {
			log.Error("Failed to create instances group", zap.Error(err))
			return nil, fmt.Errorf("failed to create new instances group: %w", err)
		}
		if err = s.ig_ctrl.Provide(ctx, destIG.GetUuid(), sp.GetUuid()); err != nil {
			log.Error("Failed to provide instances group", zap.Error(err))
			return nil, fmt.Errorf("failed to provide instances group for sp: %w", err)
		}
		if err := s.ig_ctrl.SetStatus(ctx, destIG, spb.NoCloudStatus_UP); err != nil {
			log.Error("Failed to up dest instance group", zap.Error(err))
			return nil, fmt.Errorf("failed to up new instances group: %w", err)
		}
	}

	newId, err := s.ctrl.Create(ctx, driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, destIG.GetUuid()), sp.GetUuid(), req.GetInstance())
	if err != nil {
		log.Error("Failed to create instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := s.ctrl.SetStatus(ctx, &pb.Instance{Uuid: newId}, spb.NoCloudStatus_UP); err != nil {
		log.Error("Failed to up created instance", zap.Error(err))
		return nil, fmt.Errorf("failed to up new instance: %w", err)
	}

	return connect.NewResponse(&pb.CreateResponse{
		Id:     newId,
		Result: true,
	}), nil
}

func (s *InstancesServer) Update(ctx context.Context, _req *connect.Request[pb.UpdateRequest]) (*connect.Response[pb.UpdateResponse], error) {
	log := s.log.Named("Update")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	var instance graph.Instance
	instance, err := s.ctrl.GetWithAccess(ctx, requestorId, req.GetInstance().GetUuid())
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	err = s.ctrl.Update(ctx, "", req.GetInstance(), instance.Instance)
	if err != nil {
		log.Error("Failed to update instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return connect.NewResponse(&pb.UpdateResponse{
		Result: true,
	}), nil
}

func (s *InstancesServer) Detach(ctx context.Context, _req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	log := s.log.Named("Detach")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	var instance graph.Instance
	instance, err := s.ctrl.GetWithAccess(ctx, requestorId, req.GetUuid())
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	err = s.ctrl.SetStatus(ctx, instance.Instance, spb.NoCloudStatus_DETACHED)
	if err != nil {
		return nil, err
	}

	var event = &elpb.Event{
		Entity:    schema.INSTANCES_COL,
		Uuid:      req.GetUuid(),
		Scope:     "database",
		Action:    "detach",
		Rc:        0,
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	nocloud.Log(log, event)

	return connect.NewResponse(&pb.DeleteResponse{
		Result: true,
	}), nil
}

func (s *InstancesServer) Attach(ctx context.Context, _req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	log := s.log.Named("Attach")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	var instance graph.Instance
	instance, err := s.ctrl.GetWithAccess(ctx, requestorId, req.GetUuid())
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	err = s.ctrl.SetStatus(ctx, instance.Instance, spb.NoCloudStatus_INIT)
	if err != nil {
		return nil, err
	}

	var event = &elpb.Event{
		Entity:    schema.INSTANCES_COL,
		Uuid:      req.GetUuid(),
		Scope:     "database",
		Action:    "detach",
		Rc:        0,
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	nocloud.Log(log, event)

	return connect.NewResponse(&pb.DeleteResponse{
		Result: true,
	}), nil
}

func (s *InstancesServer) TransferIG(ctx context.Context, _req *connect.Request[pb.TransferIGRequest]) (*connect.Response[pb.TransferIGResponse], error) {
	log := s.log.Named("transfer")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	igId := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, req.GetUuid())
	newSrvId := driver.NewDocumentID(schema.SERVICES_COL, req.GetService())
	ig, err := s.ig_ctrl.GetWithAccess(ctx, requestorId, igId.Key())
	if err != nil {
		log.Error("Failed to get instances group", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	srv, err := s.srv_ctrl.GetWithAccess(ctx, requestorId, newSrvId.Key())

	if err != nil {
		log.Error("Failed to get service", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if ig.GetAccess().GetLevel() < accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	if srv.GetAccess().GetLevel() < accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	srvEdge, err := s.ig_ctrl.GetEdge(ctx, igId.String(), schema.SERVICES_COL)

	if err != nil {
		log.Error("Failed to get Service", zap.Error(err))
		return nil, err
	}

	err = s.ig_ctrl.TransferIG(ctx, srvEdge, newSrvId, igId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.TransferIGResponse{
		Result: true,
		Meta:   nil,
	}), nil
}

//goland:noinspection GoErrorStringFormat
func (s *InstancesServer) TransferInstance(ctx context.Context, _req *connect.Request[pb.TransferInstanceRequest]) (*connect.Response[pb.TransferInstanceResponse], error) {
	log := s.log.Named("TransferInstance")
	req := _req.Msg
	log.Info("Request received", zap.Any("request", req))
	requester, _ := ctx.Value(nocloud.NoCloudAccount).(string)
	log = log.With(zap.Any("account", req.Account), zap.Any("ig", req.Ig), zap.String("requester", requester))

	if req.Account != nil && req.Ig != nil {
		return nil, fmt.Errorf("Account and IG cannot be set at the same time")
	}
	if req.Account == nil && req.Ig == nil {
		return nil, fmt.Errorf("Don't know where to transfer")
	}
	if !req.GetDoNotTransferInvoices() {
		return nil, fmt.Errorf("Invoices transfer feature is currently disabled!")
	}

	// Obtain service provider to use syncer
	groupResp, err := s.ctrl.GetGroup(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to get Group and ServicesProvider", zap.Error(err))
		return nil, fmt.Errorf("Failed to get SP and IG. Probably broken instance")
	}
	syncer := sync.NewDataSyncer(log.With(zap.String("caller", "TransferInstance")), s.rdb, groupResp.SP.GetUuid(), 50)
	defer syncer.Open()
	_ = syncer.WaitUntilOpenedAndCloseAfter()

	// Actually modified collections in transactions
	usedCols := []string{schema.SERVICES_COL, schema.INSTANCES_GROUPS_COL, schema.INVOICES_COL, schema.TRANSACTIONS_COL, schema.RECORDS_COL,
		schema.SERV2IG, schema.IG2INST, schema.IG2SP}
	// Collection connected in Permissions graph with some collections from usedCols
	gr, err := s.db.Graph(ctx, schema.PERMISSIONS_GRAPH.Name)
	if err != nil {
		log.Error("Failed to get permissions graph", zap.Error(err))
		return nil, fmt.Errorf("Internal error. Try again later")
	}
	edges := gr.EdgeDefinitions()
	otherCols := make([]string, 0)
	for _, e := range edges {
		otherCols = append(otherCols, e.Collection)
	}

	if req.Account != nil {
		trID, err := s.db.BeginTransaction(ctx, driver.TransactionCollections{
			Exclusive: usedCols,
			Write:     otherCols,
		}, &driver.BeginTransactionOptions{AllowImplicit: true})
		if err != nil {
			log.Error("Failed to start transaction", zap.Error(err))
			return nil, fmt.Errorf("Failed to perform transfer. Try again later")
		}
		ctx = driver.WithTransactionID(ctx, trID)

		defer func() {
			if panicErr := recover(); panicErr != nil {
				if err = s.db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{}); err != nil {
					log.Error("Failed to abort transaction", zap.Error(err))
				}
				log.Info("Recovered from panic", zap.Any("error", panicErr))
			}
		}()

		if err := s.transferToAccount(ctx, log, req.GetUuid(), req.GetAccount()); err != nil {
			if err := s.db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{}); err != nil {
				log.Error("Failed to abort transaction", zap.Error(err))
			}
			log.Error("Failed to transfer to account", zap.Error(err))
			return nil, fmt.Errorf("Failed to transfer to account. Error: " + err.Error())
		}

		if err = s.db.CommitTransaction(ctx, trID, &driver.CommitTransactionOptions{}); err != nil {
			log.Error("Failed to commit transaction", zap.Error(err))
			return nil, fmt.Errorf("Failed to perform transfer. Try again later")
		}
	} else if req.Ig != nil {
		if err := s.transferToIG(ctx, log, req.GetUuid(), req.GetIg()); err != nil {
			log.Error("Failed to transfer to IG", zap.Error(err))
			return nil, fmt.Errorf("Failed to transfer to IG. Error: " + err.Error())
		}
	}

	if req.GetDoNotTransferInvoices() {
		log.Info("Finished transfer. Not transferring invoices")
		return connect.NewResponse(&pb.TransferInstanceResponse{
			Result: true,
		}), nil
	}

	// Transfer invoices
	log.Info("Transferring invoices")
	errTmpl := "Instance was transferred, but failed to transfer invoices. Error: %w"
	acc, err := s.acc_ctrl.Get(ctx, req.GetAccount())
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return nil, fmt.Errorf(errTmpl, err)
	}
	accOwner, err := s.ctrl.GetInstanceOwner(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to get instance owner", zap.Error(err))
		return nil, fmt.Errorf(errTmpl, err)
	}
	// Transfer invoices (Ignoring PAID, TERM, CANCELLED and RETURNED invoices and take only that contains target instance and other instances that target account owns)
	invoices, err := s.inv_ctrl.List(ctx, accOwner.GetUuid())
	if err != nil {
		log.Error("Failed to list invoices", zap.Error(err))
		return nil, fmt.Errorf(errTmpl, err)
	}
	transferred := make([]*graph.Invoice, 0)
outer:
	for _, inv := range invoices {
		if inv.GetAccount() != accOwner.GetUuid() {
			log.Warn("Got invoice from other account. Must be incorrect List method filter", zap.String("account", inv.GetAccount()))
			continue
		}
		if inv.GetStatus() == billingpb.BillingStatus_PAID || inv.GetStatus() == billingpb.BillingStatus_RETURNED ||
			inv.GetStatus() == billingpb.BillingStatus_CANCELED || inv.GetStatus() == billingpb.BillingStatus_TERMINATED {
			continue
		}
		foundInstance := false
		for _, i := range inv.GetInstances() {
			if i == "" {
				continue
			}
			if i == req.GetUuid() {
				foundInstance = true
				continue
			}
			if !s.ca.HasAccess(ctx, acc.GetUuid(), driver.NewDocumentID(schema.INSTANCES_COL, i), accesspb.Level_ADMIN) {
				continue outer
			}
		}
		if !foundInstance {
			continue
		}
		if err = s.inv_ctrl.Transfer(ctx, inv.GetUuid(), acc.GetUuid(), acc.Currency); err != nil {
			log.Error("Failed to transfer invoice", zap.Error(err))
			return nil, fmt.Errorf(errTmpl, err)
		}
		if inv, err = s.inv_ctrl.Get(ctx, inv.GetUuid()); err != nil {
			log.Error("Failed to get invoice", zap.Error(err))
			return nil, fmt.Errorf(errTmpl, err)
		}
		transferred = append(transferred, inv)
	}
	// Sync with payment gateway
	gw := payments.GetPaymentGateway(acc.GetPaymentsGateway())
	_z := 0
	var success = &_z
	g := errgroup.Group{}
	m := &go_sync.Mutex{}
	for _, trInv := range transferred {
		invoice := trInv
		g.Go(func() error {
			if err = gw.CreateInvoice(ctx, invoice.Invoice); err != nil {
				return err
			}
			m.Lock()
			*success = *success + 1
			m.Unlock()
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		// If gateway data is untouched, then abort transferring
		if *success == 0 {
			return nil, fmt.Errorf(errTmpl, err)
		}
		log.Error("FATAL: Failed to sync with payment gateway, but managed to process some gateway invoices",
			zap.Error(err), zap.Int("processed", *success), zap.Int("total", len(transferred)))
	}

	log.Info("Finished transfer. Invoices were transferred", zap.Int("count", *success))
	return connect.NewResponse(&pb.TransferInstanceResponse{
		Result: true,
	}), nil
}

func (s *InstancesServer) AddNote(ctx context.Context, _req *connect.Request[notes.AddNoteRequest]) (*connect.Response[notes.NoteResponse], error) {
	log := s.log.Named("AddNote")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	var instance graph.Instance
	instance, err := s.ctrl.GetWithAccess(ctx, requestorId, req.Uuid)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() != accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	instance.AdminNotes = append(instance.AdminNotes, &notes.AdminNote{
		Admin:   requestor,
		Msg:     req.GetMsg(),
		Created: time.Now().Unix(),
	})

	err = s.ctrl.UpdateNotes(ctx, instance.Instance)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&notes.NoteResponse{
		Result:     true,
		AdminNotes: instance.GetAdminNotes(),
	}), nil
}

func (s *InstancesServer) PatchNote(ctx context.Context, _req *connect.Request[notes.PatchNoteRequest]) (*connect.Response[notes.NoteResponse], error) {
	log := s.log.Named("Patch")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	var instance graph.Instance
	instance, err := s.ctrl.GetWithAccess(ctx, requestorId, req.Uuid)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() != accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	ok := s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), accesspb.Level_ROOT)

	note := instance.GetAdminNotes()[req.GetIndex()]

	if requestor == note.GetAdmin() || ok {

		note.Admin = requestor
		note.Msg = req.GetMsg()
		note.Updated = time.Now().Unix()

		err = s.ctrl.UpdateNotes(ctx, instance.Instance)
		if err != nil {
			return nil, err
		}

		return connect.NewResponse(&notes.NoteResponse{
			Result:     true,
			AdminNotes: instance.GetAdminNotes(),
		}), nil
	}

	return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Instance notes")
}

func (s *InstancesServer) RemoveNote(ctx context.Context, _req *connect.Request[notes.RemoveNoteRequest]) (*connect.Response[notes.NoteResponse], error) {
	log := s.log.Named("Remove")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	var instance graph.Instance
	instance, err := s.ctrl.GetWithAccess(ctx, requestorId, req.Uuid)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() != accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	ok := s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), accesspb.Level_ROOT)

	note := instance.GetAdminNotes()[req.GetIndex()]
	log.Debug("Note", zap.Any("note", note))

	if requestor == note.GetAdmin() || ok {
		log.Debug("Notes before delete ", zap.Any("notes", instance.GetAdminNotes()))
		instance.AdminNotes = slices.Delete(instance.GetAdminNotes(), int(req.GetIndex()), int(req.GetIndex()+1))
		log.Debug("Notes after delete ", zap.Any("notes", instance.GetAdminNotes()))

		err = s.ctrl.UpdateNotes(ctx, instance.Instance)
		if err != nil {
			return nil, err
		}

		return connect.NewResponse(&notes.NoteResponse{
			Result:     true,
			AdminNotes: instance.GetAdminNotes(),
		}), nil
	}

	return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Instance notes")
}

func getFiltersQuery(filters map[string]*structpb.Value, bindVars map[string]interface{}) string {
	if len(filters) == 0 {
		return ""
	}

	query := ""
	for key, val := range filters {
		if key == "account" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER TO_STRING(acc._key) in @%s`, key)
			bindVars[key] = values
		} else if key == "search_param" {
			param := val.GetStringValue()
			query += fmt.Sprintf(` FILTER LOWER(node.title) LIKE LOWER("%s")
|| acc.data.email LIKE "%s"
|| acc._key LIKE "%s"
|| node._key LIKE "%s"
|| node.config.domain LIKE "%s"
|| node.data.vpsName LIKE "%s"
|| CONTAINS(TO_STRING(node.state.meta.networking.public), "%s") 
|| CONTAINS(TO_STRING(node.state.meta.networking.private), "%s")
|| CONTAINS(TO_STRING(node.data.ips_history.public), "%s")
|| CONTAINS(TO_STRING(node.data.ips_history.private), "%s")`,
				"%"+param+"%", "%"+param+"%", "%"+param+"%", "%"+param+"%", "%"+param+"%", "%"+param+"%", param, param, param, param)

		} else if key == "email" {
			query += fmt.Sprintf(` FILTER CONTAINS(acc.data.email, "%s")`, val.GetStringValue())
		} else if key == "title" {
			query += fmt.Sprintf(` FILTER CONTAINS(node.title, "%s")`, val.GetStringValue())
		} else if key == "resources.dcv" {
			query += fmt.Sprintf(` FILTER CONTAINS(node.resources.dcv, "%s")`, val.GetStringValue())
		} else if key == "resources.approver_email" {
			query += fmt.Sprintf(` FILTER CONTAINS(node.resources.approver_email, "%s")`, val.GetStringValue())
		} else if key == "config.domain" {
			query += fmt.Sprintf(` FILTER CONTAINS(node.config.domain, "%s")`, val.GetStringValue())
		} else if key == "config.configuration.vps_os" {
			query += fmt.Sprintf(` FILTER CONTAINS(node.config.configuration.vps_os, "%s")`, val.GetStringValue())
		} else if key == "type" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER ig.type in @%s`, key)
			bindVars[key] = values
		} else if key == "namespace" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER ns in @%s`, key)
			bindVars[key] = values
		} else if key == "sp" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER sp._key in @%s`, key)
			bindVars[key] = values
		} else if key == "billing_plan" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER bp._key in @%s`, key)
			bindVars[key] = values
		} else if key == "state.state" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			bindKey := "state"
			query += fmt.Sprintf(` FILTER node.state.state in @%s`, bindKey)
			bindVars[bindKey] = values
		} else if key == "estimate" {
			values := val.GetStructValue().AsMap()
			if val, ok := values["from"]; ok {
				from := val.(float64)
				query += fmt.Sprintf(` FILTER node.estimate >= %f`, from)
			}
			if val, ok := values["to"]; ok {
				to := val.(float64)
				query += fmt.Sprintf(` FILTER node.estimate <= %f`, to)
			}
		} else if key == "service" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER srv in @%s`, key)
			bindVars[key] = values
		} else if key == "period" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER node.period in @%s`, key)
			bindVars[key] = values
		} else if key == "config.location" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			bindKey := "location"
			query += fmt.Sprintf(` FILTER node.config.location in @%s`, bindKey)
			bindVars[bindKey] = values
		} else if key == "resources.cpu" {
			values := val.GetStructValue().AsMap()
			if val, ok := values["from"]; ok {
				from := int(val.(float64))
				query += fmt.Sprintf(` FILTER node.resources.cpu >= %d`, from)
			}
			if val, ok := values["to"]; ok {
				to := int(val.(float64))
				query += fmt.Sprintf(` FILTER node.resources.cpu <= %d`, to)
			}
		} else if key == "resources.drive_size" {
			values := val.GetStructValue().AsMap()
			if val, ok := values["from"]; ok {
				from := int(val.(float64))
				query += fmt.Sprintf(` FILTER node.resources.drive_size >= %d`, from)
			}
			if val, ok := values["to"]; ok {
				to := int(val.(float64))
				query += fmt.Sprintf(` FILTER node.resources.drive_size <= %d`, to)
			}
		} else if key == "resources.ram" {
			values := val.GetStructValue().AsMap()
			if val, ok := values["from"]; ok {
				from := int(val.(float64))
				query += fmt.Sprintf(` FILTER node.resources.ram >= %d`, from)
			}
			if val, ok := values["to"]; ok {
				to := int(val.(float64))
				query += fmt.Sprintf(` FILTER node.resources.ram <= %d`, to)
			}
		} else if key == "data.next_payment_date" {
			values := val.GetStructValue().AsMap()
			if val, ok := values["from"]; ok {
				from := int(val.(float64))
				query += fmt.Sprintf(` FILTER node.data.next_payment_date >= %d`, from)
			}
			if val, ok := values["to"]; ok {
				to := int(val.(float64))
				query += fmt.Sprintf(` FILTER node.data.next_payment_date <= %d`, to)
			}
		} else if key == "state.meta.networking" {
			val := val.GetStringValue()
			query += fmt.Sprintf(` FILTER CONTAINS(TO_STRING(node.state.meta.networking.public), "%s") 
|| CONTAINS(TO_STRING(node.state.meta.networking.private), "%s")
|| CONTAINS(TO_STRING(node.data.ips_history.public), "%s")
|| CONTAINS(TO_STRING(node.data.ips_history.private), "%s")`,
				val, val, val, val)
		} else if key == "product" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER node.product in @%s`, key)
			bindVars[key] = values
		} else if key == "created" {
			values := val.GetStructValue().AsMap()
			if val, ok := values["from"]; ok {
				from := int(val.(float64))
				query += fmt.Sprintf(` FILTER node.created >= %d`, from)
			}
			if val, ok := values["to"]; ok {
				to := int(val.(float64))
				query += fmt.Sprintf(` FILTER node.created <= %d`, to)
			}
		} else if key == "uuids" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER node._key in @%s`, key)
			bindVars[key] = values
		} else if key == "status" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER node.status in @%s`, "statusFilter")
			bindVars["statusFilter"] = values
		} else if key == "started" {
			keyVal := "startedFilter"
			query += fmt.Sprintf(` FILTER TO_BOOL(node.config.auto_start) == @%s`, keyVal)
			bindVars[keyVal] = val.GetBoolValue()
		}
	}

	return query
}

func getSortQuery(field, order string, customOrder []interface{}, bindVars map[string]interface{}) string {
	if field == "" || order == "" {
		return ""
	}
	query := ""

	if field == "account" {
		query += fmt.Sprintf(" SORT acc.title %s", order)
	} else if field == "email" {
		query += fmt.Sprintf(" SORT acc.data.email %s", order)
	} else if field == "balance" {
		query += fmt.Sprintf(" SORT TO_NUMBER(acc.balance) %s", order)
	} else if field == "type" {
		query += fmt.Sprintf(" SORT bp.type %s", order)
	} else if field == "billing_plan" {
		query += fmt.Sprintf(" SORT bp.title %s", order)
	} else if field == "sp" {
		query += fmt.Sprintf(" SORT sp.title %s", order)
	} else if field == "config.domain" {
		query += fmt.Sprintf(" SORT node.config.domain %s", order)
	} else if field == "state.state" {
		bindKey := "customOrder"
		query += fmt.Sprintf(" SORT POSITION(@%s, node.state.state, true) DESC", bindKey)
		bindVars[bindKey] = []interface{}{stpb.NoCloudState_RUNNING, stpb.NoCloudState_SUSPENDED, stpb.NoCloudState_INIT, stpb.NoCloudState_PENDING,
			stpb.NoCloudState_FAILURE, stpb.NoCloudState_OPERATION, stpb.NoCloudState_STOPPED, stpb.NoCloudState_UNKNOWN, stpb.NoCloudState_DELETED}
	} else {
		query += fmt.Sprintf(" SORT node.%s %s", field, order)
	}

	return query
}

const listInstancesQuery = `
LET instances = (
	FOR node, edge, path IN 0..10 OUTBOUND @account_node
	    GRAPH @permissions_graph
	    OPTIONS {order: "bfs", uniqueVertices: "global"}
	    FILTER IS_SAME_COLLECTION(@instances, node)

        LET ig = DOCUMENT(path.vertices[-2]._id)
        LET sp = LAST (
            FOR sp_node IN 1 OUTBOUND ig
	            GRAPH @permissions_graph
	            OPTIONS {order: "bfs", uniqueVertices: "global"}
	            FILTER IS_SAME_COLLECTION(@service_provider, sp_node)
	            RETURN sp_node
        )
        LET srv = path.vertices[-3]._key
        LET ns = path.vertices[-4]._key
        LET acc = DOCUMENT(CONCAT(@accounts, "/", path.vertices[-5]._key))
		LET bp = DOCUMENT(CONCAT(@bps, "/", node.billing_plan.uuid))
		
		%s
		
		RETURN {
			instance: MERGE(node, { 
				uuid: node._key, 
				billing_plan: {
					uuid: bp._key,
					title: bp.title,
					type: bp.type,
					kind: bp.kind,
					resources: bp.resources,
					products: {
						[node.product]: bp.products[node.product],
					},
					meta: bp.meta,
					fee: bp.fee,
					software: bp.software 
				}
			}
			),
			service: srv,
			sp: sp._key,
			type: sp.type,
			account: acc._key,
            namespace: ns
		}
)

return { 
	pool: (@limit > 0) ? SLICE(instances, @offset, @limit) : instances,
	count: LENGTH(instances)
}
`

func (s *InstancesServer) List(ctx context.Context, _req *connect.Request[pb.ListInstancesRequest]) (*connect.Response[pb.ListInstancesResponse], error) {
	log := s.log.Named("List")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))
	log.Debug("Request", zap.Any("req", req))

	limit, page := req.GetLimit(), req.GetPage()
	offset := (page - 1) * limit

	var query string
	bindVars := map[string]interface{}{
		"account_node":      driver.NewDocumentID(schema.ACCOUNTS_COL, requestor),
		"permissions_graph": schema.PERMISSIONS_GRAPH.Name,
		"instances":         schema.INSTANCES_COL,
		"accounts":          schema.ACCOUNTS_COL,
		"bps":               schema.BILLING_PLANS_COL,
		"service_provider":  schema.SERVICES_PROVIDERS_COL,
		"offset":            offset,
		"limit":             limit,
	}

	customOrder := req.GetCustomOrder().GetListValue().AsSlice()
	slices.Reverse(customOrder)
	query += getSortQuery(req.GetField(), req.GetSort(), customOrder, bindVars)

	query += getFiltersQuery(req.GetFilters(), bindVars)

	query = fmt.Sprintf(listInstancesQuery, query)

	s.log.Debug("Query", zap.Any("q", query))
	s.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	c, err := s.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var result pb.ListInstancesResponse
	_, err = c.ReadDocument(ctx, &result)
	if err != nil {
		return nil, err
	}

	conv := graph.NewConverter(_req.Header(), s.curr_ctrl)
	wg := &go_sync.WaitGroup{}
	for _, value := range result.Pool {
		value := value
		if value.Instance == nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			value.Instance.Period, _ = s.ctrl.GetInstancePeriod(value.Instance)
			var oneTime bool
			if value.Instance.Period != nil && *value.Instance.Period == 0 {
				oneTime = true
			}
			value.Instance.Estimate, _ = s.ctrl.CalculateInstanceEstimatePrice(value.Instance, oneTime)
			conv.ConvertObjectPrices(value.Instance)
		}()
	}
	wg.Wait()

	log.Debug("Result", zap.Any("result", &result))
	resp := connect.NewResponse(&result)
	conv.SetResponseHeader(resp.Header())
	return graph.HandleConvertionError(resp, conv)
}

const countInstancesQuery = `
LET instances = (
	FOR node, edge, path IN 0..10 OUTBOUND @account
	    GRAPH @permissions_graph
	    OPTIONS {order: "bfs", uniqueVertices: "global"}
	    FILTER IS_SAME_COLLECTION(@instances, node)

        LET ig = DOCUMENT(path.vertices[-2]._id)
        LET sp = LAST (
            FOR sp_node IN 1 OUTBOUND ig
	            GRAPH @permissions_graph
	            OPTIONS {order: "bfs", uniqueVertices: "global"}
	            FILTER IS_SAME_COLLECTION(@service_provider, sp_node)
	            RETURN sp_node
        )
        LET srv = path.vertices[-3]._key
        LET ns = path.vertices[-4]._key
        LET acc = DOCUMENT(CONCAT(@accounts, "/", path.vertices[-5]._key))
		LET bp = DOCUMENT(CONCAT(@bps, "/", node.billing_plan.uuid))
		
		%s
		
		RETURN {
			instance: MERGE(node, { 
				uuid: node._key, 
				billing_plan: {
					uuid: bp._key,
					title: bp.title,
					type: bp.type,
					kind: bp.kind,
					resources: bp.resources,
					products: {
						[node.product]: bp.products[node.product],
					},
					meta: bp.meta,
					fee: bp.fee,
					software: bp.software 
				}
			}
			),
			service: srv,
			sp: sp,
			type: bp.type
		}
)

let locations = (
 FOR inst IN instances
 FILTER inst.instance.config.location
 FILTER inst.instance.config.location != ""
 RETURN DISTINCT inst.instance.config.location
)

let periods = (
 FOR inst IN instances
 FILTER inst.instance.period
 RETURN DISTINCT inst.instance.period
)

let products = (
 FOR inst IN instances
 FILTER inst.instance.product
 FILTER inst.instance.product != ""
 RETURN DISTINCT inst.instance.product
)

let billing_plans = (
 FOR inst IN instances
 FILTER inst.instance.billing_plan
 FILTER inst.instance.billing_plan.uuid
 COLLECT uuid = inst.instance.billing_plan.uuid, title = inst.instance.billing_plan.title
 RETURN { uuid, title }
)

let service_providers = (
 FOR inst IN instances
 FILTER inst.sp
 COLLECT uuid = inst.sp.uuid, title = inst.sp.title
 RETURN { uuid, title }
)

return { 
	unique: {
        locations: locations,
        products: products,
        periods: periods,
		billing_plans: billing_plans,
        service_providers: service_providers
	},
	total: LENGTH(instances)
}
`

func (s *InstancesServer) GetUnique(ctx context.Context, _req *connect.Request[pb.GetUniqueRequest]) (*connect.Response[pb.GetUniqueResponse], error) {
	log := s.log.Named("GetCount")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	type Response struct {
		Total  int                    `json:"total"`
		Unique map[string]interface{} `json:"unique"`
	}

	var query string
	bindVars := map[string]interface{}{
		"account":           driver.NewDocumentID(schema.ACCOUNTS_COL, requestor),
		"permissions_graph": schema.PERMISSIONS_GRAPH.Name,
		"instances":         schema.INSTANCES_COL,
		"accounts":          schema.ACCOUNTS_COL,
		"bps":               schema.BILLING_PLANS_COL,
		"service_provider":  schema.SERVICES_PROVIDERS_COL,
	}

	query += getFiltersQuery(req.GetFilters(), bindVars)

	query = fmt.Sprintf(countInstancesQuery, query)

	s.log.Debug("Query", zap.Any("q", query))
	s.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	c, err := s.db.Query(ctx, query, bindVars)
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

	var result pb.GetUniqueResponse
	obj, err := structpb.NewStruct(resp.Unique)
	if err != nil {
		return nil, err
	}
	result.Unique = structpb.NewStructValue(obj)
	result.Total = uint64(resp.Total)

	return connect.NewResponse(&result), nil
}

const getInstanceQuery = `
LET instances = (
	FOR node, edge, path IN 0..10 OUTBOUND @account
	    GRAPH @permissions_graph
	    OPTIONS {order: "bfs", uniqueVertices: "global"}
	    FILTER IS_SAME_COLLECTION(@instances, node)

		FILTER node._key == @uuid

        LET ig = DOCUMENT(path.vertices[-2]._id)
        LET sp = LAST (
            FOR sp_node IN 1 OUTBOUND ig
	            GRAPH @permissions_graph
	            OPTIONS {order: "bfs", uniqueVertices: "global"}
	            FILTER IS_SAME_COLLECTION(@service_provider, sp_node)
	            RETURN sp_node
        )
        LET srv = path.vertices[-3]._key
        LET ns = path.vertices[-4]._key
        LET acc = DOCUMENT(CONCAT(@accounts, "/", path.vertices[-5]._key))
		LET bp = DOCUMENT(CONCAT(@bps, "/", node.billing_plan.uuid))
		
		%s
		
		RETURN {
			instance: MERGE(node, { 
				uuid: node._key, 
				billing_plan: {
					uuid: bp._key,
					title: bp.title,
					type: bp.type,
					kind: bp.kind,
					resources: bp.resources,
					products: {
						[node.product]: bp.products[node.product],
					},
					meta: bp.meta,
					fee: bp.fee,
					software: bp.software 
				}
			}
			),
			service: srv,
			sp: sp._key,
			type: sp.type,
			account: acc._key,
            namespace: ns
		}
)

FILTER LENGTH(instances) > 0
RETURN instances[0]
`

func (s *InstancesServer) Get(ctx context.Context, _req *connect.Request[pb.Instance]) (*connect.Response[pb.ResponseInstance], error) {
	log := s.log.Named("Get")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	var query string
	bindVars := map[string]interface{}{
		"account":           driver.NewDocumentID(schema.ACCOUNTS_COL, requestor),
		"permissions_graph": schema.PERMISSIONS_GRAPH.Name,
		"instances":         schema.INSTANCES_COL,
		"accounts":          schema.ACCOUNTS_COL,
		"bps":               schema.BILLING_PLANS_COL,
		"service_provider":  schema.SERVICES_PROVIDERS_COL,
		"uuid":              req.Uuid,
	}

	s.log.Debug("Query", zap.Any("q", query))
	s.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	query = fmt.Sprintf(getInstanceQuery, query)

	c, err := s.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var result pb.ResponseInstance
	_, err = c.ReadDocument(ctx, &result)
	if err != nil {
		return nil, err
	}

	conv := graph.NewConverter(_req.Header(), s.curr_ctrl)
	if result.Instance != nil {
		var oneTime bool
		result.Instance.Period, _ = s.ctrl.GetInstancePeriod(result.Instance)
		if result.Instance.Period != nil && *result.Instance.Period == 0 {
			oneTime = true
		}
		result.Instance.Estimate, _ = s.ctrl.CalculateInstanceEstimatePrice(result.Instance, oneTime)
		conv.ConvertObjectPrices(result.Instance)
	}

	resp := connect.NewResponse(&result)
	conv.SetResponseHeader(resp.Header())
	return graph.HandleConvertionError(resp, conv)
}

func (s *InstancesServer) transferToAccount(ctx context.Context, log *zap.Logger, uuid string, account string) (err error) {
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	requesterId := driver.NewDocumentID(schema.ACCOUNTS_COL, requester)
	instanceId := driver.NewDocumentID(schema.INSTANCES_COL, uuid)
	accountId := driver.NewDocumentID(schema.ACCOUNTS_COL, account)

	if !s.ca.HasAccess(ctx, requester, instanceId, accesspb.Level_ADMIN) {
		return fmt.Errorf("no access to instance")
	}
	if !s.ca.HasAccess(ctx, requester, accountId, accesspb.Level_ADMIN) {
		return fmt.Errorf("no access to destination account")
	}

	inst, err := s.ctrl.GetWithAccess(ctx, requesterId, uuid)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return fmt.Errorf("failed to get instance: %w", err)
	}

	state := inst.GetState().GetState()
	forbiddenStates := []stpb.NoCloudState{stpb.NoCloudState_UNKNOWN, stpb.NoCloudState_INIT,
		stpb.NoCloudState_DELETED, stpb.NoCloudState_OPERATION, stpb.NoCloudState_FAILURE}
	if slices.Contains(forbiddenStates, state) {
		log.Error("Instance in forbidden state to transfer it to other account", zap.Any("state", state))
		return fmt.Errorf("instance in forbidden state. Can't transfer while instance in state: %s", state.String())
	}

	accOwner, err := s.ctrl.GetInstanceOwner(ctx, uuid)
	if err != nil {
		log.Error("Failed to find instance owner", zap.Error(err))
		return fmt.Errorf("failed to find instance owner: %w", err)
	}
	if accOwner.GetUuid() == account {
		return fmt.Errorf("can't transfer. User already owns it")
	}

	srvResp, err := s.srv_ctrl.List(ctx, requester, &servicespb.ListRequest{
		Filters: map[string]*structpb.Value{
			"account": structpb.NewStringValue(account),
		},
	})
	if err != nil {
		log.Error("Failed to retrieve services", zap.Error(err))
		return fmt.Errorf("failed to retrieve services: %w", err)
	}
	services := srvResp.Result

	groupResp, err := s.ctrl.GetGroup(ctx, instanceId.String())
	if err != nil {
		log.Error("Failed to get Group and ServicesProvider", zap.Error(err))
		return fmt.Errorf("failed to get Group and ServicesProvider: %w", err)
	}
	if groupResp == nil {
		log.Error("Failed to get instance linked data")
		return fmt.Errorf("failed to get instance linked data. Probably broken instance")
	}
	if groupResp.SP == nil || groupResp.SP.GetUuid() == "" {
		log.Error("SP not found, instance not linked")
		return fmt.Errorf("SP not found, instance not linked. Probably broken instance")
	}
	if groupResp.Group == nil || groupResp.Group.GetUuid() == "" {
		log.Error("Group not found, instance not linked")
		return fmt.Errorf("IG not found, instance not linked. Probably broken instance")
	}
	oldIG := groupResp.Group
	oldIGInstancesCount := len(oldIG.GetInstances())
	oldIG.Instances = nil
	sp := groupResp.SP

	acc, err := s.acc_ctrl.Get(ctx, account)
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Create service for user if it doesn't exist
	var srv *servicespb.Service
	srvCount := len(services)
	if srvCount > 1 {
		log.Info("Multiple services found for account. Transferring to first", zap.Int("count", srvCount))
	}
	if srvCount == 0 {
		log.Info("Account has no services, creating new")
		ns, err := s.acc_ctrl.GetNamespace(ctx, graph.Account{
			Account: &rpb.Account{
				Uuid: account,
			},
		})
		if err != nil {
			log.Error("Failed to get namespace", zap.Error(err))
			return fmt.Errorf("failed to get namespace: %w", err)
		}
		if !s.ca.HasAccess(ctx, account, ns.ID, accesspb.Level_ADMIN) {
			log.Error("Destination account has no access to namespace")
			return fmt.Errorf("destination account has no access to namespace")
		}
		if srv, err = s.srv_ctrl.Create(ctx, &servicespb.Service{
			Version: "1",
			Title:   "SRV_" + acc.GetTitle(),
		}); err != nil {
			log.Error("Failed to create service", zap.Error(err))
			return fmt.Errorf("failed to create new service: %w", err)
		}
		if err = s.srv_ctrl.Join(ctx, srv, &ns, accesspb.Level_ADMIN, roles.OWNER); err != nil {
			log.Error("Error while creating access to service", zap.Error(err))
			return fmt.Errorf("failed to create access to new service: %w", err)
		}
		if err = s.srv_ctrl.SetStatus(ctx, srv, spb.NoCloudStatus_UP); err != nil {
			log.Error("Failed to up service", zap.Error(err))
			return fmt.Errorf("failed to up new service: %w", err)
		}
	} else {
		srv = services[0]
	}

	existingIGs := srv.GetInstancesGroups()
	// Ensure each IG object has 'sp' field set as current SP
	for _, ig := range existingIGs {
		igSp, err := s.ig_ctrl.GetSP(ctx, ig.GetUuid())
		if err != nil {
			log.Error("Failed to get IG service provider", zap.Error(err))
			return fmt.Errorf("failed to obtain IG's service provider: %w", err)
		}
		ig.Sp = &igSp.Uuid
	}

	igEdge, err := s.ctrl.GetEdge(ctx, instanceId.String(), schema.INSTANCES_GROUPS_COL)
	if err != nil {
		log.Error("Failed to get instances group edge", zap.Error(err))
		return fmt.Errorf("failed to get link between instance and old IG: %w", err)
	}

	var destIG *pb.InstancesGroup
	var _old *pb.InstancesGroup
	for _, ig := range existingIGs {
		ig.Instances = nil
		if oldIG.GetType() == ig.GetType() && ig.GetSp() == sp.GetUuid() {
			destIG = ig
			break
		}
	}

	// TODO: this explicit driver type check shouldn't be here, but it is fucking hard to fix for nebula IG
	if oldIG.GetType() == nebulaIGType {
		log.Warn("Transferring Ione instances group explicitly")
		if err = s.processNebulaIG(ctx, log, inst.Instance, oldIG, oldIGInstancesCount, existingIGs, srv, acc.GetTitle(), sp.GetUuid(), igEdge); err != nil {
			log.Error("Failed to process Ione instances group", zap.Error(err))
			return fmt.Errorf("error processing IONE: %w", err)
		}
		return nil
	}

	if destIG == nil && oldIGInstancesCount == 1 {
		log.Info("Transferring instances group directly to service")
		oldIGid := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, oldIG.GetUuid())
		srvEdge, err := s.ig_ctrl.GetEdge(ctx, oldIGid.String(), schema.SERVICES_COL)
		if err != nil {
			log.Error("Failed to get service->IG edge", zap.Error(err))
			return fmt.Errorf("failed to get link between old IG and service: %w", err)
		}
		newSrvId := driver.NewDocumentID(schema.SERVICES_COL, srv.GetUuid())
		if err = s.ig_ctrl.TransferIG(ctx, srvEdge, newSrvId, oldIGid); err != nil {
			log.Error("Failed to transfer old instances group to service", zap.Error(err))
			return fmt.Errorf("failed to transfer old instances group: %w", err)
		}
		return nil

	} else if destIG == nil {
		log.Info("Destination instances group not found, creating new")
		destIG = &pb.InstancesGroup{
			Type:  oldIG.GetType(),
			Title: acc.Title + " - " + oldIG.GetType() + " - imported",
		}
		if err = s.ig_ctrl.Create(ctx, driver.NewDocumentID(schema.SERVICES_COL, srv.GetUuid()), destIG); err != nil {
			log.Error("Failed to create instances group", zap.Error(err))
			return fmt.Errorf("failed to create new instances group: %w", err)
		}
		if err = s.ig_ctrl.Provide(ctx, destIG.GetUuid(), sp.GetUuid()); err != nil {
			log.Error("Failed to provide instances group", zap.Error(err))
			return fmt.Errorf("failed to provide instances group for sp: %w", err)
		}
		if err := s.ig_ctrl.SetStatus(ctx, destIG, spb.NoCloudStatus_UP); err != nil {
			log.Error("Failed to up dest instance group", zap.Error(err))
			return fmt.Errorf("failed to up new instances group: %w", err)
		}
	}
	// Process new IG ips
	_old = proto.Clone(destIG).(*pb.InstancesGroup)
	destIG = processIGsIPs(destIG, inst.Instance, false)
	if err = s.ig_ctrl.Update(ctx, destIG, _old); err != nil {
		log.Error("Failed to update new instances group", zap.Error(err))
		return fmt.Errorf("failed to update new instances group: %w", err)
	}
	// Process old IG ips
	_old = proto.Clone(oldIG).(*pb.InstancesGroup)
	oldIG = processIGsIPs(oldIG, inst.Instance, true)
	if err = s.ig_ctrl.Update(ctx, oldIG, _old); err != nil {
		log.Error("Failed to update old instances group", zap.Error(err))
		return fmt.Errorf("failed to update old instances group: %w", err)
	}

	if err = s.ctrl.TransferInst(ctx, igEdge, driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, destIG.GetUuid()), instanceId); err != nil {
		log.Error("Failed to transfer instance", zap.Error(err))
		return fmt.Errorf("failed to transfer instance: %w", err)
	}

	return nil
}

func (s *InstancesServer) transferToIG(ctx context.Context, log *zap.Logger, uuid string, igUuid string) error {
	return fmt.Errorf("transfer to IG option currently unavailable")
	//
	//requester := ctx.Value(nocloud.NoCloudAccount).(string)
	//requesterId := driver.NewDocumentID(schema.ACCOUNTS_COL, requester)
	//
	//instanceId := driver.NewDocumentID(schema.INSTANCES_COL, uuid)
	//newIGId := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, igUuid)
	//inst, err := s.ctrl.GetWithAccess(ctx, requesterId, uuid)
	//if err != nil {
	//	log.Error("Failed to get instance", zap.Error(err))
	//	return fmt.Errorf("failed to get instance: %w", err)
	//}
	//
	//ig, err := s.ig_ctrl.GetWithAccess(ctx, requesterId, newIGId.Key())
	//if err != nil {
	//	log.Error("Failed to get instances group", zap.Error(err))
	//	return fmt.Errorf("failed to get instances group: %w", err)
	//}
	//
	//if inst.GetAccess().GetLevel() < accesspb.Level_ROOT {
	//	log.Error("Access denied", zap.String("uuid", inst.GetUuid()))
	//	return fmt.Errorf("no access to instance")
	//}
	//
	//if ig.GetAccess().GetLevel() < accesspb.Level_ROOT {
	//	log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
	//	return fmt.Errorf("no access to IG")
	//}
	//
	//igEdge, err := s.ctrl.GetEdge(ctx, instanceId.String(), schema.INSTANCES_GROUPS_COL)
	//
	//if err != nil {
	//	log.Error("Failed to get IG edge", zap.Error(err))
	//	return fmt.Errorf("failed to get IG connection: %w", err)
	//}
	//
	//err = s.ctrl.TransferInst(ctx, igEdge, newIGId, instanceId)
	//if err != nil {
	//	log.Error("Failed to transfer instance", zap.Error(err))
	//	return fmt.Errorf("failed to transfer: %w", err)
	//}
	//
	//return nil
}

func (s *InstancesServer) processNebulaIG(ctx context.Context, log *zap.Logger, inst *pb.Instance, oldIG *pb.InstancesGroup, oldIGInstancesCount int,
	igs []*pb.InstancesGroup, srv *servicespb.Service, accTitle, spUuid, oldIGEdgeKey string) error {
	log = log.Named("processNebulaIG")

	userid := int(oldIG.GetData()["userid"].GetNumberValue())
	publicVn := int(oldIG.GetData()["public_vn"].GetNumberValue())
	privateVn := int(oldIG.GetData()["private_vn"].GetNumberValue())
	publicIpsFree := int(oldIG.GetData()["public_ips_free"].GetNumberValue())
	privateIpsFree := int(oldIG.GetData()["private_ips_free"].GetNumberValue())
	publicIpsTotal := int(oldIG.GetData()["public_ips_total"].GetNumberValue())
	privateIpsTotal := int(oldIG.GetData()["private_ips_total"].GetNumberValue())

	ipsPublic := int(oldIG.GetResources()["ips_public"].GetNumberValue())
	ipsPrivate := int(oldIG.GetResources()["ips_private"].GetNumberValue())
	// Checks
	//ipsPublic := int(oldIG.GetResources()["ips_public"].GetNumberValue())
	//ipsPrivate := int(oldIG.GetResources()["ips_private"].GetNumberValue())
	//if (ipsPublic != publicIpsFree || ipsPublic != publicIpsTotal) ||
	//	(ipsPrivate != privateIpsFree || ipsPrivate != privateIpsTotal) {
	//	return nil, fmt.Errorf("can't transfer. IONE cluster currently in unstable state")
	//}
	if inst.GetState().GetState() == stpb.NoCloudState_PENDING {
		return fmt.Errorf("can't transfer IONE instance in PENDING state")
	}
	if publicIpsFree != publicIpsTotal || privateIpsFree != privateIpsTotal {
		return fmt.Errorf("can't transfer. IONE cluster currently in unstable state")
	}
	if userid == 0 {
		return fmt.Errorf("can't transfer. IONE cluster currently not setted up")
	}

	var destIG *pb.InstancesGroup
	for _, ig := range igs {
		if oldIG.GetType() != ig.GetType() || ig.GetSp() != spUuid {
			continue
		}
		if int(ig.GetData()["userid"].GetNumberValue()) != userid ||
			int(ig.GetData()["public_vn"].GetNumberValue()) != publicVn ||
			int(ig.GetData()["private_vn"].GetNumberValue()) != privateVn {
			continue
		}
		destIG = ig
		break
	}

	if destIG == nil && oldIGInstancesCount == 1 {
		log.Info("Transferring instances group directly to service")
		oldIGid := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, oldIG.GetUuid())
		srvEdge, err := s.ig_ctrl.GetEdge(ctx, oldIGid.String(), schema.SERVICES_COL)
		if err != nil {
			log.Error("Failed to get service->IG edge", zap.Error(err))
			return fmt.Errorf("failed to get link between old IG and service: %w", err)
		}
		newSrvId := driver.NewDocumentID(schema.SERVICES_COL, srv.GetUuid())
		if err = s.ig_ctrl.TransferIG(ctx, srvEdge, newSrvId, oldIGid); err != nil {
			log.Error("Failed to transfer old instances group to service", zap.Error(err))
			return fmt.Errorf("failed to transfer old instances group: %w", err)
		}
		return nil

	} else if destIG == nil {
		log.Info("Destination instances group not found, creating new")
		destIG = &pb.InstancesGroup{
			Type:  oldIG.GetType(),
			Title: accTitle + " - " + oldIG.GetType() + " - imported",
			Data: map[string]*structpb.Value{
				"userid": structpb.NewNumberValue(float64(userid)),
			},
		}
		if publicVn > 0 {
			destIG.Data["public_vn"] = structpb.NewNumberValue(float64(publicVn))
		}
		if privateVn > 0 {
			destIG.Data["private_vn"] = structpb.NewNumberValue(float64(privateVn))
		}
		if err := s.ig_ctrl.Create(ctx, driver.NewDocumentID(schema.SERVICES_COL, srv.GetUuid()), destIG); err != nil {
			log.Error("Failed to create instances group", zap.Error(err))
			return fmt.Errorf("failed to create new instances group: %w", err)
		}
		if err := s.ig_ctrl.Provide(ctx, destIG.GetUuid(), spUuid); err != nil {
			log.Error("Failed to provide instances group", zap.Error(err))
			return fmt.Errorf("failed to provide new instances group for sp: %w", err)
		}
		if err := s.ig_ctrl.SetStatus(ctx, destIG, spb.NoCloudStatus_UP); err != nil {
			log.Error("Failed to up dest instance group", zap.Error(err))
			return fmt.Errorf("failed to up new instances group: %w", err)
		}
	}
	if destIG.Data == nil {
		destIG.Data = make(map[string]*structpb.Value)
	}
	if destIG.Resources == nil {
		destIG.Resources = make(map[string]*structpb.Value)
	}
	if oldIG.Data == nil {
		oldIG.Data = make(map[string]*structpb.Value)
	}
	if oldIG.Resources == nil {
		oldIG.Resources = make(map[string]*structpb.Value)
	}

	ipsPublicNew := int(destIG.GetResources()["ips_public"].GetNumberValue())
	ipsPrivateNew := int(destIG.GetResources()["ips_private"].GetNumberValue())
	// Process new IG
	_old := proto.Clone(destIG).(*pb.InstancesGroup)
	destIG = processIGsIPs(destIG, inst, false)
	differPublic := int(destIG.GetResources()["ips_public"].GetNumberValue()) - ipsPublicNew
	differPrivate := int(destIG.GetResources()["ips_private"].GetNumberValue()) - ipsPrivateNew
	destIG.Data["public_ips_free"] = structpb.NewNumberValue(float64(int(destIG.Data["public_ips_free"].GetNumberValue()) + differPublic))
	destIG.Data["public_ips_total"] = structpb.NewNumberValue(float64(int(destIG.Data["public_ips_total"].GetNumberValue()) + differPublic))
	destIG.Data["private_ips_free"] = structpb.NewNumberValue(float64(int(destIG.Data["private_ips_free"].GetNumberValue()) + differPrivate))
	destIG.Data["private_ips_total"] = structpb.NewNumberValue(float64(int(destIG.Data["private_ips_total"].GetNumberValue()) + differPrivate))
	destIG.Data["imported"] = structpb.NewBoolValue(true)
	if err := s.ig_ctrl.Update(ctx, destIG, _old, true); err != nil {
		log.Error("Failed to update instances group", zap.Error(err))
		return fmt.Errorf("failed to update new instances group: %w", err)
	}
	// Process old IG
	_old = proto.Clone(oldIG).(*pb.InstancesGroup)
	oldIG = processIGsIPs(oldIG, inst, true)
	differPublic = int(oldIG.GetResources()["ips_public"].GetNumberValue()) - ipsPublic
	differPrivate = int(oldIG.GetResources()["ips_private"].GetNumberValue()) - ipsPrivate
	oldIG.Data["public_ips_free"] = structpb.NewNumberValue(float64(int(oldIG.Data["public_ips_free"].GetNumberValue()) + differPublic))
	oldIG.Data["public_ips_total"] = structpb.NewNumberValue(float64(int(oldIG.Data["public_ips_total"].GetNumberValue()) + differPublic))
	oldIG.Data["private_ips_free"] = structpb.NewNumberValue(float64(int(oldIG.Data["private_ips_free"].GetNumberValue()) + differPrivate))
	oldIG.Data["private_ips_total"] = structpb.NewNumberValue(float64(int(oldIG.Data["private_ips_total"].GetNumberValue()) + differPrivate))
	if err := s.ig_ctrl.Update(ctx, oldIG, _old, true); err != nil {
		log.Error("Failed to update instances group", zap.Error(err))
		return fmt.Errorf("failed to update old instances group: %w", err)
	}

	instanceId := driver.NewDocumentID(schema.INSTANCES_COL, inst.GetUuid())
	if err := s.ctrl.TransferInst(ctx, oldIGEdgeKey, driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, destIG.GetUuid()), instanceId); err != nil {
		log.Error("Failed to transfer instance", zap.Error(err))
		return fmt.Errorf("failed to transfer instance: %w", err)
	}

	return nil
}

func processIGsIPs(ig *pb.InstancesGroup, inst *pb.Instance, decrease bool) *pb.InstancesGroup {
	if ig == nil {
		return nil
	}
	if inst == nil {
		return ig
	}

	plusOrZero := func(n int) int {
		if n < 0 {
			return 0
		}
		return n
	}

	var (
		oldAmount int
		newAmount int
	)

	mul := 1
	if decrease {
		mul = -1
	}

	if ig.Resources == nil {
		ig.Resources = make(map[string]*structpb.Value)
	}

	oldAmount = int(ig.GetResources()["ips_public"].GetNumberValue())
	newAmount = oldAmount + int(inst.GetResources()["ips_public"].GetNumberValue())*mul
	if oldAmount != newAmount {
		ig.Resources["ips_public"] = structpb.NewNumberValue(float64(plusOrZero(newAmount)))
	}

	oldAmount = int(ig.GetResources()["ips_private"].GetNumberValue())
	newAmount = oldAmount + int(inst.GetResources()["ips_private"].GetNumberValue())*mul
	if oldAmount != newAmount {
		ig.Resources["ips_private"] = structpb.NewNumberValue(float64(plusOrZero(newAmount)))
	}

	return ig
}
