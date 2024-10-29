package eventbus

import (
	"context"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"

	pb "github.com/slntopp/nocloud-proto/events"
	"go.uber.org/zap"
)

// EventBus handles interservice communication throuh RabbitMQ
type EventBus struct {
	conn     *Connection
	exchange *Exchange
	log      *zap.Logger
}

func NewEventBus(conn rabbitmq.Connection, logger *zap.Logger) (*EventBus, error) {

	log := logger.Named("EventBus")

	log.Info("creating new EventBus instance")

	connection, err := NewConnection(conn)
	if err != nil {
		return nil, err
	}

	exchange, err := NewExchange(connection, EXCHANGE_NAME, DefaultExchange)
	if err != nil {
		return nil, err
	}

	bus := &EventBus{
		conn:     connection,
		exchange: exchange,
		log:      log,
	}

	return bus, nil
}

func (bus *EventBus) Pub(ctx context.Context, event *pb.Event) error {

	bus.log.Info("publishing event", zap.Any("event", event))

	_, err := bus.exchange.DeriveQueue(Topic(event))
	if err != nil {
		return err
	}

	return bus.exchange.Send(ctx, event)
}

func (bus *EventBus) Sub(ctx context.Context, req *pb.ConsumeRequest) (<-chan *pb.Event, error) {

	topic := Topic(req)

	bus.log.Info("consuming events", zap.String("key", topic))

	// Disconnect other consumers
	if err := bus.Unsub(req); err != nil {
		return nil, err
	}

	q, err := bus.exchange.DeriveQueue(topic)
	if err != nil {
		return nil, err
	}

	ch, err := q.Consume()
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (bus *EventBus) Unsub(req *pb.ConsumeRequest) error {
	return bus.conn.Channel().Cancel(Topic(req), NO_WAIT)
}

func (bus *EventBus) List(ctx context.Context, req *pb.ConsumeRequest) ([]*pb.Event, error) {

	q, err := bus.exchange.DeriveQueue(Topic(req))
	if err != nil {
		return nil, err
	}

	return q.List(ctx)
}

func (bus *EventBus) Cancel(ctx context.Context, req *pb.CancelRequest) error {

	q, err := bus.exchange.DeriveQueue(Topic(req))
	if err != nil {
		return err
	}

	return q.Cancel(ctx, req.Id)
}
