package constant

// Auth Service Source.
const (
	KafkaSourceAuthService = "auth-service"
)

// Auth Service Event Types.
const (
	// KafkaEventTypeUserRegistered is the event type for user registration events.
	KafkaEventTypeUserRegistered = "UserRegistered"
	// KafkaEventTypeEmailVerificationRequested is the event type for email verification requested events.
	KafkaEventTypeEmailVerificationRequested = "EmailVerificationRequested"
	// KafkaEventTypeEmailVerified is the event type for email verified events.
	KafkaEventTypeEmailVerified = "EmailVerified"
	// KafkaEventTypePasswordResetRequested is the event type for password reset requested events.
	KafkaEventTypePasswordResetRequested = "PasswordResetRequested"
	// KafkaEventTypePasswordChanged is the event type for password changed events.
	KafkaEventTypePasswordChanged = "PasswordChanged"
	// KafkaEventTypeUserLoggedIn is the event type for user logged in events.
	KafkaEventTypeUserLoggedIn = "UserLoggedIn"
	// KafkaEventTypeUserLoggedOut is the event type for user logged out events.
	KafkaEventTypeUserLoggedOut = "UserLoggedOut"
)

// Topics that Auth Service produces to.
const (
	// TopicUserLifecycle is the topic for user lifecycle events.
	TopicUserLifecycle = "user.lifecycle" // UserRegistered, UserDeactivated, UserProfileUpdated
	// TopicUserAuthentication is the topic for user authentication events.
	TopicUserAuthentication = "user.authentication" // UserLoggedIn, UserLoggedOut
	// TopicUserVerification is the topic for user verification events.
	TopicUserVerification = "user.verification" // EmailVerificationRequested, EmailVerified
	// TopicUserSecurity is the topic for user security events.
	TopicUserSecurity = "user.security" // PasswordResetRequested, PasswordChanged
)
