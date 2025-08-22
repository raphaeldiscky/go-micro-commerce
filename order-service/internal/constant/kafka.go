// Package constant defines constants used in the order service for Kafka topics and event types.
package constant

// Order Service Source.
const (
	KafkaSourceOrderService = "order-service"
)

// Order Lifecycle Events.
const (
	// KafkaEventTypeOrderCreated is when customer places an order (pending).
	// Needed by: inventory reservation, payment service, fraud detection.
	KafkaEventTypeOrderCreated = "OrderCreated"
	// KafkaEventTypeOrderUpdated is when an order is updated (pending).
	// Needed by: inventory management, notification service.
	KafkaEventTypeOrderUpdated = "OrderUpdated"
	// KafkaEventTypeOrderConfirmed is after validation & inventory check (confirmed).
	// Needed by: payment service, notification service.
	KafkaEventTypeOrderConfirmed = "OrderConfirmed"
	// KafkaEventTypeOrderPaid is when payment succeeded (paid).
	// Needed by: shipping service, accounting, notification service.
	KafkaEventTypeOrderPaid = "OrderPaid"
	// KafkaEventTypeOrderShipped is when order handed to logistics (shipped).
	// Needed by: notification service, delivery tracking.
	KafkaEventTypeOrderShipped = "OrderShipped"
	// KafkaEventTypeOrderDelivered is after customer received package (delivered).
	// Needed by: loyalty points, analytics.
	KafkaEventTypeOrderDelivered = "OrderDelivered"
	// KafkaEventTypeOrderCanceled is when canceled by system or customer (canceled)
	// Needed by: inventory release, refund, analytics.
	KafkaEventTypeOrderCanceled = "OrderCanceled"
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

// Topics that Auth Service produces to.
const (
	// TopicOrderLifecycle is the topic for order lifecycle events.
	TopicOrderLifecycle = "order.lifecycle" // OrderCreated, OrderUpdated, OrderCancelled, OrderCompleted, OrderShipped, OrderDelivered
	// TopicOrderLifecycleNumPartitions is the number of partitions for the order lifecycle topic.
	TopicOrderLifecycleNumPartitions = 3
	// TopicOrderLifecycleReplicationFactor is the replication factor for the order lifecycle topic.
	TopicOrderLifecycleReplicationFactor = 1
	// TopicOrderPayment is the topic for order payment events.
	TopicOrderPayment = "order.payment" // PaymentProcessed, PaymentFailed, PaymentRefunded
	// TopicOrderInventory is the topic for order inventory events.
	TopicOrderInventory = "order.inventory" // InventoryReserved, InventoryReleased, InventoryUpdated
)

// Consumer groups for Order Service (consuming from other services).
const (
	// ConsumerGroupOrderAuthEvents is the consumer group for auth service events.
	ConsumerGroupOrderUserEvents = "order-service.user-events" // For user lifecycle
	// ConsumerGroupOrderProductEvents is the consumer group for product events.
	ConsumerGroupOrderProductEvents = "order-service.product-events" // For product availability
	// ConsumerGroupOrderInventoryEvents is the consumer group for inventory events.
	ConsumerGroupOrderInventoryEvents = "order-service.inventory-events" // For inventory management
)
