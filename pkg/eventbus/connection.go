package eventbus

import (
	"context"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
	"google.golang.org/protobuf/proto"
)

// Connection wraps amqp connection to handle reconnects
type Connection struct {
	ch   rabbitmq.Channel
	conn rabbitmq.Connection
}

func NewConnection(conn rabbitmq.Connection) (*Connection, error) {

	c := &Connection{conn: conn}

	ch, err := c.newChannel()
	if err != nil {
		return nil, err
	}

	c.ch = ch

	return c, nil
}

// Get existing channel if open. Otherwise open new channel
func (c *Connection) Channel() rabbitmq.Channel {
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

func (c *Connection) newChannel() (rabbitmq.Channel, error) {

	channel, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := channel.Qos(PREFETCH_COUNT, PREFETCH_SIZE, PREFETCH_GLOBAL); err != nil {
		return nil, err
	}

	return channel, nil
}

// Send event to given exchange
func (c *Connection) Send(ctx context.Context, exchange string, event *pb.Event) error {
	bytes, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	return c.ch.PublishWithContext(ctx, exchange, Topic(event), PUBLISH_IMEDIATE, PUBLISH_MANDATORY, amqp091.Publishing{
		Body: bytes,
	})
}
