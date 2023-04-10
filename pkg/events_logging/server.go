package events_logging

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/events_logging"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventsLoggingServer struct {
	pb.UnimplementedEventsLoggingServiceServer
	rep *SqliteRepository

	db driver.Database

	log *zap.Logger
}

func NewEventsLoggingServer(_log *zap.Logger, rep *SqliteRepository, db driver.Database) *EventsLoggingServer {
	log := _log.Named("EventsLoggingServer")
	log.Debug("New EventsLogging Server Creating")

	return &EventsLoggingServer{log: log, rep: rep, db: db}
}

func Log(log *zap.Logger, event *pb.Event) {
	log.Log(nocloud.NOCLOUD_LOG_LEVEL, "",
		zap.String("entity", event.Entity),
		zap.String("uuid", event.Uuid),
		zap.String("scope", event.Scope),
		zap.String("action", event.Action),
		zap.Int32("rc", event.Rc),
		zap.String("requestor", event.Requestor),
		zap.Int64("ts", event.Ts),
		zap.String("diff", event.Snapshot.Diff),
	)
}

func (s *EventsLoggingServer) GetEvents(ctx context.Context, req *pb.GetEventsRequest) (*pb.Events, error) {
	log := s.log.Named("GetEvents")

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	node := driver.NewDocumentID(req.GetEntity(), req.GetUuid())
	hasAccess := graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN)

	if !hasAccess {
		return nil, status.Error(codes.PermissionDenied, "Not enoguh Access Rights")
	}

	events, err := s.rep.GetEvents(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.Events{Events: events}, nil
}

func (s *EventsLoggingServer) GetTrace(ctx context.Context, req *pb.GetTraceRequest) (*pb.Events, error) {
	log := s.log.Named("GetTrace")

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	node := driver.NewDocumentID(schema.ACCOUNTS_COL, req.GetRequestor())
	hasAccess := graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN)

	if !hasAccess {
		return nil, status.Error(codes.PermissionDenied, "Not enoguh Access Rights")
	}

	events, err := s.rep.GetTrace(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.Events{Events: events}, nil
}
