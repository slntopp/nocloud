package eventbus

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
	"google.golang.org/protobuf/proto"
)

type Connection struct {
	ch   *amqp091.Channel
	conn *amqp091.Connection
}

func NewConnection(conn *amqp091.Connection) (*Connection, error) {

	c := &Connection{conn: conn}

	ch, err := c.newChannel()
	if err != nil {
		return nil, err
	}

	c.ch = ch

	return c, nil
}

func (c *Connection) Channel() *amqp091.Channel {
	if c.ch.IsClosed() {

		ch, err := c.newChannel()
		if err != nil {
			return nil
		}

		c.ch = ch
		return ch
	}

	return c.ch
}

func (c *Connection) newChannel() (*amqp091.Channel, error) {

	channel, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := channel.Qos(PREFETCH_COUNT, PREFETCH_SIZE, PREFETCH_GLOBAL); err != nil {
		return nil, err
	}

	return channel, nil
}

func (c *Connection) Send(ctx context.Context, exchange string, event *pb.Event) error {
	bytes, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	return c.ch.PublishWithContext(ctx, exchange, event.Key, PUBLISH_IMEDIATE, PUBLISH_MANDATORY, amqp091.Publishing{
		Body: bytes,
	})
}
