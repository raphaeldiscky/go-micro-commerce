package event

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/constant"
)

// EmailVerificationRequestedPayload holds the data for the email verification requested event.
type EmailVerificationRequestedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

// EmailVerificationRequestedEvent is the envelope for all email verification requested events.
type EmailVerificationRequestedEvent struct {
	metadata KafkaMetadata
	payload  EmailVerificationRequestedPayload
}

// GetPayload returns the data associated with the EmailVerificationRequestedEvent.
func (e *EmailVerificationRequestedEvent) GetPayload() interface{} {
	return e.payload
}

// GetMetadata returns the metadata associated with the EmailVerificationRequestedEvent.
func (e *EmailVerificationRequestedEvent) GetMetadata() KafkaMetadata {
	return e.metadata
}

// NewEmailVerificationRequestedEvent creates a new EmailVerificationRequestedEvent.
func NewEmailVerificationRequestedEvent(
	userID uuid.UUID,
	email string,
) *EmailVerificationRequestedEvent {
	return &EmailVerificationRequestedEvent{
		metadata: KafkaMetadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeEmailVerificationRequested,
			AggregateID: userID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceAuthService,
		},
		payload: EmailVerificationRequestedPayload{
			UserID: userID,
			Email:  email,
		},
	}
}

// EmailVerificationRequestedProducer is responsible for producing product created events.
type EmailVerificationRequestedProducer struct {
	Producer  *mq.KafkaAsyncProducer
	RetryChan chan *sarama.ProducerMessage
	topic     string
}

// NewEmailVerificationRequestedProducer creates a new instance of EmailVerificationRequestedProducer.
func NewEmailVerificationRequestedProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducer {
	pr := &EmailVerificationRequestedProducer{
		Producer:  producer,
		topic:     constant.UserVerificationTopic,
		RetryChan: make(chan *sarama.ProducerMessage, 100),
	}

	return pr
}

// Send implements the KafkaProducer interface.
func (p *EmailVerificationRequestedProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *EmailVerificationRequestedProducer) Topic() string {
	return p.topic
}
