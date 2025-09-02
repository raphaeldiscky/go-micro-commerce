package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event/payload"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
)

// UserVerifiedEvent is the envelope for all user verified events.
type UserVerifiedEvent struct {
	Metadata event.Metadata              `json:"metadata"`
	Payload  payload.UserVerifiedPayload `json:"payload"`
}

// GetPayload returns the data associated with the UserVerifiedEvent.
func (e *UserVerifiedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the UserVerifiedEvent.
func (e *UserVerifiedEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// NewUserVerifiedEvent creates a new UserVerifiedEvent.
func NewUserVerifiedEvent(
	userID uuid.UUID,
	email string,
) *UserVerifiedEvent {
	return &UserVerifiedEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeUserVerified,
			AggregateID: userID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceAuthService,
		},
		Payload: payload.UserVerifiedPayload{
			UserID: userID,
			Email:  email,
		},
	}
}

// UserVerifiedProducer is responsible for producing product created events.
type UserVerifiedProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewUserVerifiedProducer creates a new instance of UserVerifiedProducer.
func NewUserVerifiedProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &UserVerifiedProducer{
		Producer: producer,
		topic:    constant.TopicUserVerification,
	}
}

// Send implements the KafkaProducer interface.
func (p *UserVerifiedProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *UserVerifiedProducer) Topic() string {
	return p.topic
}
