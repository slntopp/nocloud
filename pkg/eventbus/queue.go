package eventbus

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
	"google.golang.org/protobuf/proto"
)

type Queue struct {
	amqp091.Queue
	ch *Channel
}

func NewQueue(ch *Channel, name string) (*Queue, error) {
	q, err := ch.QueueDeclare(name, QUEUE_DURABLE, QUEUE_AUTO_DELETE, QUEUE_EXCLUSIVE, NO_WAIT, nil)
	if err != nil {
		return nil, err
	}

	return &Queue{q, ch}, nil
}

func (q *Queue) Consume() (<-chan *pb.Event, error) {
	dels, err := q.ch.Consume(q.Name, "", CONSUME_AUTO_ACK, QUEUE_EXCLUSIVE, false, NO_WAIT, nil)
	if err != nil {
		return nil, err
	}

	ch := make(chan *pb.Event)

	go func() {
		for del := range dels {
			event := &pb.Event{}
			proto.Unmarshal(del.Body, event)
			ch <- event
		}
	}()

	return ch, nil
}

func (q *Queue) Send(ctx context.Context, event *pb.Event) error {
	return q.ch.Send(ctx, "", event)
}
