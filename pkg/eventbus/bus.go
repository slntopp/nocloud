package eventbus

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
	"go.uber.org/zap"
)

// EventBus handles interservice communication throuh RabbitMQ
type EventBus struct {
	conn     *Connection
	exchange *Exchange
	log      *zap.Logger
}

func NewEventBus(conn *amqp.Connection, logger *zap.Logger) (*EventBus, error) {

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

	_, err := bus.exchange.DeriveQueue(event.Key)
	if err != nil {
		return err
	}

	return bus.exchange.Send(ctx, event)
}

func (bus *EventBus) Sub(key string) (<-chan *pb.Event, error) {

	bus.log.Info("consuming events", zap.String("key", key))

	// Disconnect other consumers
	if err := bus.conn.Channel().Cancel(key, NO_WAIT); err != nil {
		return nil, err
	}

	q, err := bus.exchange.DeriveQueue(key)
	if err != nil {
		return nil, err
	}

	ch, err := q.Consume()
	if err != nil {
		return nil, err
	}

	return ch, nil
}
