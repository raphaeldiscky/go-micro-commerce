package constant

// Topics that Order Service consumes from other services.
const (
	// TopicProductLifecycle is the topic for product lifecycle events.
	TopicProductLifecycle = "product.lifecycle"
)

// Consumer groups for Order Service (consuming from other services).
const (
	// ConsumerGroupOrderAuthEvents is the consumer group for auth service events.
	ConsumerGroupOrderUserEvents = "payment-service.user-events" // For user lifecycle
	// ConsumerGroupOrderProductEvents is the consumer group for product events.
	ConsumerGroupOrderProductEvents = "payment-service.product-events" // For product lifecycle
	// ConsumerGroupOrderInventoryEvents is the consumer group for inventory events.
	ConsumerGroupOrderInventoryEvents = "payment-service.inventory-events" // For inventory management
)

// Product Service Event Types.
const (
	// KafkaEventTypeProductCreated is the event type for product created events.
	KafkaEventTypeProductCreated = "ProductCreated"
	// KafkaEventTypeProductUpdated is the event type for product updated events.
	KafkaEventTypeProductUpdated = "ProductUpdated"
	// KafkaEventTypeProductDeleted is the event type for product deleted events.
	KafkaEventTypeProductDeleted = "ProductDeleted"
	// KafkaEventTypeProductDeleted is the event type for product deleted events.
)

// Order Payment Events.
const (
	// KafkaEventTypePaymentProcessed is the event type for payment processing events.
	KafkaEventTypePaymentProcessed = "PaymentProcessed"
	// KafkaEventTypePaymentFailed is the event type for payment failure events.
	KafkaEventTypePaymentFailed = "PaymentFailed"
	// KafkaEventTypePaymentRefunded is the event type for payment refund events.
	KafkaEventTypePaymentRefunded = "PaymentRefunded"
)

// Order Inventory Events.
const (
	// KafkaEventTypeInventoryReserved is the event type for inventory reservation events.
	KafkaEventTypeInventoryReserved = "InventoryReserved"
	// KafkaEventTypeInventoryReleased is the event type for inventory release events.
	KafkaEventTypeInventoryReleased = "InventoryReleased"
	// KafkaEventTypeInventoryUpdated is the event type for inventory update events.
	KafkaEventTypeInventoryUpdated = "InventoryUpdated"
)
