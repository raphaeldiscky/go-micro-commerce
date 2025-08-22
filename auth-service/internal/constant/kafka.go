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
	TopicUserVerification = "user.verification" // EmailVerificationRequested, EmailVerified
	// TopicUserVerificationNumPartitions is the number of partitions for the user verification topic.
	TopicUserVerificationNumPartitions = 3
	// TopicUserVerificationReplicationFactor is the replication factor for the user verification topic.
	TopicUserVerificationReplicationFactor = 1
	// TopicUserSecurity is the topic for user security events.
	TopicUserSecurity = "user.security" // PasswordResetRequested, PasswordChanged
)
