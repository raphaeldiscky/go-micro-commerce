// Package constant defines constants used in the order service for Kafka topics and event types.
package constant

// Notification Service Source.
const (
	KafkaSourceNotificationService = "notification-service"
)

// Notification Service Event Types.
const (
	// KafkaEventTypeEmailSent is the event type for email sent events.
	KafkaEventTypeEmailSent = "EmailSent"
	// KafkaEventTypeEmailFailed is the event type for email failed events.
	KafkaEventTypeEmailFailed = "EmailFailed"
)

// Topics that Notification Service produces to.
const (
	// UserVerificationTopic is the topic for user verification events.
	UserVerificationTopic = "user.verification"
)

// Consumer groups for Notification Service (consuming from other services).
const (
	// ConsumerGroupNotificationUserEvents is the consumer group for user events.
	ConsumerGroupNotificationUserEvents = "notification-service.user-events" // For user lifecycle
	// ConsumerGroupNotificationOrderEvents is the consumer group for order events.
	ConsumerGroupNotificationOrderEvents = "notification-service.order-events" // For order lifecycle
)
