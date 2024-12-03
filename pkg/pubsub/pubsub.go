package ps

import (
	"context"
	"errors"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"time"
)

const DEFAULT_EXCHANGE = "nocloud"

var errNoNack = errors.New("no nack")

func IsNoNackErr(err error) bool {
	return errors.Is(err, errNoNack)
}
func NoNackErr(err error) error {
	return fmt.Errorf("%w: %w", errNoNack, err)
}

type PubSub[T proto.Message] struct {
	log  *zap.Logger
	conn rabbitmq.Connection
	ch   rabbitmq.Channel
}

type ConsumeOptions struct {
	Durable   bool
	NoWait    bool
	Exclusive bool
}

func NewPubSub[T proto.Message](conn rabbitmq.Connection, log *zap.Logger) *PubSub[T] {
	log = log.Named("PubSub")
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel", zap.Error(err))
	}
	return &PubSub[T]{
		log:  log,
		conn: conn,
		ch:   ch,
	}
}

func (ps *PubSub[T]) Channel() rabbitmq.Channel {
	if !ps.ch.IsClosed() {
		return ps.ch
	}
	ch, err := ps.conn.Channel()
	if err != nil {
		ps.log.Fatal("Failed to reopen a channel", zap.Error(err))
	}
	ps.ch = ch
	return ps.ch
}

func (ps *PubSub[T]) Consume(name, exchange, topic string, options ...ConsumeOptions) (<-chan amqp091.Delivery, error) {
	var (
		exclusive bool
		durable   = true
		noWait    bool
	)
	if len(options) > 0 {
		o := options[0]
		exclusive = o.Exclusive
		durable = o.Durable
		noWait = o.NoWait
	}
	topic = exchange + "." + topic

	log := ps.log.Named("Consume." + name)
	ch, err := ps.conn.Channel()
	if err != nil {
		log.Error("Failed to open channel", zap.Error(err))
		return nil, err
	}

	if exchange != "" {
		if err := ch.ExchangeDeclare(exchange, "topic", durable, false, false, noWait, nil); err != nil {
			log.Error("Failed to declare a exchange", zap.Error(err))
			return nil, err
		}
	}

	q, err := ch.QueueDeclare(name, durable, false, exclusive, noWait, nil)
	if err != nil {
		log.Error("Failed to declare a queue", zap.Error(err))
		return nil, err
	}

	if exchange != "" {
		err = ch.QueueBind(q.Name, topic, exchange, noWait, nil)
		if err != nil {
			log.Error("Failed to bind a queue", zap.Error(err))
			return nil, err
		}
	}

	msgs, err := ch.Consume(q.Name, name, false, exclusive, false, noWait, nil)
	if err != nil {
		log.Error("Failed to register a consumer", zap.Error(err))
		return nil, err
	}

	return msgs, nil
}

func (ps *PubSub[T]) Publish(exchange, topic string, msg T) error {
	topic = exchange + "." + topic
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	body, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return ps.Channel().PublishWithContext(ctx, exchange, topic, false, false, amqp091.Publishing{
		ContentType:  "text/plain",
		DeliveryMode: amqp091.Persistent,
		Body:         body,
	})
}

func (ps *PubSub[T]) Publisher(exchange, topic string) func(obj T) error {
	return func(obj T) error {
		return ps.Publish(exchange, topic, obj)
	}
}

func (ps *PubSub[T]) ConsumerInit(name, exchange, topic string, processor func(obj T) error) error {
	msgs, err := ps.Consume(name, exchange, topic)
	if err != nil {
		return err
	}
	go func() {
		log := ps.log.Named("Consumer." + name)
		for msg := range msgs {
			var req T
			err := proto.Unmarshal(msg.Body, req)
			if err != nil {
				if err = msg.Ack(false); err != nil {
					log.Error("Failed to ack the delivery", zap.Error(err))
				}
				continue
			}
			if err = processor(req); err != nil {
				if err = msg.Nack(false, false); err != nil {
					log.Error("Failed to nack the delivery", zap.Error(err))
				}
			}
		}
	}()
	return nil
}
