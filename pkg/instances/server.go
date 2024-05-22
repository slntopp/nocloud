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
	"context"
	"fmt"
	"slices"
	"time"

	elpb "github.com/slntopp/nocloud-proto/events_logging"
	"github.com/slntopp/nocloud-proto/notes"

	"github.com/arangodb/go-driver"
	amqp "github.com/rabbitmq/amqp091-go"
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
)

type InstancesServer struct {
	pb.UnimplementedInstancesServiceServer
	log *zap.Logger

	ctrl    *graph.InstancesController
	ig_ctrl *graph.InstancesGroupsController

	drivers map[string]driverpb.DriverServiceClient

	db driver.Database
}

func NewInstancesServiceServer(logger *zap.Logger, db driver.Database, rbmq *amqp.Connection) *InstancesServer {
	log := logger.Named("instances")
	log.Debug("New Instances Server Creating")
	ig_ctrl := graph.NewInstancesGroupsController(logger, db)

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
		drivers: make(map[string]driverpb.DriverServiceClient),
	}
}

func (s *InstancesServer) RegisterDriver(type_key string, client driverpb.DriverServiceClient) {
	s.drivers[type_key] = client
}

func (s *InstancesServer) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
	log := s.log.Named("invoke")
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

	if instance.GetAccess().GetLevel() < accesspb.Level_MGMT {
		log.Error("Access denied", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	if instance.GetState().GetState() == stpb.NoCloudState_SUSPENDED && instance.GetAccess().GetLevel() < accesspb.Level_ROOT {
		log.Error("Machine is suspended. Functionality is limited", zap.String("uuid", instance.GetUuid()))
		return nil, status.Error(codes.Unavailable, "Machine is suspended. Functionality is limited")
	}

	r, err := s.ctrl.GetGroup(ctx, instance_id.String())
	if err != nil {
		log.Error("Failed to get Group and ServicesProvider", zap.Error(err))
		return nil, err
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
		return invoke, err
	}

	nocloud.Log(log, event)

	return invoke, nil
}

func (s *InstancesServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	log := s.log.Named("delete")
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

	return &pb.DeleteResponse{
		Result: true,
	}, nil
}

func (s *InstancesServer) Detach(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	log := s.log.Named("Detach")
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

	return &pb.DeleteResponse{
		Result: true,
	}, nil
}

func (s *InstancesServer) Attach(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	log := s.log.Named("Attach")
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

	return &pb.DeleteResponse{
		Result: true,
	}, nil
}

func (s *InstancesServer) TransferIG(ctx context.Context, req *pb.TransferIGRequest) (*pb.TransferIGResponse, error) {
	log := s.log.Named("transfer")
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

	return &pb.TransferIGResponse{
		Result: true,
		Meta:   nil,
	}, nil
}

func (s *InstancesServer) TransferInstance(ctx context.Context, req *pb.TransferInstanceRequest) (*pb.TransferInstanceResponse, error) {
	log := s.log.Named("transfer")
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

	return &pb.TransferInstanceResponse{
		Result: true,
		Meta:   nil,
	}, nil
}

func (s *InstancesServer) AddNote(ctx context.Context, req *notes.AddNoteRequest) (*notes.NoteResponse, error) {
	log := s.log.Named("AddNote")
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

	return &notes.NoteResponse{
		Result:     true,
		AdminNotes: instance.GetAdminNotes(),
	}, nil
}

func (s *InstancesServer) PatchNote(ctx context.Context, req *notes.PatchNoteRequest) (*notes.NoteResponse, error) {
	log := s.log.Named("Patch")
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

		return &notes.NoteResponse{
			Result:     true,
			AdminNotes: instance.GetAdminNotes(),
		}, nil
	}

	return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Instance notes")
}

func (s *InstancesServer) RemoveNote(ctx context.Context, req *notes.RemoveNoteRequest) (*notes.NoteResponse, error) {
	log := s.log.Named("Remove")
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
		instance.AdminNotes = slices.Delete(instance.GetAdminNotes(), int(req.GetIndex()), int(req.GetIndex()+1))

		err = s.ctrl.UpdateNotes(ctx, instance.Instance)
		if err != nil {
			return nil, err
		}

		return &notes.NoteResponse{
			Result:     true,
			AdminNotes: instance.GetAdminNotes(),
		}, nil
	}

	return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Instance notes")
}

const listInstancesQuery = `
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
	            RETURN sp_node._key
        )
        LET srv = path.vertices[-3]._key
        LET ns = path.vertices[-4]._key
        LET acc = path.vertices[-5]._key
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

return { 
	pool: (@limit > 0) ? SLICE(instances, @offset, @limit) : instances,
	count: LENGTH(instances)
}
`

func (s *InstancesServer) List(ctx context.Context, req *pb.ListInstancesRequest) (*pb.ListInstancesResponse, error) {
	log := s.log.Named("List")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	limit, page := req.GetLimit(), req.GetPage()
	offset := (page - 1) * limit

	var query string
	bindVars := map[string]interface{}{
		"account":           driver.NewDocumentID(schema.ACCOUNTS_COL, requestor),
		"permissions_graph": schema.PERMISSIONS_GRAPH.Name,
		"instances":         schema.INSTANCES_COL,
		"bps":               schema.BILLING_PLANS_COL,
		"service_provider":  schema.SERVICES_PROVIDERS_COL,
		"offset":            offset,
		"limit":             limit,
	}

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT node.%s %s`
		field, sort := req.GetField(), req.GetSort()

		query += fmt.Sprintf(subQuery, field, sort)
	}

	for key, val := range req.GetFilters() {
		if key == "account" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER acc in @%s`, key)
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
		}
	}

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

	// Calculate estimate and period values if not presented
	for _, value := range result.Pool {
		inst := value.Instance
		if inst.GetEstimate() == 0 {
			inst.Estimate, _ = s.ctrl.CalculateInstanceEstimatePeriodicPrice(inst)
		}
		if inst.GetPeriod() == 0 {
			inst.Period, _ = s.ctrl.GetInstancePeriod(inst)
		}
	}

	return &result, nil
}
