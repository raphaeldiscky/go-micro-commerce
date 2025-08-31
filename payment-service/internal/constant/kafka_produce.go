// Package constant defines constants used in the payment service for Kafka topics and event types.
package constant

// Payment Service Source.
const (
	KafkaSourcePaymentService = "payment-service"
)

// Payment Lifecycle Events.
const (
	// KafkaEventTypePaymentCreated is the event type for payment created events.
	KafkaEventTypePaymentCreated = "PaymentCreated"
	// KafkaEventTypePaymentProcessing is the event type for payment processing events.
	KafkaEventTypePaymentProcessing = "PaymentProcessing"
	// KafkaEventTypePaymentCompleted is the event type for payment completed events.
	KafkaEventTypePaymentCompleted = "PaymentCompleted"
	// KafkaEventTypePaymentFailed is the event type for payment failure events.
	KafkaEventTypePaymentFailed = "PaymentFailed"
	// KafkaEventTypePaymentRefunded is the event type for payment refund events.
	KafkaEventTypePaymentRefunded = "PaymentRefunded"
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
