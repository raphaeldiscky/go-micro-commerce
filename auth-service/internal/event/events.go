// Package event defines event structures and interfaces for the auth service.
package event

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event interface.
type DomainEvent interface {
	GetEventID() uuid.UUID
	GetEventType() string
	GetAggregateID() uuid.UUID
	GetOccurredAt() time.Time
	GetData() interface{}
}

// BaseEvent represents a base event structure.
type BaseEvent struct {
	EventID   uuid.UUID `json:"event_id"`
	EventType string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	UserID    uuid.UUID `json:"user_id"`
}

// UserRegisteredEvent represents a user registration event.
type UserRegisteredEvent struct {
	BaseEvent
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// EmailVerificationRequestedEvent represents an email verification request event.
type EmailVerificationRequestedEvent struct {
	BaseEvent
	Email             string `json:"email"`
	VerificationToken string `json:"verification_token"`
}

// EmailVerifiedEvent represents an email verification event.
type EmailVerifiedEvent struct {
	BaseEvent
	Email string `json:"email"`
}

// PasswordResetRequestedEvent represents a password reset request event.
type PasswordResetRequestedEvent struct {
	BaseEvent
	Email      string `json:"email"`
	ResetToken string `json:"reset_token"`
}

// PasswordChangedEvent represents a password change event.
type PasswordChangedEvent struct {
	BaseEvent
	Email string `json:"email"`
}

// UserLoggedInEvent represents a user login event.
type UserLoggedInEvent struct {
	BaseEvent
	Email     string `json:"email"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

// UserLoggedOutEvent represents a user logout event.
type UserLoggedOutEvent struct {
	BaseEvent
	Email string `json:"email"`
}

// PublisherInterface defines the contract for event publishing.
type PublisherInterface interface {
	// PublishUserRegistered publishes a user registered event
	PublishUserRegistered(event *UserRegisteredEvent) error

	// PublishEmailVerificationRequested publishes an email verification requested event
	PublishEmailVerificationRequested(event *EmailVerificationRequestedEvent) error

	// PublishEmailVerified publishes an email verified event
	PublishEmailVerified(event *EmailVerifiedEvent) error

	// PublishPasswordResetRequested publishes a password reset requested event
	PublishPasswordResetRequested(event *PasswordResetRequestedEvent) error

	// PublishPasswordChanged publishes a password changed event
	PublishPasswordChanged(event *PasswordChangedEvent) error

	// PublishUserLoggedIn publishes a user logged in event
	PublishUserLoggedIn(event *UserLoggedInEvent) error

	// PublishUserLoggedOut publishes a user logged out event
	PublishUserLoggedOut(event *UserLoggedOutEvent) error
}

// NewBaseEvent creates a new base event.
func NewBaseEvent(eventType string, userID uuid.UUID) BaseEvent {
	return BaseEvent{
		EventID:   uuid.New(),
		EventType: eventType,
		Timestamp: time.Now(),
		UserID:    userID,
	}
}
