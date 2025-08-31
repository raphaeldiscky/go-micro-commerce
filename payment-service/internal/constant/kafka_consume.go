package constant

// Topics that Payment Service consumes from other services.
const (
	// TopicOrderLifecycle is the topic for order lifecycle events.
	TopicOrderLifecycle = "order.lifecycle"
	// Note: TopicPaymentLifecycle is defined in kafka_produce.go to avoid redeclaration.
)

// Consumer groups for Payment Service (consuming from other services).
const (
	// ConsumerGroupPaymentOrderEvents is the consumer group for order events.
	ConsumerGroupPaymentOrderEvents = "payment-service.order-events" // For order lifecycle
	// ConsumerGroupPaymentEvents is the consumer group for payment request events.
	ConsumerGroupPaymentEvents = "payment-service.payment-events" // For payment requests
)

// Order Service Event Types.
const (
	// KafkaEventTypeOrderCreated is the event type for order created events.
	KafkaEventTypeOrderCreated = "OrderCreated"
	// KafkaEventTypeOrderUpdated is the event type for order updated events.
	KafkaEventTypeOrderUpdated = "OrderUpdated"
	// KafkaEventTypeOrderDeleted is the event type for order deleted events.
	KafkaEventTypeOrderDeleted = "OrderDeleted"
)

// Payment Request Event Types from Order Service.
const (
	// KafkaEventTypePaymentRequested is the event type for payment request events.
	KafkaEventTypePaymentRequested = "PaymentRequested"
)
