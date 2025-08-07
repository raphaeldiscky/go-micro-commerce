// Package event provides event publishing implementation.
package event

import (
	"encoding/json"
	"log"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/config"
)

// SimplePublisher is a simple event publisher implementation.
type SimplePublisher struct {
	config *config.EventPublisherConfig
}

// NewSimplePublisher creates a new simple event publisher.
func NewSimplePublisher(cfg *config.EventPublisherConfig) PublisherInterface {
	return &SimplePublisher{
		config: cfg,
	}
}

// PublishUserRegistered publishes a user registered event.
func (p *SimplePublisher) PublishUserRegistered(event *UserRegisteredEvent) error {
	return p.publishEvent(event)
}

// PublishEmailVerificationRequested publishes an email verification requested event.
func (p *SimplePublisher) PublishEmailVerificationRequested(
	event *EmailVerificationRequestedEvent,
) error {
	return p.publishEvent(event)
}

// PublishEmailVerified publishes an email verified event.
func (p *SimplePublisher) PublishEmailVerified(event *EmailVerifiedEvent) error {
	return p.publishEvent(event)
}

// PublishPasswordResetRequested publishes a password reset requested event.
func (p *SimplePublisher) PublishPasswordResetRequested(event *PasswordResetRequestedEvent) error {
	return p.publishEvent(event)
}

// PublishPasswordChanged publishes a password changed event.
func (p *SimplePublisher) PublishPasswordChanged(event *PasswordChangedEvent) error {
	return p.publishEvent(event)
}

// PublishUserLoggedIn publishes a user logged in event.
func (p *SimplePublisher) PublishUserLoggedIn(event *UserLoggedInEvent) error {
	return p.publishEvent(event)
}

// PublishUserLoggedOut publishes a user logged out event.
func (p *SimplePublisher) PublishUserLoggedOut(event *UserLoggedOutEvent) error {
	return p.publishEvent(event)
}

// publishEvent is a generic event publisher (for now just logs, can be extended to use Kafka/RabbitMQ).
func (p *SimplePublisher) publishEvent(event interface{}) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// For now, just log the event. In production, this would publish to Kafka/RabbitMQ
	log.Printf("Publishing event to topic %s: %s", p.config.Topic, string(eventJSON))

	return nil
}
