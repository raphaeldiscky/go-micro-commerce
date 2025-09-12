// Package mq provides the event definitions and handlers for the fulfillment service.
package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/repository"
)

// FulfillmentRequestEvent is the envelope for fulfillment request events.
type FulfillmentRequestEvent struct {
	Metadata event.Metadata                  `json:"metadata"`
	Payload  event.FulfillmentRequestPayload `json:"payload"`
}

// GetMetadata returns the metadata associated with the FulfillmentRequestEvent.
func (e *FulfillmentRequestEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// GetPayload returns the payload associated with the FulfillmentRequestEvent.
func (e *FulfillmentRequestEvent) GetPayload() any {
	return e.Payload
}

// FulfillmentRequestConsumer handles the logic for processing fulfillment request events.
type FulfillmentRequestConsumer struct {
	logger        logger.Logger
	datastore     repository.DataStore
	carrierClient client.CarrierClientInterface
}

// NewFulfillmentRequestConsumer creates a new consumer for fulfillment request events.
func NewFulfillmentRequestConsumer(
	appLogger logger.Logger,
	ds repository.DataStore,
	carrierClient client.CarrierClientInterface,
) *FulfillmentRequestConsumer {
	return &FulfillmentRequestConsumer{
		logger:        appLogger,
		datastore:     ds,
		carrierClient: carrierClient,
	}
}

// Handler is the method that implements mq.KafkaHandler. It contains the business logic.
func (c *FulfillmentRequestConsumer) Handler(ctx context.Context, body []byte) error {
	c.logger.Infof("Received fulfillment request event: %s", string(body))
	// First, extract metadata to understand the event
	var meta struct {
		Metadata event.Metadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	// Store event in inbox for idempotent processing
	inboxEvent := entity.NewInboxEvent(
		meta.Metadata.EventID,
		"fulfillment", // aggregate type
		meta.Metadata.AggregateID,
		meta.Metadata.EventType,
		kafka.FulfillmentRequestTopic, // topic
		meta.Metadata.Source,
		json.RawMessage(body),
		nil, // correlation_id - could be extracted from metadata if available
		nil, // causation_id - could be extracted from metadata if available
	)

	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		inboxRepo := ds.InboxRepository()

		// Store event in inbox (handles duplicates automatically)
		storedEvent, err := inboxRepo.Create(ctx, inboxEvent)
		if err != nil {
			return fmt.Errorf("failed to store event in inbox: %w", err)
		}

		// If it's a duplicate, just log and return successfully
		if storedEvent.Status == constant.InboxStatusDuplicate {
			c.logger.Infof(
				"Duplicate event received: %s, skipping processing",
				meta.Metadata.EventID,
			)

			return nil
		}

		// Mark as processing
		if err = inboxRepo.MarkAsProcessing(ctx, storedEvent.ID); err != nil {
			return fmt.Errorf("failed to mark event as processing: %w", err)
		}

		// Process the event based on type
		var processingErr error

		switch meta.Metadata.EventType {
		case kafka.FulfillmentRequestedEventType:
			processingErr = c.processFulfillmentRequested(ctx, ds, body)
		default:
			c.logger.Warnf("ignoring unknown event type: %s", meta.Metadata.EventType)
			// Mark as processed even for unknown events to avoid reprocessing
			processingErr = nil
		}

		// Update inbox event status based on processing result
		if processingErr != nil {
			c.logger.Errorf("Failed to process event %s: %v", meta.Metadata.EventID, processingErr)

			if err = inboxRepo.MarkAsFailed(ctx, storedEvent.ID, processingErr.Error()); err != nil {
				return fmt.Errorf("failed to mark event as failed: %w", err)
			}

			return processingErr
		}

		if err = inboxRepo.MarkAsProcessed(ctx, storedEvent.ID); err != nil {
			return fmt.Errorf("failed to mark event as processed: %w", err)
		}

		return nil
	})
}

// processFulfillmentRequested handles fulfillment requested events to create fulfillment records.
func (c *FulfillmentRequestConsumer) processFulfillmentRequested(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var evt FulfillmentRequestEvent
	if err := sonic.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal fulfillment request event: %w", err)
	}

	c.logger.Infof("Processing fulfillment request for order ID: %s", evt.Payload.OrderID)

	// Check if fulfillment already exists
	if exists, err := c.checkExistingFulfillment(ctx, ds, evt.Payload.OrderID); err != nil {
		return err
	} else if exists {
		return nil
	}

	// Calculate shipping and create fulfillment
	fulfillment, err := c.createFulfillmentFromEvent(ctx, &evt)
	if err != nil {
		return fmt.Errorf("failed to create fulfillment from event: %w", err)
	}

	// Save fulfillment and publish event
	return c.saveFulfillmentAndPublishEvent(ctx, ds, fulfillment, evt.Payload.OrderID)
}

// checkExistingFulfillment checks if a fulfillment already exists for the given order ID.
func (c *FulfillmentRequestConsumer) checkExistingFulfillment(
	ctx context.Context,
	ds repository.DataStore,
	orderID uuid.UUID,
) (bool, error) {
	fulfillmentRepo := ds.FulfillmentRepository()

	existingFulfillment, err := fulfillmentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return false, fmt.Errorf("failed to check existing fulfillment: %w", err)
	}

	if existingFulfillment != nil {
		c.logger.Infof("Fulfillment already exists for order %s, skipping creation", orderID)

		return true, nil
	}

	return false, nil
}

// createFulfillmentFromEvent creates a fulfillment entity from the event payload.
func (c *FulfillmentRequestConsumer) createFulfillmentFromEvent(
	ctx context.Context,
	evt *FulfillmentRequestEvent,
) (*entity.Fulfillment, error) {
	// Mock for now
	toAddress := entity.ToAddress{
		City:       evt.Payload.Shipping.ToAddress.City,
		State:      evt.Payload.Shipping.ToAddress.State,
		PostalCode: evt.Payload.Shipping.ToAddress.PostalCode,
		Country:    evt.Payload.Shipping.ToAddress.Country,
	}
	fromAddress := entity.FromAddress{
		City:       evt.Payload.Shipping.FromAddress.City,
		State:      evt.Payload.Shipping.FromAddress.State,
		PostalCode: evt.Payload.Shipping.FromAddress.PostalCode,
		Country:    evt.Payload.Shipping.FromAddress.Country,
	}

	dimensions := entity.Dimensions{
		Length: evt.Payload.Shipping.Dimensions.Length,
		Height: evt.Payload.Shipping.Dimensions.Height,
		Width:  evt.Payload.Shipping.Dimensions.Width,
		Unit:   evt.Payload.Shipping.Dimensions.Unit,
	}

	weightKG := evt.Payload.Shipping.WeightKG

	// Create shipping request
	shippingRequest := &dto.ShippingRequest{
		OrderID:     evt.Payload.OrderID,
		CarrierID:   constant.CarrierID(evt.Payload.Shipping.CarrierID),
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		WeightKG:    weightKG,
		Dimensions:  dimensions,
	}

	rate, err := c.carrierClient.GetRate(ctx, shippingRequest)
	if err != nil {
		c.logger.Warnf("Failed to get shipping rates: %v, using default values", err)
	}

	shippingRequest.CarrierID = rate.CarrierID
	shippingCost := rate.ShippingCost
	estimatedDelivery := rate.EstimatedDelivery

	label, err := c.carrierClient.CreateShipment(ctx, shippingRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create shipping label: %w", err)
	}

	trackingNumber := label.TrackingNumber

	// Create fulfillment entity
	fulfillment, err := entity.NewFulfillment(evt.Payload.OrderID,
		trackingNumber,
		evt.Payload.Currency,
		shippingCost,
		weightKG,
		fromAddress,
		toAddress,
		estimatedDelivery,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create fulfillment entity: %w", err)
	}

	return fulfillment, nil
}

// saveFulfillmentAndPublishEvent saves the fulfillment to database and publishes the created event.
func (c *FulfillmentRequestConsumer) saveFulfillmentAndPublishEvent(
	ctx context.Context,
	ds repository.DataStore,
	fulfillment *entity.Fulfillment,
	orderID uuid.UUID,
) error {
	fulfillmentRepo := ds.FulfillmentRepository()

	// Save to database
	createdFulfillment, err := fulfillmentRepo.Create(ctx, fulfillment)
	if err != nil {
		return fmt.Errorf("failed to create fulfillment record: %w", err)
	}

	c.logger.Infof(
		"Successfully created fulfillment %s for order %s with tracking number %s",
		createdFulfillment.ID,
		orderID,
		createdFulfillment.TrackingNumber,
	)

	// Publish fulfillment created event
	return c.publishFulfillmentCreatedEvent(ctx, ds, createdFulfillment)
}

// publishFulfillmentCreatedEvent publishes the fulfillment created event to notify order service.
func (c *FulfillmentRequestConsumer) publishFulfillmentCreatedEvent(
	ctx context.Context,
	ds repository.DataStore,
	fulfillment *entity.Fulfillment,
) error {
	outboxRepo := ds.OutboxRepository()

	// Create fulfillment created event
	fulfillmentCreatedEvent := NewFulfillmentLifecycleEvent(fulfillment)

	payload, err := json.Marshal(fulfillmentCreatedEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal fulfillment created event: %w", err)
	}

	outboxEvent := &entity.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "fulfillment",
		AggregateID:   fulfillment.ID,
		EventType:     kafka.FulfillmentCreatedEventType,
		Topic:         kafka.FulfillmentLifecycleTopic,
		Payload:       payload,
		Status:        constant.OutboxStatusPending,
		CreatedAt:     time.Now().UTC(),
		ScheduledFor:  time.Now().UTC(),
		Attempts:      0,
	}

	if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
		return fmt.Errorf("failed to create fulfillment created outbox event: %w", err)
	}

	c.logger.Infof("Fulfillment created event published for order %s", fulfillment.OrderID)

	return nil
}
