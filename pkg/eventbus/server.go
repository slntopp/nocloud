package eventbus

import (
	"context"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
	"go.uber.org/zap"
)

type EventBusServer struct {
	pb.UnimplementedEventsServiceServer
	log *zap.Logger
	bus *EventBus
}

func NewServer(logger *zap.Logger, conn *amqp091.Connection) *EventBusServer {

	log := logger.Named("EventBusServer")

	log.Info("creating new EvenBusServer instance")

	bus, err := NewEventBus(conn, log)
	if err != nil {
		log.Fatal("cannot create EventBus", zap.Error(err))
	}

	return &EventBusServer{
		log: log,
		bus: bus,
	}
}

func (s *EventBusServer) Publish(ctx context.Context, event *pb.Event) (*pb.Response, error) {

	s.log.Info("got publish request", zap.Any("event", event))

	event.Id = uuid.New().String()

	if err := s.bus.Pub(ctx, event); err != nil {
		return nil, err
	}

	return &pb.Response{}, nil
}

func (s *EventBusServer) Consume(req *pb.ConsumeRequest, srv pb.EventsService_ConsumeServer) error {

	defer s.bus.Unsub(req)

	s.log.Info("got consume request", zap.Any("request", req))

	ch, err := s.bus.Sub(srv.Context(), req)
	if err != nil {
		return err
	}

	done := srv.Context().Done()

	go func() {
		for msg := range ch {
			srv.Send(msg)
		}
	}()

	<-done

	return nil
}

func (s *EventBusServer) List(ctx context.Context, req *pb.ConsumeRequest) (*pb.Events, error) {

	s.log.Info("got list request", zap.Any("request", req))

	events, err := s.bus.List(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.Events{Events: events}, nil
}

func (s *EventBusServer) Cancel(ctx context.Context, req *pb.CancelRequest) (*pb.Response, error) {

	s.log.Info("got cancel request", zap.Any("request", req))

	return &pb.Response{}, s.bus.Cancel(ctx, req)
}
