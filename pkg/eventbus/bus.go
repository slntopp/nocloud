package eventbus

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
)

type EventBus struct {
	ch       *Channel
	exchange *Exchange
}

func NewEventBus(conn *amqp.Connection) (*EventBus, error) {

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	ch := NewChannel(channel)

	exchange, err := NewExchange(ch, EXCHANGE_NAME, AlternateExchange)
	if err != nil {
		return nil, err
	}

	bus := &EventBus{
		ch:       ch,
		exchange: exchange,
	}

	go bus.sync()

	return bus, nil
}

func (bus *EventBus) Pub(ctx context.Context, event *pb.Event) (err error) {
	return bus.exchange.Send(ctx, event)
}

func (bus *EventBus) Sub(key string) (<-chan *pb.Event, error) {
	q, err := NewQueue(bus.ch, "")
	if err != nil {
		return nil, err
	}

	bus.exchange.Bind(q, key)

	ch, err := q.Consume()
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (bus *EventBus) sync() {

	q, err := NewQueue(bus.ch, EXCHANGE_BUFFER)
	if err != nil {
		return
	}

	ch, err := q.Consume()
	if err != nil {
		return
	}

	for event := range ch {
		_, err = NewQueue(bus.ch, event.Key)
		if err != nil {
			continue
		}

		bus.exchange.Send(context.Background(), event)
	}
}
