package events_logging

import (
	pb "github.com/slntopp/nocloud-proto/events_logging"
	"go.uber.org/zap"
)

type EventsLoggingServer struct {
	pb.UnimplementedEventsLoggingServiceServer
	rep *SqliteRepository

	log *zap.Logger
}

func NewEventsLoggingServer(_log *zap.Logger, rep *SqliteRepository) *EventsLoggingServer {
	log := _log.Named("EventsLoggingServer")
	log.Debug("New EventsLogging Server Creating")

	return &EventsLoggingServer{log: log, rep: rep}
}
