// Package consumer provides the event definitions and handlers for the order service.
package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// FulfillmentLifecycleEvent is the envelope for all Fulfillment events.
type FulfillmentLifecycleEvent struct {
	Metadata kafkaevent.Metadata                    `json:"metadata"`
	Payload  kafkaevent.FulfillmentLifecyclePayload `json:"payload"`
}

// FulfillmentLifecycleConsumer handles the logic for processing fulfillment lifecycle events.
type FulfillmentLifecycleConsumer struct {
	logger            logger.Logger
	datastore         repository.DataStore
	fulfillmentClient client.FulfillmentClient
}

// NewFulfillmentLifecycleConsumer creates a new consumer for fulfillment lifecycle events.
func NewFulfillmentLifecycleConsumer(
	appLogger logger.Logger,
	ds repository.DataStore,
	fulfillmentClient client.FulfillmentClient,
) *FulfillmentLifecycleConsumer {
	return &FulfillmentLifecycleConsumer{
		logger:            appLogger,
		datastore:         ds,
		fulfillmentClient: fulfillmentClient,
	}
}

// Handler is the method that implements mq.KafkaHandler. It contains the business logic.
func (c *FulfillmentLifecycleConsumer) Handler(ctx context.Context, body []byte) error {
	// First, extract metadata to understand the event
	var meta struct {
		Metadata kafkaevent.Metadata `json:"metadata"`
	}

	if err := json.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	// Store event in inbox for idempotent processing
	inboxEvent := entity.NewInboxEvent(
		meta.Metadata.EventID,
		"fulfillment", // aggregate type
		meta.Metadata.AggregateID,
		meta.Metadata.EventType,
		kafka.FulfillmentLifecycleTopic, // topic
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
		processingErr := c.processEvent(ctx, ds, meta.Metadata.EventType, body)

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

// processEvent handles the event processing based on event type.
func (c *FulfillmentLifecycleConsumer) processEvent(
	ctx context.Context,
	ds repository.DataStore,
	eventType string,
	body []byte,
) error {
	switch eventType {
	case kafka.FulfillmentCreatedEventType:
		return c.processFulfillmentCreated(ctx, ds, body)
	case kafka.FulfillmentProcessingEventType:
		return c.processFulfillmentProcessing(ctx, ds, body)
	case kafka.FulfillmentShippedEventType:
		return c.processFulfillmentShipped(ctx, ds, body)
	case kafka.FulfillmentInTransitEventType:
		return c.processFulfillmentInTransit(ctx, ds, body)
	case kafka.FulfillmentDeliveredEventType:
		return c.processFulfillmentDelivered(ctx, ds, body)
	case kafka.FulfillmentCanceledEventType:
		return c.processFulfillmentCanceled(ctx, ds, body)
	case kafka.FulfillmentReturnedEventType:
		return c.processFulfillmentReturned(ctx, ds, body)
	default:
		c.logger.Warnf("ignoring unknown event type: %s", eventType)
		// Mark as processed even for unknown events to avoid reprocessing
		return nil
	}
}

// processFulfillmentCreated handles fulfillment created events.
func (c *FulfillmentLifecycleConsumer) processFulfillmentCreated(
	_ context.Context,
	_ repository.DataStore,
	body []byte,
) error {
	var evt FulfillmentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal fulfillment created event: %w", err)
	}

	c.logger.Infof("Handling fulfillment created event for order ID: %s", evt.Payload.OrderID)

	// Notify waiting saga with fulfillment response via client
	response := &dto.FulfillmentResponse{
		FulfillmentID:  evt.Payload.FulfillmentID,
		TrackingNumber: evt.Payload.TrackingNumber,
		ShippingCost:   evt.Payload.ShippingCost,
		Status:         evt.Payload.Status,
		OrderID:        evt.Payload.OrderID,
	}

	c.fulfillmentClient.NotifyWaitingSaga(response)

	return nil
}

// processFulfillmentProcessing handles fulfillment processing events.
func (c *FulfillmentLifecycleConsumer) processFulfillmentProcessing(
	_ context.Context,
	_ repository.DataStore,
	body []byte,
) error {
	var evt FulfillmentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal fulfillment processing event: %w", err)
	}

	c.logger.Infof("Handling fulfillment processing event for order ID: %s", evt.Payload.OrderID)

	// No order status change needed for fulfillment processing
	return nil
}

// processFulfillmentShipped handles fulfillment shipped events.
func (c *FulfillmentLifecycleConsumer) processFulfillmentShipped(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var evt FulfillmentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal fulfillment shipped event: %w", err)
	}

	c.logger.Infof("Handling fulfillment shipped event for order ID: %s", evt.Payload.OrderID)

	// Update order status to shipped
	orderRepo := ds.OrderRepository()

	order, err := orderRepo.FindByID(ctx, evt.Payload.OrderID)
	if err != nil {
		if err.Error() == constant.OrderNotFoundErrorMessage {
			return fmt.Errorf("order not found for fulfillment event: %s", evt.Payload.OrderID)
		}

		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		c.logger.Warnf("Order not found for fulfillment shipped event: %s", evt.Payload.OrderID)

		return nil
	}

	// Update order status to shipped
	order.Status = constant.OrderStatusShipped
	if _, err = orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	c.logger.Infof("Order %s status updated to shipped", evt.Payload.OrderID)

	return nil
}

// processFulfillmentInTransit handles fulfillment in transit events.
func (c *FulfillmentLifecycleConsumer) processFulfillmentInTransit(
	_ context.Context,
	_ repository.DataStore,
	body []byte,
) error {
	var evt FulfillmentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal fulfillment in transit event: %w", err)
	}

	c.logger.Infof("Handling fulfillment in transit event for order ID: %s", evt.Payload.OrderID)

	// Order remains in shipped status during transit
	return nil
}

// processFulfillmentDelivered handles fulfillment delivered events.
func (c *FulfillmentLifecycleConsumer) processFulfillmentDelivered(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var evt FulfillmentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal fulfillment delivered event: %w", err)
	}

	c.logger.Infof("Handling fulfillment delivered event for order ID: %s", evt.Payload.OrderID)

	// Update order status to delivered
	orderRepo := ds.OrderRepository()
	outboxRepo := ds.OutboxRepository()
	sagaStateRepo := ds.SagaStateRepository()

	order, err := orderRepo.FindByID(ctx, evt.Payload.OrderID)
	if err != nil {
		if err.Error() == constant.OrderNotFoundErrorMessage {
			return fmt.Errorf("order not found for fulfillment event: %s", evt.Payload.OrderID)
		}

		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		c.logger.Warnf("Order not found for fulfillment delivered event: %s", evt.Payload.OrderID)

		return nil
	}

	// Update order status to delivered
	order.Status = constant.OrderStatusDelivered
	if _, err = orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	sagaState, err := sagaStateRepo.FindByOrderID(ctx, order.ID)
	if err != nil && err.Error() != constant.SagaStateNotFoundErrorMessage {
		return fmt.Errorf("failed to get order saga: %w", err)
	}

	// If no saga state found, skip email extraction (saga might not be used for this order)
	if sagaState == nil {
		c.logger.Infof(
			"No saga state found for order %s, skipping customer email extraction",
			order.ID,
		)

		return nil
	}

	// Extract customer email from saga state
	customerEmail, ok := sagaState.Data["customer_email"].(string)
	if !ok {
		return errors.New("customer email not found in saga state")
	}

	// Extract reserved products from saga state
	reservedProductsData, ok := sagaState.Data["reserved_products"]
	if !ok {
		return errors.New("reserved products not found in saga state")
	}

	// Convert reserved products any to []entity.Product
	var reservedProducts []entity.Product

	productsBytes, err := json.Marshal(reservedProductsData)
	if err != nil {
		return fmt.Errorf("failed to marshal reserved products: %w", err)
	}

	if err = json.Unmarshal(productsBytes, &reservedProducts); err != nil {
		return fmt.Errorf("failed to unmarshal reserved products: %w", err)
	}

	// Send notification
	notificationEvent := producer.NewNotificationRequestEvent(
		order,
		reservedProducts,
		customerEmail,
		"Customer Name",
		&evt.Payload.TrackingNumber,
		pkgconstant.TemplateOrderDelivered,
		"Your Order Has Been Delivered!",
	)

	payload, err := json.Marshal(notificationEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal notification event: %w", err)
	}

	outboxEvent := &entity.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "notification",
		AggregateID:   order.ID,
		EventType:     kafka.NotificationRequestedEventType,
		Topic:         kafka.NotificationRequestTopic,
		Payload:       payload,
		Status:        constant.OutboxStatusPending,
		CreatedAt:     time.Now().UTC(),
		ScheduledFor:  time.Now().UTC(),
		Attempts:      0,
	}

	if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
		return fmt.Errorf("failed to create order delivered notification event: %w", err)
	}

	c.logger.Infof(
		"Successfully created order delivered notification for order: %s",
		order.ID,
	)

	c.logger.Infof("Order %s status updated to delivered", evt.Payload.OrderID)

	return nil
}

// processFulfillmentCanceled handles fulfillment canceled events.
func (c *FulfillmentLifecycleConsumer) processFulfillmentCanceled(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var evt FulfillmentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal fulfillment canceled event: %w", err)
	}

	c.logger.Infof("Handling fulfillment canceled event for order ID: %s", evt.Payload.OrderID)

	// Update order status to canceled
	orderRepo := ds.OrderRepository()

	order, err := orderRepo.FindByID(ctx, evt.Payload.OrderID)
	if err != nil {
		if err.Error() == constant.OrderNotFoundErrorMessage {
			return fmt.Errorf("order not found for fulfillment event: %s", evt.Payload.OrderID)
		}

		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		c.logger.Warnf("Order not found for fulfillment canceled event: %s", evt.Payload.OrderID)

		return nil
	}

	// Update order status to canceled
	order.Status = constant.OrderStatusCanceled
	if _, err = orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	c.logger.Infof("Order %s status updated to canceled", evt.Payload.OrderID)

	return nil
}

// processFulfillmentReturned handles fulfillment returned events.
func (c *FulfillmentLifecycleConsumer) processFulfillmentReturned(
	_ context.Context,
	_ repository.DataStore,
	body []byte,
) error {
	var evt FulfillmentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal fulfillment returned event: %w", err)
	}

	c.logger.Infof("Handling fulfillment returned event for order ID: %s", evt.Payload.OrderID)

	// For returns, we might want to keep it as delivered but add a return flag
	// or create a separate return status - for now, just logging
	c.logger.Infof("Order %s has been returned", evt.Payload.OrderID)

	return nil
}
