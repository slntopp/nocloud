package main

import (
	"fmt"
	"log"
	"time"

	"github.com/slntopp/nocloud/pkg/eventbus"
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

type Message struct {
	Text string `json:"text"`
}

func main() {

	rbmq, err := amqp.Dial(rbmq)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rbmq.Close()

	bus := eventbus.New(rbmq)

	counter := 0
	for {
		time.Sleep(time.Second)

		if err := bus.Pub(Message{
			Text: fmt.Sprint(counter),
		}, ""); err != nil {
			log.Println(err)
			continue
		}

		counter++

		log.Println("message sent")
	}

}
