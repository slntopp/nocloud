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
	"github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"github.com/slntopp/nocloud/pkg/nocloud/sync"
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

	ctrl    *graph.InstancesController
	ig_ctrl *graph.InstancesGroupsController
	sp_ctrl *graph.ServicesProvidersController

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
		ctrl:    ig_ctrl.Instances(),
		ig_ctrl: ig_ctrl,
		sp_ctrl: &sp_ctrl,
		drivers: make(map[string]driverpb.DriverServiceClient),
		rdb:     rdb,
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

func (s *InstancesServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

var methodsToSync = []string{
	"manual_renew",
	"free_renew",
	"cancel_renew",
}

func (s *InstancesServer) Invoke(ctx context.Context, _req *connect.Request[pb.InvokeRequest]) (*connect.Response[pb.InvokeResponse], error) {
	log := s.log.Named("invoke")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
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
		syncer := sync.NewDataSyncer(log.With(zap.String("caller", "Invoke")), s.rdb, r.SP.GetUuid(), 5)
		defer syncer.Open()
		_ = syncer.WaitUntilOpenedAndCloseAfter()
		//log.Debug("Locking mutex")
		//m := s.spSyncers[r.SP.GetUuid()]
		//if m == nil {
		//	m = &go_sync.Mutex{}
		//	s.spSyncers[r.SP.GetUuid()] = m
		//}
		//m.Lock()
		//defer func() {
		//	log.Debug("Unlocking mutex")
		//	m.Unlock()
		//}()
	}

	var instance graph.Instance
	instance, err = graph.GetInstanceWithAccess(
		ctx, s.db,
		instance_id,
	)
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

func (s *InstancesServer) Delete(ctx context.Context, _req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	log := s.log.Named("delete")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	instance_id := driver.NewDocumentID(schema.INSTANCES_COL, req.Uuid)
	var instance graph.Instance
	instance, err := graph.GetWithAccess[graph.Instance](
		ctx, s.db,
		instance_id,
	)
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
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	igId := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, req.GetIg())
	var ig graph.InstancesGroup
	ig, err := graph.GetWithAccess[graph.InstancesGroup](
		ctx, s.db,
		igId,
	)
	if err != nil {
		log.Error("Failed to get instance group", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if ig.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	// TODO: set sp here to log service prodiver
	newId, err := s.ctrl.Create(ctx, igId, "", req.GetInstance())
	if err != nil {
		log.Error("Failed to create instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
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
	log.Debug("Requestor", zap.String("id", requestor))

	igId := driver.NewDocumentID(schema.INSTANCES_COL, req.GetInstance().GetUuid())
	var instance graph.Instance
	instance, err := graph.GetWithAccess[graph.Instance](
		ctx, s.db,
		igId,
	)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	// TODO: set sp here to log service prodiver
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
	log.Debug("Requestor", zap.String("id", requestor))

	instance_id := driver.NewDocumentID(schema.INSTANCES_COL, req.Uuid)
	var instance graph.Instance
	instance, err := graph.GetWithAccess[graph.Instance](
		ctx, s.db,
		instance_id,
	)
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
	log.Debug("Requestor", zap.String("id", requestor))

	instance_id := driver.NewDocumentID(schema.INSTANCES_COL, req.Uuid)
	var instance graph.Instance
	instance, err := graph.GetWithAccess[graph.Instance](
		ctx, s.db,
		instance_id,
	)
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
	log.Debug("Requestor", zap.String("id", requestor))

	igId := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, req.GetUuid())
	newSrvId := driver.NewDocumentID(schema.SERVICES_COL, req.GetService())
	ig, err := graph.GetWithAccess[graph.InstancesGroup](ctx, s.db, igId)

	if err != nil {
		log.Error("Failed to get instances group", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	srv, err := graph.GetWithAccess[graph.Service](ctx, s.db, newSrvId)

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

func (s *InstancesServer) TransferInstance(ctx context.Context, _req *connect.Request[pb.TransferInstanceRequest]) (*connect.Response[pb.TransferInstanceResponse], error) {
	log := s.log.Named("transfer")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	instanceId := driver.NewDocumentID(schema.INSTANCES_COL, req.Uuid)
	newIGId := driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, req.GetIg())
	inst, err := graph.GetWithAccess[graph.Instance](ctx, s.db, instanceId)

	if err != nil {
		log.Error("Failed to get instances group", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	ig, err := graph.GetWithAccess[graph.InstancesGroup](ctx, s.db, newIGId)

	if err != nil {
		log.Error("Failed to get service", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if inst.GetAccess().GetLevel() < accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	if ig.GetAccess().GetLevel() < accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", ig.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	igEdge, err := s.ctrl.GetEdge(ctx, instanceId.String(), schema.INSTANCES_GROUPS_COL)

	if err != nil {
		log.Error("Failed to get Service", zap.Error(err))
		return nil, err
	}

	err = s.ctrl.TransferInst(ctx, igEdge, newIGId, instanceId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.TransferInstanceResponse{
		Result: true,
		Meta:   nil,
	}), nil
}

func (s *InstancesServer) AddNote(ctx context.Context, _req *connect.Request[notes.AddNoteRequest]) (*connect.Response[notes.NoteResponse], error) {
	log := s.log.Named("AddNote")
	req := _req.Msg
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	instance_id := driver.NewDocumentID(schema.INSTANCES_COL, req.Uuid)
	var instance graph.Instance
	instance, err := graph.GetInstanceWithAccess(
		ctx, s.db,
		instance_id,
	)
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
	log.Debug("Requestor", zap.String("id", requestor))

	instance_id := driver.NewDocumentID(schema.INSTANCES_COL, req.Uuid)
	var instance graph.Instance
	instance, err := graph.GetInstanceWithAccess(
		ctx, s.db,
		instance_id,
	)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() != accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	ok := graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), accesspb.Level_ROOT)

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
	log.Debug("Requestor", zap.String("id", requestor))

	instance_id := driver.NewDocumentID(schema.INSTANCES_COL, req.Uuid)
	var instance graph.Instance
	instance, err := graph.GetInstanceWithAccess(
		ctx, s.db,
		instance_id,
	)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if instance.GetAccess().GetLevel() != accesspb.Level_ROOT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	ok := graph.HasAccess(ctx, s.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), accesspb.Level_ROOT)

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
			query += fmt.Sprintf(` FILTER LOWER(node.title) LIKE LOWER("%s")
|| acc.data.email LIKE "%s"
|| acc._key LIKE "%s"
|| node._key LIKE "%s"`,
				"%"+val.GetStringValue()+"%", "%"+val.GetStringValue()+"%", "%"+val.GetStringValue()+"%", "%"+val.GetStringValue()+"%")
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
			query += fmt.Sprintf(` FILTER sp in @%s`, key)
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
			query += fmt.Sprintf(` FILTER CONTAINS(TO_STRING(node.state.meta.networking.public), "%s") || CONTAINS(TO_STRING(node.state.meta.networking.private), "%s")`, val, val)
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
	} else if field == "type" {
		query += fmt.Sprintf(" SORT bp.type %s", order)
	} else if field == "billing_plan" {
		query += fmt.Sprintf(" SORT bp.title %s", order)
	} else if field == "sp" {
		query += fmt.Sprintf(" SORT sp.title %s", order)

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

	//// Calculate estimate and period values if not presented
	//for _, value := range result.Pool {
	//	inst := value.Instance
	//	if inst.GetEstimate() == 0 {
	//		inst.Estimate, _ = s.ctrl.CalculateInstanceEstimatePeriodicPrice(inst)
	//	}
	//	if inst.GetPeriod() == 0 {
	//		inst.Period, _ = s.ctrl.GetInstancePeriod(inst)
	//	}
	//}

	log.Debug("Result", zap.Any("result", &result))
	return connect.NewResponse(&result), nil
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

	return connect.NewResponse(&result), nil
}
