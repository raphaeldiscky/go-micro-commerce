package kafka

const (
	// PaymentDLQEventType is the event type for payment DLQ events.
	PaymentDLQEventType = "PaymentDLQ"
	// OrderDLQEventType is the event type for order DLQ events.
	OrderDLQEventType = "OrderDLQ"
	// FulfillmentDLQEventType is the event type for fulfillment DLQ events.
	FulfillmentDLQEventType = "FulfillmentDLQ"
	// NotificationDLQEventType is the event type for notification DLQ events.
	NotificationDLQEventType = "NotificationDLQ"
)

const (
	// PaymentDLQTopic is the dead-letter queue topic for failed payment events.
	PaymentDLQTopic = "payment.dlq"
	// OrderDLQTopic is the dead-letter queue topic for failed order events.
	OrderDLQTopic = "order.dlq"
	// FulfillmentDLQTopic is the dead-letter queue topic for failed fulfillment events.
	FulfillmentDLQTopic = "fulfillment.dlq"
	// NotificationDLQTopic is the dead-letter queue topic for failed notification events.
	NotificationDLQTopic = "notification.dlq"
)
