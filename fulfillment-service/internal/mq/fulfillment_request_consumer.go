// Package mq provides the event definitions and handlers for the fulfillment service.
package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/shopspring/decimal"

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
func (e *FulfillmentRequestEvent) GetPayload() interface{} {
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
		"order-service",               // source service
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
		if err := inboxRepo.MarkAsProcessing(ctx, storedEvent.ID); err != nil {
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

			if err := inboxRepo.MarkAsFailed(ctx, storedEvent.ID, processingErr.Error()); err != nil {
				return fmt.Errorf("failed to mark event as failed: %w", err)
			}

			return processingErr
		}

		if err := inboxRepo.MarkAsProcessed(ctx, storedEvent.ID); err != nil {
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

	fulfillmentRepo := ds.FulfillmentRepository()

	// Check if fulfillment already exists for this order
	existingFulfillment, err := fulfillmentRepo.FindByOrderID(ctx, evt.Payload.OrderID)
	if err != nil {
		return fmt.Errorf("failed to check existing fulfillment: %w", err)
	}

	if existingFulfillment != nil {
		c.logger.Infof(
			"Fulfillment already exists for order %s, skipping creation",
			evt.Payload.OrderID,
		)

		return nil
	}

	// Step 1: Convert shipping address from payload to DTO
	toAddress := c.convertShippingAddress(&evt.Payload.ShippingAddress)

	// Step 2: Create shipping request for rate calculation
	// Estimate package weight based on items (simplified logic)
	totalWeight := c.estimatePackageWeight(evt.Payload.Items)

	shippingRequest := &dto.ShippingRequest{
		OrderID: evt.Payload.OrderID,
		Carrier: string(constant.CarrierTypeJNE), // Default carrier for rate checking
		Service: "JNE Regular",
		FromAddress: dto.ShippingAddress{
			Name:       "Fulfillment Center",
			Company:    "E-Commerce Platform",
			Address1:   "Jl. Fulfillment Center No. 1",
			City:       "Jakarta",
			State:      "DKI Jakarta",
			PostalCode: "12345",
			Country:    "Indonesia",
			Phone:      "+62-21-12345678",
		},
		ToAddress: toAddress,
		Package: dto.Package{
			Weight: totalWeight,
			Dimensions: map[string]decimal.Decimal{
				"width":  decimal.NewFromInt(20), // Default dimensions in cm
				"height": decimal.NewFromInt(15),
				"length": decimal.NewFromInt(30),
			},
		},
		InsuranceAmount: evt.Payload.TotalPrice,
		Signature: evt.Payload.TotalPrice.GreaterThan(
			decimal.NewFromInt(1000000),
		), // Require signature for high-value items
	}

	// Step 3: Get shipping rates from carrier
	rates, err := c.carrierClient.GetRates(ctx, shippingRequest)
	if err != nil {
		c.logger.Warnf("Failed to get shipping rates: %v, using default values", err)
		// Continue with default values if carrier service is unavailable
	}

	// Step 4: Select best rate (for simplicity, use the first available rate or default)
	var selectedRate *dto.ShippingRate
	if len(rates) > 0 {
		selectedRate = &rates[0] // Use first rate for simplicity
	}

	// Step 5: Create shipping label
	var shippingLabel *dto.ShippingLabel

	var shippingCost decimal.Decimal

	var estimatedDelivery time.Time

	if selectedRate != nil {
		shippingRequest.Carrier = string(selectedRate.Carrier)
		shippingRequest.Service = selectedRate.Service
		shippingCost = selectedRate.Cost
		estimatedDelivery = selectedRate.EstimatedDelivery

		label, err := c.carrierClient.CreateShipment(ctx, shippingRequest)
		if err != nil {
			c.logger.Errorf("Failed to create shipping label: %v", err)

			return fmt.Errorf("failed to create shipping label: %w", err)
		}

		shippingLabel = label
	} else {
		// Use default values if carrier integration fails
		shippingCost = decimal.NewFromInt(25000)           // Default shipping cost
		estimatedDelivery = time.Now().Add(72 * time.Hour) // Default 3 days
	}

	// Step 6: Create fulfillment record
	trackingNumber := ""
	if shippingLabel != nil {
		trackingNumber = shippingLabel.TrackingNumber
	}

	// Create fulfillment using the constructor with proper parameters
	fulfillment, err := entity.NewFulfillment(
		evt.Payload.OrderID,
		trackingNumber,
		shippingCost,
		totalWeight,
		estimatedDelivery,
	)
	if err != nil {
		return fmt.Errorf("failed to create fulfillment entity: %w", err)
	}

	// Set additional fields not handled by constructor
	fulfillment.Currency = evt.Payload.Currency
	if shippingLabel != nil {
		fulfillment.Carrier = &shippingLabel.Carrier
		fulfillment.ShippingLabelURL = &shippingLabel.LabelURL
	}

	// Step 7: Save to database
	createdFulfillment, err := fulfillmentRepo.Create(ctx, fulfillment)
	if err != nil {
		return fmt.Errorf("failed to create fulfillment record: %w", err)
	}

	c.logger.Infof(
		"Successfully created fulfillment %s for order %s with tracking number %s",
		createdFulfillment.ID,
		evt.Payload.OrderID,
		trackingNumber,
	)

	// Step 8: Publish fulfillment created event (if needed)
	// TODO: Publish FulfillmentCreated event to notify order service

	return nil
}

// convertShippingAddress converts event payload address to DTO address.
func (c *FulfillmentRequestConsumer) convertShippingAddress(
	addr *event.ShippingAddressPayload,
) dto.ShippingAddress {
	return dto.ShippingAddress{
		Name:       "Customer", // Default name, could be enhanced with customer data
		Address1:   addr.Street,
		City:       addr.City,
		State:      addr.State,
		PostalCode: addr.PostalCode,
		Country:    addr.Country,
	}
}

// estimatePackageWeight estimates the total weight based on items.
func (c *FulfillmentRequestConsumer) estimatePackageWeight(
	items []event.FulfillmentItemPayload,
) decimal.Decimal {
	// Simple estimation: assume each item weighs 0.5kg on average
	totalItems := int64(0)
	for _, item := range items {
		totalItems += item.Quantity
	}

	// Minimum weight of 0.1kg, plus 0.5kg per item
	baseWeight := decimal.NewFromFloat(0.1)
	itemWeight := decimal.NewFromInt(totalItems).Mul(decimal.NewFromFloat(0.5))

	return baseWeight.Add(itemWeight)
}
