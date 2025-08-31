package constant

// Topics that Order Service consumes from other services.
const (
	// TopicOrderLifecycle is the topic for order lifecycle events.
	TopicOrderLifecycle = "order.lifecycle"
)

// Consumer groups for Order Service (consuming from other services).
const (
	// ConsumerGroupPaymentOrderEvents is the consumer group for order events.
	ConsumerGroupPaymentOrderEvents = "payment-service.order-events" // For order lifecycle
)

// Order Service Event Types.
const (
	// KafkaEventTypeOrderCreated is the event type for order created events.
	KafkaEventTypeOrderCreated = "OrderCreated"
	// KafkaEventTypeOrderUpdated is the event type for order updated events.
	KafkaEventTypeOrderUpdated = "OrderUpdated"
	// KafkaEventTypeOrderDeleted is the event type for order deleted events.
	KafkaEventTypeOrderDeleted = "OrderDeleted"
	// KafkaEventTypeOrderPaymentRequested is the event type for payment request events.
	KafkaEventTypeOrderPaymentRequested = "OrderPaymentRequested"
)
