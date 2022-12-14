package eventbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
	"google.golang.org/protobuf/proto"
)

type Queue struct {
	amqp091.Queue
	conn *Connection
}

type QueueType int64

const (
	DefaultQueue QueueType = iota
	UniqueQueue            // Add suffix to queue name
)

func NewQueue(ch *Connection, name string, t QueueType) (*Queue, error) {

	if t == UniqueQueue {
		name = fmt.Sprintf("%s.%s", name, uuid.New())
	}

	q, err := ch.Channel().QueueDeclare(name, QUEUE_DURABLE, QUEUE_AUTO_DELETE, QUEUE_EXCLUSIVE, NO_WAIT, nil)
	if err != nil {
		return nil, err
	}

	return &Queue{q, ch}, nil
}

// Consume events from the queue
func (q *Queue) Consume() (<-chan *pb.Event, error) {

	dels, err := q.conn.Channel().Consume(q.Name, q.Name, CONSUME_AUTO_ACK, QUEUE_EXCLUSIVE, false, NO_WAIT, nil)
	if err != nil {
		return nil, err
	}

	ch := make(chan *pb.Event)

	go func() {
		for del := range dels {
			event := &pb.Event{}
			if err := proto.Unmarshal(del.Body, event); err == nil {
				ch <- event
				del.Ack(false)
			}
		}
	}()

	return ch, nil
}

// Send event to default exchange with routing key equal queue name
func (q *Queue) Send(ctx context.Context, event *pb.Event) error {
	return q.conn.Send(ctx, "", event)
}
