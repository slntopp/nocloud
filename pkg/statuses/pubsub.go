/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package statuses

import (
	"context"
	"log"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/cskr/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
	spb "github.com/slntopp/nocloud-proto/statuses"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	RabibtMQConn string
	ps           *pubsub.PubSub
)

type StatusesPubSub struct {
	log  *zap.Logger
	db   *driver.Database
	rbmq *amqp.Connection
}

func NewStatesPubSub(log *zap.Logger, db *driver.Database, rbmq *amqp.Connection) *StatusesPubSub {
	sps := &StatusesPubSub{
		log: log.Named("StatusesPubSub"), rbmq: rbmq,
	}
	if db != nil {
		sps.db = db
	}
	return sps
}

func (s *StatusesPubSub) Channel() *amqp.Channel {
	log := s.log.Named("Channel")

	ch, err := s.rbmq.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel", zap.Error(err))
	}
	return ch
}

func (s *StatusesPubSub) TopicExchange(ch *amqp.Channel, name string) {
	log := s.log.Named("TopicExchange")

	err := ch.ExchangeDeclare(
		name, "topic", true, false, false, false, nil,
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange", zap.Error(err))
	}
}

func (s *StatusesPubSub) StatusesConsumerInit(ch *amqp.Channel, exchange, subtopic, col string) {
	if s.db == nil {
		log.Fatal("Failed to initialize statuses consumer, database is not set")
	}
	topic := exchange + "." + subtopic
	q, err := ch.QueueDeclare(
		topic, false, false, true, false, nil,
	)
	if err != nil {
		log.Fatal("Failed to declare a queue", zap.Error(err))
	}

	err = ch.QueueBind(q.Name, topic, exchange, false, nil)
	if err != nil {
		log.Fatal("Failed to bind a queue", zap.Error(err))
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to register a consumer", zap.Error(err))
	}
	go s.Consumer(col, msgs)
}

func (s *StatusesPubSub) Consumer(col string, msgs <-chan amqp.Delivery) {

}

type Pub func(msg *spb.ObjectStatus) (int, error)

func (s *StatusesPubSub) Publisher(ch *amqp.Channel, exchange, subtopic string) Pub {
	topic := exchange + "." + subtopic
	return func(msg *spb.ObjectStatus) (int, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		body, err := proto.Marshal(msg)
		if err != nil {
			return 0, err
		}
		return len(body), ch.PublishWithContext(ctx, exchange, topic, false, false, amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		})
	}
}
