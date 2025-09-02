package event

const (
	// UserVerificationTopic is the topic for user verification events.
	UserVerificationTopic = "user.verification" // EmailVerificationRequested, EmailVerified
	// UserSecurityTopic is the topic for user security events.
	UserSecurityTopic = "user.security" // PasswordResetRequested, PasswordChanged
)

const (
	// EmailVerificationRequestedEventType is the event type for email verification requested events.
	EmailVerificationRequestedEventType = "EmailVerificationRequested"
	// UserVerifiedEventType is the event type for user verified events.
	UserVerifiedEventType = "UserVerified"
)
