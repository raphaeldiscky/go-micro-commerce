// Package consumer provides the event definitions and handlers for the notification service.
package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/repository"
)

// UserVerifiedEvent is the envelope for all user verified events.
type UserVerifiedEvent struct {
	Metadata kafkaevent.Metadata            `json:"metadata"`
	Payload  kafkaevent.UserVerifiedPayload `json:"payload"`
}

// EmailVerificationRequestedEvent is the envelope for all email verification requested events.
type EmailVerificationRequestedEvent struct {
	Metadata kafkaevent.Metadata                          `json:"metadata"`
	Payload  kafkaevent.EmailVerificationRequestedPayload `json:"payload"`
}

// UserVerificationConsumer handles storing user verification events in inbox for exactly-once processing.
type UserVerificationConsumer struct {
	dataStore repository.DataStore
	logger    logger.Logger
}

// NewUserVerificationConsumer creates a new consumer for user verification requested events.
func NewUserVerificationConsumer(
	dataStore repository.DataStore,
	appLogger logger.Logger,
) *UserVerificationConsumer {
	return &UserVerificationConsumer{
		dataStore: dataStore,
		logger:    appLogger,
	}
}

// Handler processes user verification events by storing them in inbox for exactly-once processing.
func (c *UserVerificationConsumer) Handler(ctx context.Context, body []byte) error {
	var meta struct {
		Metadata kafkaevent.Metadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	c.logger.Infof("Received user verification event: %s of type %s from %s",
		meta.Metadata.EventID, meta.Metadata.EventType, meta.Metadata.Source)

	inboxEvent := entity.NewInboxEvent(
		meta.Metadata.EventID,
		"user", // aggregate type
		meta.Metadata.AggregateID,
		meta.Metadata.EventType,
		kafka.UserVerificationTopic,
		meta.Metadata.Source,
		json.RawMessage(body),
		nil, // correlation ID - not available in current metadata
		nil, // causation ID - not available in current metadata
	)

	inboxRepo := c.dataStore.InboxRepository()

	_, err := inboxRepo.Create(ctx, inboxEvent)
	if err != nil {
		return fmt.Errorf("failed to store user verification event in inbox: %w", err)
	}

	c.logger.Infof(
		"Successfully stored user verification event %s in inbox for processing",
		meta.Metadata.EventID,
	)

	return nil
}
