package instances

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	epb "github.com/slntopp/nocloud-proto/events"
	pb "github.com/slntopp/nocloud-proto/instances"
	ps "github.com/slntopp/nocloud/pkg/pubsub"
	"github.com/slntopp/nocloud/pkg/pubsub/services_registry"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *InstancesServer) ProcessInvokeCommands(_ *zap.Logger, ctx context.Context, event *epb.Event) error {
	if event.GetKey() != services_registry.CommandInstanceInvoke {
		return nil
	}

	if event.GetUuid() == "" {
		return ps.NoNackErr(fmt.Errorf("no instance uuid"))
	}
	if event.Data == nil {
		event.Data = make(map[string]*structpb.Value)
	}
	req := connect.NewRequest(&pb.InvokeRequest{
		Uuid:   event.GetUuid(),
		Method: event.GetType(),
		Params: event.Data,
	})
	if _, err := s.Invoke(ctx, req); err != nil {
		return fmt.Errorf("invoke failed: %w", err)
	}

	return nil
}

func (s *InstancesServer) ConsumeInvokeCommands(log *zap.Logger, ctx context.Context, p *ps.PubSub[*epb.Event]) {
	log = log.Named("ConsumeInvokeCommands")
	opt := ps.ConsumeOptions{
		Durable:    true,
		NoWait:     false,
		Exclusive:  false,
		WithRetry:  true,
		DelayMilli: 300 * 1000, // Every 5 minute
		MaxRetries: 36,         // 3 hours in general
	}
	msgs, err := p.Consume("instances-invoke", ps.DEFAULT_EXCHANGE, services_registry.Topic("instances-commands"), opt)
	if err != nil {
		log.Fatal("Failed to start consumer")
		return
	}

	for msg := range msgs {
		var event epb.Event
		if err = proto.Unmarshal(msg.Body, &event); err != nil {
			log.Error("Failed to unmarshal event. Incorrect delivery. Skip", zap.Error(err))
			if err = msg.Ack(false); err != nil {
				log.Error("Failed to acknowledge the delivery", zap.Error(err))
			}
			continue
		}
		log.Debug("Pubsub event received", zap.String("key", event.Key), zap.String("type", event.Type))
		if err = s.ProcessInvokeCommands(log, ctx, &event); err != nil {
			ps.HandleAckNack(log, msg, err)
			continue
		}
		if err = msg.Ack(false); err != nil {
			log.Error("Failed to acknowledge the delivery", zap.Error(err))
		}
	}
}
