package kafka

const (
	// PaymentDLQEventType is the event type for payment DLQ events.
	PaymentDLQEventType = "PaymentDLQ"
	// OrderDLQEventType is the event type for order DLQ events.
	OrderDLQEventType = "OrderDLQ"
)

const (
	// PaymentDLQTopic is the dead-letter queue topic for failed payment events.
	PaymentDLQTopic = "payment.dlq"
	// OrderDLQTopic is the dead-letter queue topic for failed order events.
	OrderDLQTopic = "order.dlq"
)
