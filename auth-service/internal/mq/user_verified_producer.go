package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// UserVerifiedEvent is the envelope for all user verified events.
type UserVerifiedEvent struct {
	Metadata event.Metadata            `json:"metadata"`
	Payload  event.UserVerifiedPayload `json:"payload"`
}

// NewUserVerifiedEvent creates a new UserVerifiedEvent.
func NewUserVerifiedEvent(
	userID uuid.UUID,
	email string,
) *UserVerifiedEvent {
	return &UserVerifiedEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.UserVerifiedEventType,
			AggregateID: userID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.AuthServiceName,
		},
		Payload: event.UserVerifiedPayload{
			UserID: userID,
			Email:  email,
		},
	}
}

// GetPayload returns the data associated with the UserVerifiedEvent.
func (e *UserVerifiedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the UserVerifiedEvent.
func (e *UserVerifiedEvent) GetMetadata() event.Metadata {
	return e.Metadata
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
		topic:    kafka.UserVerificationTopic,
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
