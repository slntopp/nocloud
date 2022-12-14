package eventbus

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
)

type Exchange struct {
	conn *Connection
	Name string
}

type ExchangeType int64

const (
	DefaultExchange ExchangeType = iota
	AlternateExchange
)

func NewExchange(conn *Connection, name string, t ExchangeType) (*Exchange, error) {

	args := amqp091.Table{}

	if t == AlternateExchange {
		args["alternate-exchange"] = EXCHANGE_BUFFER
	}

	if err := conn.Channel().ExchangeDeclare(name, EXCHANGE_KIND, EXCHANGE_DURABLE, EXCHANGE_AUTO_DELETE, EXCHANGE_INTERNAL, NO_WAIT, args); err != nil {
		return nil, err
	}

	return &Exchange{
		conn: conn,
		Name: name,
	}, nil
}

// Bind exchange to queue so that:
//
//	Exchange -> Queue
func (e *Exchange) Bind(q *Queue, key string) error {
	return e.conn.Channel().QueueBind(
		q.Name,
		key,
		e.Name,
		NO_WAIT,
		nil,
	)
}

// Send event to the exchange
func (e *Exchange) Send(ctx context.Context, event *pb.Event) error {
	return e.conn.Send(ctx, e.Name, event)
}

// Create queue that is binded to exchange
func (e *Exchange) DeriveQueue(name string) (*Queue, error) {

	q, err := NewQueue(e.conn, name, DefaultQueue)
	if err != nil {
		return nil, err
	}

	if err := e.Bind(q, name); err != nil {
		return nil, err
	}

	return q, nil
}
