package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// PaymentLifecycleEvent is the envelope for payment lifecycle events.
type PaymentLifecycleEvent struct {
	Metadata event.Metadata                `json:"metadata"`
	Payload  event.PaymentLifecyclePayload `json:"payload"`
}

// PaymentLifecycleConsumer handles the logic for processing payment lifecycle events.
type PaymentLifecycleConsumer struct {
	logger        logger.Logger
	datastore     repository.DataStore
	paymentClient client.PaymentClientInterface
}

// NewPaymentLifecycleConsumer creates a new consumer for payment lifecycle events.
func NewPaymentLifecycleConsumer(
	appLogger logger.Logger,
	ds repository.DataStore,
	paymentClient client.PaymentClientInterface,
) *PaymentLifecycleConsumer {
	return &PaymentLifecycleConsumer{
		logger:        appLogger,
		datastore:     ds,
		paymentClient: paymentClient,
	}
}

// Handler is the method that implements mq.KafkaHandler. It contains the business logic.
func (c *PaymentLifecycleConsumer) Handler(ctx context.Context, body []byte) error {
	// First, extract metadata to understand the event
	var meta struct {
		Metadata event.Metadata `json:"metadata"`
	}

	if err := json.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	// Store event in inbox for idempotent processing
	inboxEvent := entity.NewInboxEvent(
		meta.Metadata.EventID,
		"payment", // aggregate type
		meta.Metadata.AggregateID,
		meta.Metadata.EventType,
		kafka.PaymentLifecycleTopic, // topic
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
		case kafka.PaymentCreatedEventType:
			processingErr = c.processPaymentCreated(ctx, ds, body)
		case kafka.PaymentProcessingEventType:
			processingErr = c.processPaymentProcessing(ctx, ds, body)
		case kafka.PaymentCompletedEventType:
			processingErr = c.processPaymentCompleted(ctx, ds, body)
		case kafka.PaymentFailedEventType:
			processingErr = c.processPaymentFailed(ctx, ds, body)
		case kafka.PaymentRefundedEventType:
			processingErr = c.processPaymentRefunded(ctx, ds, body)
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

// processPaymentCreated handles payment created events.
func (c *PaymentLifecycleConsumer) processPaymentCreated(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var evt PaymentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal payment created event: %w", err)
	}

	c.logger.Infof("Handling payment created event for order ID: %s", evt.Payload.OrderID)

	// Update order status to indicate payment is being processed
	orderRepo := ds.OrderRepository()

	order, err := orderRepo.FindByID(ctx, evt.Payload.OrderID)
	if err != nil {
		if err.Error() == constant.OrderNotFoundErrorMessage {
			return fmt.Errorf("order not found for payment event: %s", evt.Payload.OrderID)
		}

		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		c.logger.Warnf("Order not found for payment created event: %s", evt.Payload.OrderID)

		return nil
	}

	// Update order status to processing if it's still pending
	if order.Status == constant.OrderStatusPending {
		order.Status = constant.OrderStatusProcessing
		if _, err = orderRepo.Update(ctx, order); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		c.logger.Infof("Order %s status updated to processing", evt.Payload.OrderID)
	}

	// Notify waiting saga about payment creation
	if c.paymentClient != nil {
		response := &dto.PaymentResponse{
			PaymentID: evt.Payload.PaymentID,
			Status:    evt.Payload.Status,
			OrderID:   evt.Payload.OrderID,
			Error:     nil,
		}
		c.paymentClient.NotifyWaitingSaga(response)
	}

	return nil
}

// processPaymentProcessing handles payment processing events.
func (c *PaymentLifecycleConsumer) processPaymentProcessing(
	_ context.Context,
	_ repository.DataStore,
	body []byte,
) error {
	var evt PaymentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal payment processing event: %w", err)
	}

	c.logger.Infof("Handling payment processing event for order ID: %s", evt.Payload.OrderID)

	// Order is already in processing state, no action needed
	return nil
}

// processPaymentCompleted handles payment completed events.
func (c *PaymentLifecycleConsumer) processPaymentCompleted(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var evt PaymentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal payment completed event: %w", err)
	}

	c.logger.Infof("Handling payment completed event for order ID: %s", evt.Payload.OrderID)

	// Update order status to paid
	orderRepo := ds.OrderRepository()

	order, err := orderRepo.FindByID(ctx, evt.Payload.OrderID)
	if err != nil {
		if err.Error() == constant.OrderNotFoundErrorMessage {
			return fmt.Errorf("order not found for payment event: %s", evt.Payload.OrderID)
		}

		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		c.logger.Warnf("Order not found for payment completed event: %s", evt.Payload.OrderID)

		return nil
	}

	// Update order status to paid
	order.Status = constant.OrderStatusPaid
	if _, err = orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	c.logger.Infof("Order %s status updated to paid", evt.Payload.OrderID)

	// Notify waiting saga about payment completion
	if c.paymentClient != nil {
		response := &dto.PaymentResponse{
			PaymentID: evt.Payload.PaymentID,
			Status:    "completed",
			OrderID:   evt.Payload.OrderID,
			Error:     nil,
		}
		c.paymentClient.NotifyWaitingSaga(response)
	}

	return nil
}

// processPaymentFailed handles payment failed events.
func (c *PaymentLifecycleConsumer) processPaymentFailed(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var evt PaymentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal payment failed event: %w", err)
	}

	c.logger.Infof("Handling payment failed event for order ID: %s", evt.Payload.OrderID)

	// Update order status to failed
	orderRepo := ds.OrderRepository()

	order, err := orderRepo.FindByID(ctx, evt.Payload.OrderID)
	if err != nil {
		if err.Error() == constant.OrderNotFoundErrorMessage {
			return fmt.Errorf("order not found for payment event: %s", evt.Payload.OrderID)
		}

		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		c.logger.Warnf("Order not found for payment failed event: %s", evt.Payload.OrderID)

		return nil
	}

	// Update order status to failed
	order.Status = constant.OrderStatusFailed
	if _, err = orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	c.logger.Infof("Order %s status updated to failed", evt.Payload.OrderID)

	// Notify waiting saga about payment failure
	if c.paymentClient != nil {
		response := &dto.PaymentResponse{
			PaymentID: evt.Payload.PaymentID,
			Status:    "failed",
			OrderID:   evt.Payload.OrderID,
			Error:     errors.New("payment failed"),
		}
		c.paymentClient.NotifyWaitingSaga(response)
	}

	return nil
}

// processPaymentRefunded handles payment refunded events.
func (c *PaymentLifecycleConsumer) processPaymentRefunded(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var evt PaymentLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal payment refunded event: %w", err)
	}

	c.logger.Infof("Handling payment refunded event for order ID: %s", evt.Payload.OrderID)

	// Update order status to canceled (refunded implies cancellation)
	orderRepo := ds.OrderRepository()

	order, err := orderRepo.FindByID(ctx, evt.Payload.OrderID)
	if err != nil {
		if err.Error() == constant.OrderNotFoundErrorMessage {
			return fmt.Errorf("order not found for payment event: %s", evt.Payload.OrderID)
		}

		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		c.logger.Warnf("Order not found for payment refunded event: %s", evt.Payload.OrderID)

		return nil
	}

	// Update order status to canceled
	order.Status = constant.OrderStatusCanceled
	if _, err = orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	c.logger.Infof("Order %s status updated to canceled due to refund", evt.Payload.OrderID)

	return nil
}
