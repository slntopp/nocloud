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
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"log"
	"time"

	"github.com/arangodb/go-driver"
	amqp "github.com/rabbitmq/amqp091-go"
	spb "github.com/slntopp/nocloud-proto/statuses"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type StatusesPubSub struct {
	log  *zap.Logger
	db   *driver.Database
	rbmq rabbitmq.Connection
}

func NewStatusesPubSub(log *zap.Logger, db *driver.Database, rbmq rabbitmq.Connection) *StatusesPubSub {
	sps := &StatusesPubSub{
		log: log.Named("StatusesPubSub"), rbmq: rbmq,
	}
	if db != nil {
		sps.db = db
	}
	return sps
}

func (s *StatusesPubSub) Channel() rabbitmq.Channel {
	log := s.log.Named("Channel")

	ch, err := s.rbmq.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel", zap.Error(err))
	}
	return ch
}

func (s *StatusesPubSub) TopicExchange(ch rabbitmq.Channel, name string) {
	log := s.log.Named("TopicExchange")

	err := ch.ExchangeDeclare(
		name, "topic", true, false, false, false, nil,
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange", zap.Error(err))
	}
}

func (s *StatusesPubSub) StatusesConsumerInit(ch rabbitmq.Channel, exchange, subtopic, col string) {
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

const updateStatusQuery = `
UPDATE DOCUMENT (@@collection, @key) WITH { status: @status } IN @@collection OPTIONS {mergeObjects: false}
`

func (s *StatusesPubSub) Consumer(col string, msgs <-chan amqp.Delivery) {
	log := s.log.Named(col)
	for msg := range msgs {
		log.Debug("Status upd message", zap.Any("message", msg))
		var req spb.ObjectStatus
		err := proto.Unmarshal(msg.Body, &req)
		if err != nil {
			log.Error("Failed to unmarshal request", zap.Error(err))
			if err = msg.Ack(false); err != nil {
				log.Warn("Failed to Ack while unmarshal", zap.Error(err))
			}
			continue
		}
		log.Debug("Status req", zap.Any("status", req.GetStatus()))

		if req.Status == nil {
			log.Warn("Status is nil", zap.String("obj", col))
			if err = msg.Ack(false); err != nil {
				log.Warn("Failed to Acknowledge the delivery when Status is nil", zap.Error(err))
			}
			continue
		}

		db := *s.db
		ctx := context.Background()
		trID, err := db.BeginTransaction(ctx, driver.TransactionCollections{
			Write: []string{col},
		}, &driver.BeginTransactionOptions{})
		if err != nil {
			log.Error("Failed to start transaction to update status", zap.Error(err))
			if err = msg.Nack(false, false); err != nil {
				log.Error("Failed to Acknowledge the delivery while start transaction to update status", zap.Error(err))
			}
			continue
		}
		ctx = driver.WithTransactionID(ctx, trID)
		c, err := (*s.db).Query(ctx, updateStatusQuery, map[string]interface{}{
			"@collection": col,
			"key":         req.Uuid,
			"status":      req.Status.GetStatus(),
		})
		if err != nil {
			_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
			log.Error("Failed to update status", zap.Error(err))
			if err = msg.Nack(false, false); err != nil {
				log.Warn("Failed to Negatively Acknowledge the delivery while Update db", zap.Error(err))
			}
			continue
		}
		if err := db.CommitTransaction(ctx, trID, &driver.CommitTransactionOptions{}); err != nil {
			log.Error("Failed to commit transaction to update status", zap.Error(err))
			if err = msg.Nack(false, false); err != nil {
				log.Error("Failed to Acknowledge the delivery while commit transaction to update status", zap.Error(err))
			}
			continue
		}

		log.Debug("Updated status", zap.String("type", col), zap.String("uuid", req.Uuid))
		if err = c.Close(); err != nil {
			log.Warn("Failed to close database cursor connection", zap.Error(err))
		}
		if err = msg.Ack(false); err != nil {
			log.Warn("Failed to Acknowledge the delivery", zap.Error(err))
		}
	}
}

type Pub func(msg *spb.ObjectStatus) (int, error)

func (s *StatusesPubSub) Publisher(ch rabbitmq.Channel, exchange, subtopic string) Pub {
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
