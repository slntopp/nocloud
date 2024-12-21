package rabbitmq

import (
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func FatalOnConnectionClose(log *zap.Logger, conn *amqp091.Connection) {
	go func() {
		ch := make(chan *amqp091.Error)
		ch = conn.NotifyClose(ch)
		err := <-ch
		if err == nil {
			log.Info("RabbitMQ Connection was closed by the opener (normal close)")
			return
		}
		log.Fatal("RabbitMQ Connection was closed unexpected", zap.Any("attached_error", err))
	}()
}
