package eventbus

const (
	// Consume properties
	CONSUME_AUTO_ACK = false

	// Common properties
	NO_WAIT = false

	// Exchange properties
	EXCHANGE_NAME        = "nocloud-event-bus"
	EXCHANGE_BUFFER      = EXCHANGE_NAME + "-buffer"
	EXCHANGE_DURABLE     = true // essential for retention
	EXCHANGE_AUTO_DELETE = false
	EXCHANGE_INTERNAL    = false
	EXCHANGE_NO_WAIT     = false
	EXCHANGE_KIND        = "topic"

	// Queue properties
	QUEUE_DURABLE     = true
	QUEUE_AUTO_DELETE = true
	QUEUE_EXCLUSIVE   = true
)
