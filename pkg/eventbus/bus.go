package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EXCHANGE_NAME        = "nocloud-event-bus"
	EXCHANGE_DURABLE     = true // essential for retention
	EXCHANGE_AUTO_DELETE = false
	EXCHANGE_INTERNAL    = false
	EXCHANGE_NO_WAIT     = false
)

type EventBus struct {
	ch *amqp.Channel
}

func New(conn *amqp.Connection) *EventBus {

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("cannot create rbmq channel")
	}

	if err := channel.ExchangeDeclare(
		EXCHANGE_NAME,
		"topic",
		EXCHANGE_DURABLE,
		EXCHANGE_AUTO_DELETE,
		EXCHANGE_INTERNAL,
		EXCHANGE_NO_WAIT,
		nil,
	); err != nil {
		log.Fatalf("cannot create rbmq exchange")
	}

	return &EventBus{
		ch: channel,
	}
}

func (bus *EventBus) Pub(msg any, topic string) (err error) {
	var payload []byte

	switch msg.(type) {
	case []byte:
		payload = (msg).([]byte)
	default:
		payload, err = json.Marshal(msg)
		if err != nil {
			return err
		}
	}

	return bus.ch.PublishWithContext(context.Background(), EXCHANGE_NAME, topic, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Body:         payload,
	})
}

func (bus *EventBus) Sub(topic string) (<-chan []byte, error) {

	queueName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), topic)

	if _, err := bus.ch.QueueDeclare(queueName, true, true, true, false, nil); err != nil {
		return nil, err
	}

	if err := bus.ch.QueueBind(
		queueName,
		topic,
		EXCHANGE_NAME,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	deliveries, err := bus.ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		return nil, errors.Wrap(err, "consume")
	}

	ch := make(chan []byte)

	go func() {
		for delivery := range deliveries {
			ch <- delivery.Body
		}
	}()

	return ch, nil
}

func Cast[T any](ch <-chan []byte) <-chan T {
	buffer := make(chan T)
	go func() {
		for bytes := range ch {
			var msg T
			if err := json.Unmarshal(bytes, &msg); err == nil {
				buffer <- msg
			}
		}
	}()

	return buffer
}
