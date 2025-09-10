// Package consumer provides event consumers for the search service.
package consumer

import (
	"context"
	"encoding/json"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/repository"
)

// ProductCreatedEvent is the envelope for product created events.
type ProductCreatedEvent struct {
	Metadata event.Metadata              `json:"metadata"`
	Payload  event.ProductCreatedPayload `json:"payload"`
}

// ProductUpdatedEvent is the envelope for product updated events.
type ProductUpdatedEvent struct {
	Metadata event.Metadata              `json:"metadata"`
	Payload  event.ProductUpdatedPayload `json:"payload"`
}

// ProductDeletedEvent is the envelope for product deleted events.
type ProductDeletedEvent struct {
	Metadata event.Metadata              `json:"metadata"`
	Payload  event.ProductDeletedPayload `json:"payload"`
}

// SearchEventsConsumer handles storing search-related events in inbox for exactly-once processing.
type SearchEventsConsumer struct {
	dataStore repository.DataStore
	logger    logger.Logger
}

// NewSearchEventsConsumer creates a new search events consumer.
func NewSearchEventsConsumer(
	dataStore repository.DataStore,
	appLogger logger.Logger,
) *SearchEventsConsumer {
	return &SearchEventsConsumer{
		dataStore: dataStore,
		logger:    appLogger,
	}
}

// Handler returns the Kafka handler function that processes search events using the inbox pattern.
func (c *SearchEventsConsumer) Handler(ctx context.Context, body []byte) error {
	return c.storeEventInInbox(ctx, body)
}

// storeEventInInbox stores an event in the inbox for reliable processing.
func (c *SearchEventsConsumer) storeEventInInbox(ctx context.Context, body []byte) error {
	var genericEvent event.GenericEvent
	if err := json.Unmarshal(body, &genericEvent); err != nil {
		c.logger.Errorf("Failed to unmarshal generic event: %v", err)

		return err
	}

	metadata := genericEvent.Metadata

	// Extract aggregate information based on event type
	var aggregateType string

	var sourceService string

	var topic string

	switch metadata.EventType {
	case "ProductCreated", "ProductUpdated", "ProductDeleted":
		aggregateType = "product"
		sourceService = "product-service"
		topic = kafka.ProductLifecycleTopic
	default:
		c.logger.Warnf("Unknown event type: %s", metadata.EventType)

		return nil
	}

	// Create inbox event with available fields
	inboxEvent := entity.NewInboxEvent(
		metadata.EventID, // Use EventID as MessageID
		aggregateType,
		metadata.AggregateID, // Use AggregateID from metadata
		metadata.EventType,
		topic,
		sourceService,
		json.RawMessage(body),
		nil, // CorrelationID not available in current event structure
		nil, // CausationID not available in current event structure
	)

	// Store in inbox - repository will handle duplicates
	_, err := c.dataStore.InboxRepository().Create(ctx, inboxEvent)
	if err != nil {
		c.logger.Errorf("Failed to store event in inbox: %v", err)

		return err
	}

	c.logger.Infof("Successfully stored %s event in inbox: %s", metadata.EventType, inboxEvent.ID)

	return nil
}
