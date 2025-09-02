package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
)

// EmailVerificationRequestedPayload holds the data for the email verification requested event.
type EmailVerificationRequestedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Token  string    `json:"token"`
}

// EmailVerificationRequestedEvent is the envelope for all email verification requested events.
type EmailVerificationRequestedEvent struct {
	Metadata kafka.Metadata                    `json:"metadata"`
	Payload  EmailVerificationRequestedPayload `json:"payload"`
}

// GetPayload returns the data associated with the EmailVerificationRequestedEvent.
func (e *EmailVerificationRequestedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the EmailVerificationRequestedEvent.
func (e *EmailVerificationRequestedEvent) GetMetadata() KafkaMetadata {
	return e.Metadata
}

// NewEmailVerificationRequestedEvent creates a new EmailVerificationRequestedEvent.
func NewEmailVerificationRequestedEvent(
	userID uuid.UUID,
	email string,
	token string,
) *EmailVerificationRequestedEvent {
	return &EmailVerificationRequestedEvent{
		Metadata: KafkaMetadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeEmailVerificationRequested,
			AggregateID: userID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceAuthService,
		},
		Payload: EmailVerificationRequestedPayload{
			UserID: userID,
			Email:  email,
			Token:  token,
		},
	}
}

// EmailVerificationRequestedProducer is responsible for producing product created events.
type EmailVerificationRequestedProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewEmailVerificationRequestedProducer creates a new instance of EmailVerificationRequestedProducer.
func NewEmailVerificationRequestedProducer(
	producer *kafka.AsyncProducer,
) kafka.ProducerInterface {
	return &EmailVerificationRequestedProducer{
		Producer: producer,
		topic:    constant.TopicUserVerification,
	}
}

// Send implements the KafkaProducer interface.
func (p *EmailVerificationRequestedProducer) Send(
	ctx context.Context,
	event kafka.BaseEvent,
) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *EmailVerificationRequestedProducer) Topic() string {
	return p.topic
}
