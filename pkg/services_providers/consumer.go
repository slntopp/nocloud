package services_providers

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	"github.com/slntopp/nocloud-proto/events"
	ipb "github.com/slntopp/nocloud-proto/instances"
	"github.com/slntopp/nocloud/pkg/graph"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var consumers = map[string]struct{}{}
var ch *amqp.Channel

const exchange = "actions"
const topicTemplate = "actions.%s"
const queueTemplate = "actions.%s"

func (s *ServicesProviderServer) DriverActionsConsumersInit(conn *amqp.Connection) error {
	log := s.log.Named("DriverActionsConsumersInit")
	log.Debug("Starting")

	providers, err := s.ctrl.List(context.Background(), "0", true)
	if err != nil {
		log.Error("Failed to get services providers", zap.Error(err))
		return err
	}

	if ch == nil {
		ch, err = conn.Channel()
		if err != nil {
			log.Error("Failed to open a channel", zap.Error(err))
			return err
		}
	}

	if err = ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		log.Error("Failed to declare an exchange", zap.Error(err))
		return err
	}

	for _, sp := range providers {
		log := log.With(zap.String("sp", sp.GetUuid()))
		if _, ok := consumers[sp.GetUuid()]; !ok {
			if err = s.RegisterConsumer(log, ch, sp); err != nil {
				return err
			}
			consumers[sp.GetUuid()] = struct{}{}
		}
	}

	return nil
}

func (s *ServicesProviderServer) RegisterConsumer(log *zap.Logger, ch *amqp.Channel, sp *graph.ServicesProvider) error {
	log = log.Named("RegisterConsumer")

	q, err := ch.QueueDeclare(fmt.Sprintf(queueTemplate, sp.GetUuid()), true, false, false, false, nil)
	if err != nil {
		log.Error("Failed to declare a queue", zap.Error(err))
		return err
	}

	err = ch.QueueBind(q.Name, fmt.Sprintf(topicTemplate, sp.GetUuid()), exchange, false, nil)
	if err != nil {
		log.Error("Failed to bind a queue", zap.Error(err))
		return err
	}

	messages, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Error("Failed to register a consumer", zap.Error(err))
		return err
	}

	for msg := range messages {
		var event *events.Event
		err := proto.Unmarshal(msg.Body, event)
		if err != nil {
			log.Error("Failed to unmarshal request", zap.Error(err))
			continue
		}
		log.Debug("Received event", zap.Any("event", event))
		if err = s.Consumer(log, msg, event, sp); err != nil {
			log.Error("Failed to process event", zap.Error(err))
			continue
		}
	}

	return nil
}

func (s *ServicesProviderServer) Consumer(log *zap.Logger, msg amqp.Delivery, event *events.Event, sp *graph.ServicesProvider) error {
	log = log.Named("Consumer")

	if event.Key == "invoke" {
		log.Debug("Received invoke event", zap.Any("event", event))
		var invokeReq = &ipb.InvokeRequest{}
		jsonBody := event.Data["encoded"].GetStringValue()
		if err := json.Unmarshal([]byte(jsonBody), invokeReq); err != nil {
			log.Error("Failed to unmarshal request", zap.Error(err))
			return err
		}
		if invokeReq == nil || invokeReq.GetUuid() == "" {
			log.Error("Incorrect invoke struct", zap.Any("struct", invokeReq))
			return fmt.Errorf("incorrect invoke struct")
		}

		resp, err := s.instances.GetGroup(context.TODO(), invokeReq.GetUuid())
		if err != nil {
			log.Error("Failed to get group", zap.Error(err))
			return err
		}
		if resp.SP.GetUuid() != sp.GetUuid() {
			log.Error("Got response from wrong provider", zap.String("sp", sp.GetUuid()))
			return fmt.Errorf("got response from wrong provider")
		}

		client, ok := s.drivers[sp.GetType()]
		if !ok {
			log.Error("Failed to get driver", zap.String("type", sp.GetType()))
			return fmt.Errorf("driver not found")
		}

		client.Invoke(context.TODO(), &driverpb.InvokeRequest{
			Instance: invokeReq.GetUuid(),
		})
	}

	return nil
}
