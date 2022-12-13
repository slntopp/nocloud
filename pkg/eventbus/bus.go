package eventbus

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
)

const (
	// Consume properties
	CONSUME_AUTO_ACK = false
	// Common properties
	NO_WAIT = false
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

	exchange, err := NewExchange(ch, EXCHANGE_NAME, Alternate)
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

func (bus *EventBus) Pub(event *pb.Event) (err error) {
	return
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
