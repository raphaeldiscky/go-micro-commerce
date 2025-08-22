// Package constant defines constants used in the order service for Kafka topics and event types.
package constant

// Notification Service Source.
const (
	KafkaSourceNotificationService = "notification-service"
)

// Auth Service Event Types.
const (
	// KafkaEventTypeEmailVerificationRequested is the event type for email verification requested events.
	KafkaEventTypeEmailVerificationRequested = "EmailVerificationRequested"
	// KafkaEventTypeUserVerified is the event type for user verified events.
	KafkaEventTypeUserVerified = "UserVerified"
)

// Topics that Notification Service produces to.
const (
	// TopicUserVerification is the topic for user verification events.
	TopicUserVerification = "user.verification" // EmailVerificationRequested, UserVerified
)

// Consumer groups for Notification Service (consuming from other services).
const (
	// ConsumerGroupNotificationUserEvents is the consumer group for user events.
	ConsumerGroupNotificationUserEvents = "notification-service.user-events" // For user lifecycle
	// ConsumerGroupNotificationOrderEvents is the consumer group for order events.
	ConsumerGroupNotificationOrderEvents = "notification-service.order-events" // For order lifecycle
)
