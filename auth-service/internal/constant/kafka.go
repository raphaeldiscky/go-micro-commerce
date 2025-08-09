// Package constant defines constants used in the auth service.
package constant

// AuthTopics defines the topics used by the auth service for event publishing.
type AuthTopics struct {
	UserLifecycle      string
	UserAuthentication string
	UserVerification   string
	UserSecurity       string
}

// NewAuthTopics initializes and returns a AuthTopics instance with predefined topics.
func NewAuthTopics() AuthTopics {
	return AuthTopics{
		UserVerification: TopicUserVerification,
		UserSecurity:     TopicUserSecurity,
	}
}

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
	// TopicUserSecurity is the topic for user security events.
	TopicUserSecurity = "user.security" // PasswordResetRequested, PasswordChanged
)
