package rabbitmq

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Connection is an interface wrapper for [amqp.Connection]
type Connection interface {
	Channel() (Channel, error)
	IsClosed() bool
	Close() error
}

// Channel is an interface wrapper for [amqp.Channel]
type Channel interface {
	ExchangeDeclare(name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args amqp.Table) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name string, key string, exchange string, noWait bool, args amqp.Table) error
	QueueDelete(name string, ifUnused, ifEmpty, noWait bool) (int, error)
	Consume(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	PublishWithContext(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error
	Close() error
	IsClosed() bool
	Cancel(consumer string, noWait bool) error
	Qos(prefetchCount int, prefetchSize int, global bool) error
	Get(queue string, autoAck bool) (msg amqp.Delivery, ok bool, err error)
}

type client struct {
	*amqp.Connection
}

func (c *client) Channel() (Channel, error) {
	ch, err := c.Connection.Channel()
	return ch, err
}

func NewRabbitMQConnection(conn *amqp.Connection) Connection {
	return &client{
		Connection: conn,
	}
}
