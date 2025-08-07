// Package constant defines constants used in the order service for Kafka topics and event types.
package constant

// Order Service Source
const (
	KafkaSourceOrderService = "order-service"
)

// Order Service Event Types
const (
	// KafkaEventTypeOrderCreated is the event type for order creation events.
	KafkaEventTypeOrderCreated = "OrderCreated"
	// KafkaEventTypeOrderUpdated is the event type for order update events.
	KafkaEventTypeOrderUpdated = "OrderUpdated"
	// KafkaEventTypeOrderCancelled is the event type for order cancellation events.
	KafkaEventTypeOrderCancelled = "OrderCancelled"
	// KafkaEventTypeOrderCompleted is the event type for order completion events.
	KafkaEventTypeOrderCompleted = "OrderCompleted"
	// KafkaEventTypeOrderShipped is the event type for order shipping events.
	KafkaEventTypeOrderShipped = "OrderShipped"
	// KafkaEventTypeOrderDelivered is the event type for order delivery events.
	KafkaEventTypeOrderDelivered = "OrderDelivered"
	// KafkaEventTypePaymentProcessed is the event type for payment processing events.
	KafkaEventTypePaymentProcessed = "PaymentProcessed"
	// KafkaEventTypePaymentFailed is the event type for payment failure events.
	KafkaEventTypePaymentFailed = "PaymentFailed"
	// KafkaEventTypePaymentRefunded is the event type for payment refund events.
	KafkaEventTypePaymentRefunded = "PaymentRefunded"
	// KafkaEventTypeInventoryReserved is the event type for inventory reservation events.
	KafkaEventTypeInventoryReserved = "InventoryReserved"
	// KafkaEventTypeInventoryReleased is the event type for inventory release events.
	KafkaEventTypeInventoryReleased = "InventoryReleased"
	// KafkaEventTypeInventoryUpdated is the event type for inventory update events.
)

// Topics that Auth Service produces to
const (
	// TopicOrderLifecycle is the topic for order lifecycle events.
	TopicOrderLifecycle = "order.lifecycle" // OrderCreated, OrderUpdated, OrderCancelled, OrderCompleted, OrderShipped, OrderDelivered
	// TopicOrderPayment is the topic for order payment events.
	TopicOrderPayment = "order.payment" // PaymentProcessed, PaymentFailed, PaymentRefunded
	// TopicOrderInventory is the topic for order inventory events.
	TopicOrderInventory = "order.inventory" // InventoryReserved, InventoryReleased, InventoryUpdated
)

// Consumer groups for Order Service (consuming from other services)
const (
	// ConsumerGroupOrderAuthEvents is the consumer group for auth service events.
	ConsumerGroupOrderUserEvents = "order-service.user-events" // For user lifecycle
	// ConsumerGroupOrderProductEvents is the consumer group for product events.
	ConsumerGroupOrderProductEvents = "order-service.product-events" // For product availability
)
