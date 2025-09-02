package kafka

const (
	// PaymentRequestTopic is the topic for payment events.
	PaymentRequestTopic = "payment.request"
	// PaymentLifecycleTopic is the topic for payment lifecycle events.
	PaymentLifecycleTopic = "payment.lifecycle" // PaymentCreated, PaymentUpdated, PaymentCancelled, PaymentCompleted, PaymentShipped, PaymentDelivered
)

const (
	// PaymentRequestedEventType is when payment is requested for an order (pending).
	PaymentRequestedEventType = "PaymentRequested"
	// PaymentCreatedEventType is the event type for payment created events.
	PaymentCreatedEventType = "PaymentCreated"
	// PaymentProcessingEventType is the event type for payment processing events.
	PaymentProcessingEventType = "PaymentProcessing"
	// PaymentCompletedEventType is the event type for payment completed events.
	PaymentCompletedEventType = "PaymentCompleted"
	// PaymentFailedEventType is the event type for payment failure events.
	PaymentFailedEventType = "PaymentFailed"
	// PaymentRefundedEventType is the event type for payment refund events.
	PaymentRefundedEventType = "PaymentRefunded"
)

const (
	// PaymentOrderEventsConsumerGroup is the consumer group for order events.
	PaymentOrderEventsConsumerGroup = "payment-service.order-events" // For order lifecycle
	// PaymentEventsConsumerGroup is the consumer group for payment request events.
	PaymentEventsConsumerGroup = "payment-service.payment-events" // For payment requests
)
