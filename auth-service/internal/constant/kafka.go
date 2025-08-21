// Package constant defines constants used in the auth service.
package constant

// Auth Service Source.
const (
	KafkaSourceAuthService = "auth-service"
)

// Auth Service Event Types.
const (
	// KafkaEventTypeEmailVerificationRequested is the event type for email verification requested events.
	KafkaEventTypeEmailVerificationRequested = "EmailVerificationRequested"
	// KafkaEventTypeUserVerified is the event type for user verified events.
	KafkaEventTypeUserVerified = "UserVerified"
)

// Topics that Auth Service produces to.
const (
	UserVerificationTopic = "user.verification" // EmailVerificationRequested, EmailVerified
	// UserVerificationTopicNumPartitions is the number of partitions for the user verification topic.
	UserVerificationTopicNumPartitions = 3
	// UserVerificationTopicReplicationFactor is the replication factor for the user verification topic.
	UserVerificationTopicReplicationFactor = 1
	// UserSecurityTopic is the topic for user security events.
	UserSecurityTopic = "user.security" // PasswordResetRequested, PasswordChanged
)
