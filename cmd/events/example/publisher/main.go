package main

import (
	"log"

	"github.com/spf13/viper"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

var (
	rbmq string
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("RABBITMQ_CONN", "amqp://nocloud:secret@localhost:5672/")
	rbmq = viper.GetString("RABBITMQ_CONN")
}

func main() {

	rbmq, err := amqp.Dial(rbmq)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rbmq.Close()

	// bus := event.NewBus(rbmq)
	// messageBus := event.NewTypedBus[]()

}
