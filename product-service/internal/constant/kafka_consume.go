// Package constant defines constants used in the order service for Kafka topics and event types.
package constant

// Topics that Product Service consumes from.
const (
	// TopicOrderLifecycle is the topic for order lifecycle events.
	TopicOrderLifecycle = "order.lifecycle" // OrderCreated, OrderUpdated, OrderCancelled, OrderCompleted, OrderShipped, OrderDelivered
)

// Consumer groups for Product Service.
const (
	// ConsumerGroupProductOrderEvents is the consumer group for order lifecycle events.
	ConsumerGroupProductOrderEvents = "product-service.order-events"
)

// Order Lifecycle Events.
const (
	// KafkaEventTypeOrderCreated is when customer places an order (pending).
	// Needed by: inventory reservation, payment service, fraud detection.
	KafkaEventTypeOrderCreated = "OrderCreated"
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
