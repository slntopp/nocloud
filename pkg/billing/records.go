/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

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
package billing

import (
	"context"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "github.com/slntopp/nocloud/pkg/billing/proto"
	"github.com/slntopp/nocloud/pkg/graph"
	healthpb "github.com/slntopp/nocloud/pkg/health/proto"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/access"
)

var (
	RabbitMQConn string
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("RABBITMQ_CONN", "amqp://nocloud:secret@rabbitmq:5672/")
	RabbitMQConn = viper.GetString("RABBITMQ_CONN")
}

type RecordsServiceServer struct {
	pb.UnimplementedRecordsServiceServer
	log     *zap.Logger
	rbmq    *amqp.Connection
	records graph.RecordsController

	db driver.Database

	ConsumerStatus *healthpb.RoutineStatus
}

func NewRecordsServiceServer(logger *zap.Logger, db driver.Database) *RecordsServiceServer {
	log := logger.Named("RecordsService")
	rbmq, err := amqp.Dial(RabbitMQConn)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}

	records := graph.NewRecordsController(log, db)

	return &RecordsServiceServer{
		log:     log,
		rbmq:    rbmq,
		records: records,

		db: db,
		ConsumerStatus: &healthpb.RoutineStatus{
			Routine: "Records Consumer",
			Status: &healthpb.ServingStatus{
				Service: "Billing Machine",
				Status:  healthpb.Status_STOPPED,
			},
		},
	}
}

func (s *RecordsServiceServer) Consume(ctx context.Context) {
	log := s.log.Named("Consumer")
init:
	ch, err := s.rbmq.Channel()
	if err != nil {
		log.Error("Failed to open a channel", zap.Error(err))
		time.Sleep(time.Second)
		goto init
	}

	queue, _ := ch.QueueDeclare(
		"records",
		true, false, false, true, nil,
	)

	records, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Error("Failed to register a consumer", zap.Error(err))
		time.Sleep(time.Second)
		goto init
	}

	s.ConsumerStatus.Status.Status = healthpb.Status_RUNNING

	for msg := range records {
		log.Debug("Received a message")
		var record pb.Record
		err = proto.Unmarshal(msg.Body, &record)
		log.Debug("Message unmarshalled", zap.Any("record", &record))
		if err != nil {
			log.Error("Failed to unmarshal record", zap.Error(err))
			continue
		}
		if record.Total == 0 {
			log.Warn("Got zero record, skipping", zap.Any("record", &record))
			continue
		}

		s.records.Create(ctx, &record)
		s.ConsumerStatus.LastExecution = time.Now().Format("2006-01-02T15:04:05Z07:00")
	}
}

func (s *BillingServiceServer) GetRecords(ctx context.Context, req *pb.Transaction) (*pb.Records, error) {
	log := s.log.Named("GetRecords")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if req.Uuid == "" {
		log.Error("Request has no UUID", zap.String("requestor", requestor))
		return nil, status.Error(codes.InvalidArgument, "Request has no UUID")
	}

	tr, err := s.transactions.Get(ctx, req.Uuid)
	if err != nil {
		log.Error("Failed to get transaction", zap.String("requestor", requestor), zap.String("uuid", req.Uuid))
		return nil, status.Error(codes.NotFound, "Transaction not found")
	}

	ok := graph.HasAccess(ctx, s.db, requestor, tr.Account, access.SUDO)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Permission denied")
	}

	recs, err := s.records.Get(ctx, tr.Records)
	if err != nil {
		log.Error("Failed to get records", zap.String("requestor", requestor), zap.String("uuid", req.Uuid))
		return nil, status.Error(codes.Internal, "Failed to get Records")
	}

	pool := make([]*pb.Record, len(recs))
	for i, rec := range recs {
		pool[i] = rec.Record
	}

	return &pb.Records{
		Pool: pool,
	}, nil
}
