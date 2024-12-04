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
const dlx = "nocloud.dlx"
const dlxQueue = "dlx-queue"

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
	Durable    bool
	NoWait     bool
	Exclusive  bool
	WithRetry  bool
	MaxRetries int
}

func HandleAckNack(log *zap.Logger, del amqp091.Delivery, err error) {
	if !IsNoNackErr(err) {
		log.Error("Failed to process delivery event (nack)", zap.Error(err))
		if err = del.Nack(false, false); err != nil {
			log.Error("Failed to nack the delivery", zap.Error(err))
		}
	} else {
		log.Warn("Failed to process delivery event (ack)", zap.Error(err))
		if err = del.Ack(false); err != nil {
			log.Error("Failed to acknowledge the delivery", zap.Error(err))
		}
	}
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
		exclusive  bool
		durable    = true
		noWait     bool
		withRetry  = true
		maxRetries = 3
	)
	if len(options) > 0 {
		o := options[0]
		exclusive = o.Exclusive
		durable = o.Durable
		noWait = o.NoWait
		withRetry = o.WithRetry
		maxRetries = o.MaxRetries
	}
	topic = exchange + "." + topic

	log := ps.log.Named("Consume." + name)
	ch, err := ps.conn.Channel()
	if err != nil {
		log.Error("Failed to open channel", zap.Error(err))
		return nil, err
	}

	var dlxQ *amqp091.Queue
	if withRetry {
		err = ch.ExchangeDeclare(dlx, "topic", durable, false, false, false, nil)
		if err != nil {
			log.Error("Failed to declare a dlx", zap.Error(err))
			return nil, err
		}
		_dlxQ, err := ch.QueueDeclare(name+"."+dlxQueue, durable, false, false, false, map[string]interface{}{
			"x-dead-letter-exchange": exchange,
			"x-overflow":             "reject-publish",
			"x-message-ttl":          5000,
		})
		if err != nil {
			log.Error("Failed to declare a dlx queue", zap.Error(err))
			return nil, err
		}
		err = ch.QueueBind(_dlxQ.Name, topic, dlx, false, nil)
		if err != nil {
			log.Error("Failed to bind dlx queue", zap.Error(err))
			return nil, err
		}
		dlxQ = &_dlxQ
	}

	if exchange != "" {
		if err := ch.ExchangeDeclare(exchange, "topic", durable, false, false, noWait, nil); err != nil {
			log.Error("Failed to declare a exchange", zap.Error(err))
			return nil, err
		}
	}

	var params amqp091.Table
	if withRetry {
		params = amqp091.Table{
			"x-dead-letter-exchange": dlx,
		}
	}
	q, err := ch.QueueDeclare(name, durable, false, exclusive, noWait, params)
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

	if dlxQ != nil {
		go ps.consumeDlx(log, ch, dlxQ.Name, maxRetries)
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

func (ps *PubSub[T]) consumeDlx(log *zap.Logger, ch rabbitmq.Channel, dlxQueue string, maxRetries int) {
	log = log.Named("DLX." + dlxQueue)

	msgs, err := ch.Consume(dlxQueue, dlxQueue, false, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to register a dlx consumer", zap.Error(err))
	}

	for msg := range msgs {
		log.Info("Received a message from dlx", zap.Any("routine_key", msg.RoutingKey))
		var req = new(T)
		err := proto.Unmarshal(msg.Body, *req)
		if err != nil {
			log.Error("Failed to unmarshal request", zap.Error(err))
			if err = msg.Ack(false); err != nil {
				log.Error("Failed to ack the delivery", zap.Error(err))
			}
			continue
		}
		log.Info("Unmarshalled message", zap.Any("message", req))
		if msg.Headers["x-death"] != nil {
			deaths := msg.Headers["x-death"].([]interface{})
			log.Info("Dead lettered message info", zap.Any("deaths", deaths))
			total := int64(0)
			for _, death := range deaths {
				deathMap := death.(amqp091.Table)
				count := deathMap["count"].(int64)
				total += count
			}
			if total > int64(maxRetries) {
				log.Info("Max retries reached", zap.Int64("total", total))
				if err = msg.Ack(false); err != nil {
					log.Error("Failed to ack the delivery", zap.Error(err))
				}
				continue
			}
			log.Info("Retrying again", zap.Int64("current", total), zap.Int("max", maxRetries))
			if err = msg.Nack(false, false); err != nil {
				log.Error("Failed to nack the delivery", zap.Error(err))
			}
			continue
		}
		log.Info("x-death not found", zap.Any("routine_key", msg.RoutingKey))
		if err = msg.Ack(false); err != nil {
			log.Error("Failed to ack the delivery", zap.Error(err))
		}
	}
}
