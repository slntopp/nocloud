package eventbus

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
	"go.uber.org/zap"
)

type EventBusServer struct {
	pb.UnimplementedEventsServiceServer
	log *zap.Logger
	bus *EventBus
}

func NewServer(log *zap.Logger, conn *amqp091.Connection) *EventBusServer {
	return &EventBusServer{
		log: log.Named("EventBusServer"),
		bus: New(conn),
	}
}

func (s *EventBusServer) Publish(ctx context.Context, event *pb.Event) (*pb.Response, error) {

	s.log.Info("got publish request")

	if err := s.bus.Pub(event.Body, event.Topic); err != nil {
		return nil, err
	}

	return &pb.Response{}, nil
}

func (s *EventBusServer) Consume(req *pb.ConsumeRequest, srv pb.EventsService_ConsumeServer) error {

	s.log.Info("got consume request")

	ch, err := s.bus.Sub(req.Topic)
	if err != nil {
		return err
	}

	for msg := range ch {
		srv.Send(&pb.Event{
			Body: msg,
		})
	}

	return nil
}
