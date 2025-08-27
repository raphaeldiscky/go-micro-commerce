// Package constant defines constants used in the payment service for Kafka topics and event types.
package constant

// Payment Service Source.
const (
	KafkaSourcePaymentService = "payment-service"
)

// Payment Lifecycle Events.
const (
	// KafkaEventTypePaymentCreated is when customer places an payment (pending).
	// Needed by: inventory reservation, payment service, fraud detection.
	KafkaEventTypePaymentCreated = "PaymentCreated"
	// KafkaEventTypePaymentConfirmed is after validation & inventory check (confirmed).
	// Needed by: payment service, notification service.
	KafkaEventTypePaymentConfirmed = "PaymentConfirmed"
	// KafkaEventTypePaymentPaid is when payment succeeded (paid).
	// Needed by: shipping service, accounting, notification service.
	KafkaEventTypePaymentPaid = "PaymentPaid"
	// KafkaEventTypePaymentShipped is when payment handed to logistics (shipped).
	// Needed by: notification service, delivery tracking.
	KafkaEventTypePaymentShipped = "PaymentShipped"
	// KafkaEventTypePaymentDelivered is after customer received package (delivered).
	// Needed by: loyalty points, analytics.
	KafkaEventTypePaymentDelivered = "PaymentDelivered"
	// KafkaEventTypePaymentCanceled is when canceled by system or customer (canceled)
	// Needed by: inventory release, refund, analytics.
	KafkaEventTypePaymentCanceled = "PaymentCanceled"
)

// DLQ Event Types.
const (
	KafkaEventTypePaymentDLQ = "PaymentDLQ"
)

// Topics that Payment Service produces to.
const (
	// TopicPaymentLifecycle is the topic for payment lifecycle events.
	TopicPaymentLifecycle = "payment.lifecycle" // PaymentCreated, PaymentUpdated, PaymentCancelled, PaymentCompleted, PaymentShipped, PaymentDelivered
	// TopicPaymentLifecycleNumPartitions is the number of partitions for the payment lifecycle topic.
	TopicPaymentLifecycleNumPartitions = 3
	// TopicPaymentLifecycleReplicationFactor is the replication factor for the payment lifecycle topic.
	TopicPaymentLifecycleReplicationFactor = 1
)

const (
	// TopicPaymentDLQ is the dead-letter queue topic for failed payment events.
	TopicPaymentDLQ = "payment.dlq"
	// TopicPaymentDLQNumPartitions is the number of partitions for the payment DLQ topic.
	TopicPaymentDLQNumPartitions = 1
	// TopicPaymentDLQReplicationFactor is the replication factor for the payment DLQ topic.
	TopicPaymentDLQReplicationFactor = 1
)
