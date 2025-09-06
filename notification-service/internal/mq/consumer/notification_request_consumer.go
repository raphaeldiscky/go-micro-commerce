package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/repository"
)

// NotificationRequestEvent is the envelope for notification request events.
type NotificationRequestEvent struct {
	Metadata event.Metadata                   `json:"metadata"`
	Payload  event.NotificationRequestPayload `json:"payload"`
}

// NotificationRequestConsumer handles storing notification request events in inbox for exactly-once processing.
type NotificationRequestConsumer struct {
	dataStore repository.DataStore
	logger    logger.Logger
}

// NewNotificationRequestConsumer creates a new consumer for notification request events.
func NewNotificationRequestConsumer(
	dataStore repository.DataStore,
	appLogger logger.Logger,
) *NotificationRequestConsumer {
	return &NotificationRequestConsumer{
		dataStore: dataStore,
		logger:    appLogger,
	}
}

// Handler processes notification request events by storing them in inbox for exactly-once processing.
func (c *NotificationRequestConsumer) Handler(ctx context.Context, body []byte) error {
	var meta struct {
		Metadata event.Metadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	c.logger.Infof("Received notification request event: %s from %s",
		meta.Metadata.EventID, meta.Metadata.Source)

	inboxEvent := entity.NewInboxEvent(
		meta.Metadata.EventID,
		"notification", // aggregate type
		meta.Metadata.AggregateID,
		meta.Metadata.EventType,
		kafka.NotificationRequestTopic,
		meta.Metadata.Source,
		json.RawMessage(body),
		nil, // correlation ID - not available in current metadata
		nil, // causation ID - not available in current metadata
	)
	inboxRepo := c.dataStore.InboxRepository()

	_, err := inboxRepo.Create(ctx, inboxEvent)
	if err != nil {
		return fmt.Errorf("failed to store event in inbox: %w", err)
	}

	c.logger.Infof("Successfully stored event %s in inbox for processing", meta.Metadata.EventID)

	return nil
}
