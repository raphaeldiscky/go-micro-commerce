// Package constant defines constants used in the order service for Kafka topics and event types.
package constant

// Order Service Source.
const (
	KafkaSourceOrderService = "order-service"
)

// Order Lifecycle Events.
const (
	// TopicOrderLifecycle is the topic for order lifecycle events.
	TopicOrderLifecycle = "order.lifecycle" // OrderCreated, OrderUpdated, OrderCancelled, OrderCompleted, OrderShipped, OrderDelivered
	// TopicOrderLifecycleNumPartitions is the number of partitions for the order lifecycle topic.
	TopicOrderLifecycleNumPartitions = 3
	// TopicOrderLifecycleReplicationFactor is the replication factor for the order lifecycle topic.
	TopicOrderLifecycleReplicationFactor = 1
	// KafkaEventTypeOrderCreated is when customer places an order (pending).
	// Needed by: inventory reservation, payment service, fraud detection.
	KafkaEventTypeOrderCreated = "OrderCreated"
	// KafkaEventTypeOrderPaid is when payment succeeded (paid).
	// Needed by: shipping service, accounting, notification service.
	// KafkaEventTypeOrderProcessing is when order is being processed (processing).
	KafkaEventTypeOrderProcessing = "OrderProcessing"
	// KafkaEventTypeOrderConfirmed is when order is confirmed (confirmed).
	// Needed by: inventory service to update stock, notification service.
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
	// KafkaEventTypeOrderFailed is when order processing failed (failed).
	// Needed by: notification service, analytics.
	KafkaEventTypeOrderFailed = "OrderFailed"
)

// Topics that Order Service produces to.
const (
	// TopicPaymentRequest is the topic for payment events.
	TopicPaymentRequest = "payment.request"
	// TopicPaymentRequestNumPartitions is the number of partitions for the payment request topic.
	TopicPaymentRequestNumPartitions = 3
	// TopicPaymentRequestReplicationFactor is the replication factor for the payment request topic.
	TopicPaymentRequestReplicationFactor = 1
	// KafkaEventTypePaymentRequested is when payment is requested for an order (pending).
	// Needed by: payment service to create payment record and process payment.
	KafkaEventTypePaymentRequested = "PaymentRequested"
)

const (
	// TopicOrderDLQ is the dead-letter queue topic for failed order events.
	TopicOrderDLQ = "order.dlq"
	// TopicPaymentDLQ is the dead-letter queue topic for failed payment events.
	TopicPaymentDLQ = "payment.dlq"
	// KafkaEventTypeOrderDLQ is the event type for order DLQ events.
	KafkaEventTypeOrderDLQ = "OrderDLQ"
	// KafkaEventTypePaymentDLQ is the event type for payment DLQ events.
	KafkaEventTypePaymentDLQ = "PaymentDLQ"
	// TopicOrderDLQNumPartitions is the number of partitions for the order DLQ topic.
	TopicOrderDLQNumPartitions = 1
	// TopicOrderDLQReplicationFactor is the replication factor for the order DLQ topic.
	TopicOrderDLQReplicationFactor = 1
	// TopicPaymentDLQNumPartitions is the number of partitions for the payment DLQ topic.
	TopicPaymentDLQNumPartitions = 1
	// TopicPaymentDLQReplicationFactor is the replication factor for the payment DLQ topic.
	TopicPaymentDLQReplicationFactor = 1
	// TopicOrderDLQNumPartitions is the number of partitions for the order DLQ topic.
)
