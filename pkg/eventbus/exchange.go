package eventbus

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
)

type Exchange struct {
	ch   *Channel
	Name string
}

type ExchangeType int64

const (
	DefaultExchange ExchangeType = iota
	AlternateExchange
)

func NewExchange(ch *Channel, name string, t ExchangeType) (*Exchange, error) {

	args := amqp091.Table{}

	if t == AlternateExchange {
		args["alternate-exchange"] = EXCHANGE_BUFFER
	}

	if err := ch.ExchangeDeclare(name, EXCHANGE_KIND, EXCHANGE_DURABLE, EXCHANGE_AUTO_DELETE, EXCHANGE_INTERNAL, NO_WAIT, args); err != nil {
		return nil, err
	}

	return &Exchange{
		ch:   ch,
		Name: name,
	}, nil
}

// Bind exchange to queue so that:
//
//	Exchange -> Queue
func (e *Exchange) Bind(q *Queue, key string) error {
	return e.ch.QueueBind(
		q.Name,
		key,
		e.Name,
		NO_WAIT,
		nil,
	)
}

func (e *Exchange) Send(ctx context.Context, msg *pb.Event) error {
	return e.ch.Send(ctx, e.Name, msg)
}
