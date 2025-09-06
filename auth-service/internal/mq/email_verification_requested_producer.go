// Package mq provides the event definitions and handlers for the auth service.
package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// EmailVerificationRequestedEvent is the envelope for all email verification requested events.
type EmailVerificationRequestedEvent struct {
	Metadata event.Metadata                          `json:"metadata"`
	Payload  event.EmailVerificationRequestedPayload `json:"payload"`
}

// GetPayload returns the data associated with the EmailVerificationRequestedEvent.
func (e *EmailVerificationRequestedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the EmailVerificationRequestedEvent.
func (e *EmailVerificationRequestedEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// NewEmailVerificationRequestedEvent creates a new EmailVerificationRequestedEvent.
func NewEmailVerificationRequestedEvent(
	userID uuid.UUID,
	email string,
	token string,
	tokenExpiresAt time.Time,
) *EmailVerificationRequestedEvent {
	return &EmailVerificationRequestedEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.EmailVerificationRequestedEventType,
			AggregateID: userID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.AuthServiceName,
		},
		Payload: event.EmailVerificationRequestedPayload{
			UserID:         userID,
			Email:          email,
			Token:          token,
			TokenExpiresAt: tokenExpiresAt,
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
		topic:    kafka.UserVerificationTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *EmailVerificationRequestedProducer) Send(
	ctx context.Context,
	evt event.BaseEvent,
) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *EmailVerificationRequestedProducer) Topic() string {
	return p.topic
}
