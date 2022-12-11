package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

const EXCHANGE_NAME = "nocloud-event-bus"

type Bus struct {
	rbmq *amqp.Connection
	ch   *amqp.Channel
}

func NewBus(conn *amqp.Connection) *Bus {

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("cannot create rbmq channel")
	}

	if err := channel.ExchangeDeclare(EXCHANGE_NAME, "topic", true, false, false, false, nil); err != nil {
		log.Fatalf("cannot create rbmq exchange")
	}

	return &Bus{
		rbmq: conn,
		ch:   channel,
	}
}

func (bus *Bus) Publish(msg []byte, topic string) error {
	return bus.ch.PublishWithContext(context.Background(), EXCHANGE_NAME, topic, false, false, amqp.Publishing{
		Body: msg,
	})
}

func (bus *Bus) Sub(topic string) (<-chan amqp.Delivery, error) {
	queueName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), topic)

	_, err := bus.ch.QueueDeclare(
		queueName,
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	deliveries, err := bus.ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "consume")
	}

	return deliveries, nil
}

type EventBus[T any] struct {
	bus *Bus
}

func NewTypedBus[T any](bus *Bus) *EventBus[T] {
	return &EventBus[T]{
		bus: bus,
	}
}

func (bus *EventBus[T]) Publish(msg T, topic string) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return bus.bus.Publish(bytes, topic)
}

func (bus *EventBus[T]) Sub(topic string) (<-chan T, error) {
	ch := make(chan T)

	delivery, err := bus.bus.Sub(topic)
	if err != nil {
		return nil, err
	}

	go func() {
		for msg := range delivery {
			var val T
			err := json.Unmarshal(msg.Body, &val)
			if err != nil {
				continue
			}
			ch <- val
		}
	}()
	return ch, nil
}
