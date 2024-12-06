package ps

import (
	"context"
	"errors"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	epb "github.com/slntopp/nocloud-proto/events"
	elpb "github.com/slntopp/nocloud-proto/events_logging"
	ipb "github.com/slntopp/nocloud-proto/instances"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"strings"
	"time"
)

const DEFAULT_EXCHANGE = "nocloud"

const dlx = "nocloud.dlx"
const dlxTTL = "nocloud.dlx.ttl"
const dlxQueue = "dlx-queue-proc"
const dlxQueueTTL = "dlx-queue-ttl"

var errNoNack = errors.New("no nack")

func IsNoNackErr(err error) bool {
	return errors.Is(err, errNoNack)
}
func NoNackErr(err error) error {
	return fmt.Errorf("%w: %w", errNoNack, err)
}

func queueDeclare(conn rabbitmq.Connection, name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error) {
	channel := func() (rabbitmq.Channel, func()) {
		ch, _ := conn.Channel()
		return ch, func() { _ = ch.Close() }
	}
	retried := false
retry:
	ch, term := channel()
	defer term()
	q, err := ch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
	if err != nil {
		if strings.Contains(err.Error(), "PRECONDITION_FAILED") && !retried {
			ch, term = channel()
			defer term()
			if _, err = ch.QueueDelete(name, false, false, false); err != nil {
				return amqp091.Queue{}, err
			}
			retried = true
			goto retry
		}
	}
	return q, err
}

func unwrapProtoMessage(b []byte) (proto.Message, error) {
	types := []proto.Message{&epb.Event{}, &ipb.Context{}}
	for _, t := range types {
		if err := proto.Unmarshal(b, t); err == nil {
			return t, nil
		}
	}
	return nil, fmt.Errorf("not found")
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
	DelayMilli int
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
		delayMilli = 5000
	)
	if len(options) > 0 {
		o := options[0]
		exclusive = o.Exclusive
		durable = o.Durable
		noWait = false
		withRetry = o.WithRetry
		maxRetries = o.MaxRetries
		delayMilli = o.DelayMilli
	}
	topic = exchange + "." + topic

	log := ps.log.Named("Consume." + name)
	ch, err := ps.conn.Channel()
	if err != nil {
		log.Error("Failed to open channel", zap.Error(err))
		return nil, err
	}

	var queueDlx amqp091.Queue
	if withRetry {
		err = ch.ExchangeDeclare(dlx, "direct", true, false, false, false, nil)
		if err != nil {
			log.Error("Failed to declare a dlx", zap.Error(err))
			return nil, err
		}
		err = ch.ExchangeDeclare(dlxTTL, "direct", true, false, false, false, nil)
		if err != nil {
			log.Error("Failed to declare a dlx ttl", zap.Error(err))
			return nil, err
		}
		dlxQ, err := queueDeclare(ps.conn, name+"."+dlxQueue, durable, false, false, false, map[string]interface{}{
			"x-dead-letter-exchange": dlxTTL,
		})
		if err != nil {
			log.Error("Failed to declare a dlx queue", zap.Error(err))
			return nil, err
		}
		queueDlx = dlxQ
		dlxQTtl, err := queueDeclare(ps.conn, name+"."+dlxQueueTTL, durable, false, false, false, map[string]interface{}{
			"x-dead-letter-exchange": exchange,
			"x-message-ttl":          delayMilli,
		})
		if err != nil {
			log.Error("Failed to declare a dlx ttl queue", zap.Error(err))
			return nil, err
		}
		if err = ch.QueueBind(dlxQ.Name, name, dlx, false, nil); err != nil {
			log.Error("Failed to bind dlx queue", zap.Error(err))
			return nil, err
		}
		if err = ch.QueueBind(dlxQTtl.Name, name, dlxTTL, false, nil); err != nil {
			log.Error("Failed to bind dlx ttl queue", zap.Error(err))
			return nil, err
		}
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
			"x-dead-letter-exchange":    dlx,
			"x-dead-letter-routing-key": name,
		}
	}
	q, err := queueDeclare(ps.conn, name, durable, false, exclusive, noWait, params)
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
		if withRetry {
			err = ch.QueueBind(q.Name, q.Name, exchange, false, nil)
			if err != nil {
				log.Error("Failed to bind a queue", zap.Error(err))
				return nil, err
			}
		}
	}

	msgs, err := ch.Consume(q.Name, name, false, exclusive, false, noWait, nil)
	if err != nil {
		log.Error("Failed to register a consumer", zap.Error(err))
		return nil, err
	}

	if withRetry {
		go ps.consumeDlx(log, ch, queueDlx.Name, q.Name, maxRetries)
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

func (ps *PubSub[T]) consumeDlx(log *zap.Logger, ch rabbitmq.Channel, dlxQueue string, originalQueue string, maxRetries int) {
	log = log.Named("DLX." + dlxQueue)
	msgs, err := ch.Consume(dlxQueue, dlxQueue, false, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to register a dlx consumer", zap.Error(err))
	}

	for msg := range msgs {
		log := log.With(zap.String("routing_key", msg.RoutingKey))
		if msg.Headers["x-death"] != nil {
			deaths := msg.Headers["x-death"].([]interface{})
			total := int64(0)
			if len(deaths) == 0 {
				goto nack
			}
			for _, death := range deaths {
				deathMap := death.(amqp091.Table)
				if deathMap["queue"].(string) != originalQueue {
					continue
				}
				count := deathMap["count"].(int64)
				total += count
			}
			total--
			if total >= int64(maxRetries) {
				log.Debug("Max retries reached", zap.Int64("retries_done", total), zap.Int("max", maxRetries))
				t, _ := unwrapProtoMessage(msg.Body)
				js, err := protojson.Marshal(t)
				if err != nil {
					log.Error("Failed to marshal message. Probably couldn't find proto.Message type", zap.Error(err))
				}
				diffTmpl := `{"value":"%s","op":"ERROR","path":"LOG"}`
				nocloud.Log(log, &elpb.Event{
					Entity:    "system",
					Uuid:      "no_uuid",
					Scope:     "errors",
					Action:    "system_action_deadlined",
					Rc:        int32(total),
					Requestor: "system",
					Ts:        time.Now().Unix(),
					Priority:  1,
					Snapshot: &elpb.Snapshot{
						Diff: fmt.Sprintf(diffTmpl,
							fmt.Sprintf("System action were declined after max retries (%d). RoutineKey: %s; RawDeliveredMessage: %s",
								total, msg.RoutingKey, string(js))),
					},
				})
				if err = msg.Ack(false); err != nil {
					log.Error("Failed to ack the delivery", zap.Error(err))
				}
				continue
			}
		nack:
			log.Debug("Retrying...", zap.Int64("current_retry", total+1), zap.Int("max", maxRetries))
			if err = msg.Nack(false, false); err != nil {
				log.Error("Failed to nack the delivery", zap.Error(err))
			}
			continue
		}
		log.Error("Header x-death not found. This should not happen in DLX queue")
		if err = msg.Ack(false); err != nil {
			log.Error("Failed to ack the delivery", zap.Error(err))
		}
	}
}
