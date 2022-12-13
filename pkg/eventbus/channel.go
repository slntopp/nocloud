package eventbus

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
	"google.golang.org/protobuf/proto"
)

type Channel struct {
	*amqp091.Channel
}

func NewChannel(ch *amqp091.Channel) *Channel {
	return &Channel{ch}
}

func (c *Channel) Send(ctx context.Context, exchange string, event *pb.Event) error {
	bytes, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	return c.PublishWithContext(ctx, exchange, event.Key, false, false, amqp091.Publishing{
		Body: bytes,
	})
}
