package billing

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/instances"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"google.golang.org/protobuf/proto"
	"time"
)

func NewInstanceDataPublisher(ch rabbitmq.Channel) Pub {
	return Publisher(ch, "datas", "instances")
}

type Pub func(msg *pb.ObjectData) (int, error)

func Publisher(ch rabbitmq.Channel, exchange, subtopic string) Pub {
	topic := exchange + "." + subtopic
	return func(msg *pb.ObjectData) (int, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		body, err := proto.Marshal(msg)
		if err != nil {
			return 0, err
		}
		return len(body), ch.PublishWithContext(ctx, exchange, topic, false, false, amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         body,
		})
	}
}
