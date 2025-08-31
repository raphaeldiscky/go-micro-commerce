package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/repository"
)

// OrderItemPayload holds the data for each item in the order.
type OrderItemPayload struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}

// OrderLifecyclePayload holds the data for the Order Lifecycle event.
type OrderLifecyclePayload struct {
	OrderID    uuid.UUID            `json:"order_id"`
	UserID     uuid.UUID            `json:"user_id"`
	Status     constant.OrderStatus `json:"status"`
	TotalPrice decimal.Decimal      `json:"total_price"`
	Items      []OrderItemPayload   `json:"items"`
}

// OrderPaymentRequestPayload holds the data for payment request events.
type OrderPaymentRequestPayload struct {
	OrderID       uuid.UUID       `json:"order_id"`
	CustomerID    uuid.UUID       `json:"customer_id"`
	TotalPrice    decimal.Decimal `json:"total_price"`
	Currency      string          `json:"currency"`
	PaymentMethod string          `json:"payment_method"`
}

// OrderLifecycleEvent is the envelope for all Order events.
type OrderLifecycleEvent struct {
	Metadata mq.KafkaMetadata      `json:"metadata"`
	Payload  OrderLifecyclePayload `json:"payload"`
}

// OrderPaymentRequestEvent is the envelope for payment request events.
type OrderPaymentRequestEvent struct {
	Metadata mq.KafkaMetadata           `json:"metadata"`
	Payload  OrderPaymentRequestPayload `json:"payload"`
}

// OrderLifecycleConsumer handles the logic for processing product created events.
type OrderLifecycleConsumer struct {
	logger    logger.Logger
	datastore repository.DataStore
}

// NewOrderLifecycleConsumer creates a new consumer for product lifecycle events.
func NewOrderLifecycleConsumer(
	appLogger logger.Logger,
	ds repository.DataStore,
) *OrderLifecycleConsumer {
	return &OrderLifecycleConsumer{
		logger:    appLogger,
		datastore: ds,
	}
}

// Handler is the method that implements mq.KafkaHandler. It contains the business logic.
func (c *OrderLifecycleConsumer) Handler(ctx context.Context, body []byte) error {
	// First, extract metadata to understand the event
	var meta struct {
		Metadata mq.KafkaMetadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	// Store event in inbox for idempotent processing
	inboxEvent := entity.NewInboxEvent(
		meta.Metadata.EventID,
		"order", // aggregate type
		meta.Metadata.AggregateID,
		meta.Metadata.EventType,
		"order.lifecycle", // topic
		"order-service",   // source service
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
		case constant.KafkaEventTypeOrderCreated:
			processingErr = c.processCreatedOrder(ctx, ds, body)
		case constant.KafkaEventTypeOrderUpdated:
			processingErr = c.processUpdatedOrder(ctx, ds, body)
		case constant.KafkaEventTypeOrderDeleted:
			processingErr = c.processDeletedOrder(ctx, ds, body)
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

// processCreatedOrder handles order created events to create payment records.
func (c *OrderLifecycleConsumer) processCreatedOrder(
	_ context.Context,
	_ repository.DataStore,
	body []byte,
) error {
	var event OrderLifecycleEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order created event: %w", err)
	}

	c.logger.Infof("Handling order created event for order ID: %s", event.Payload.OrderID)

	// For order created events, we don't automatically create payments
	// Payments are created when payment is requested
	// This consumer can be used for other order lifecycle tracking if needed
	c.logger.Infof(
		"Order %s created, payment will be created when payment is requested",
		event.Payload.OrderID,
	)

	return nil
}

// processUpdatedOrder handles order status updates.
func (c *OrderLifecycleConsumer) processUpdatedOrder(
	_ context.Context,
	_ repository.DataStore,
	body []byte,
) error {
	var event OrderLifecycleEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order updated event: %w", err)
	}

	c.logger.Infof("Handling order updated event for order ID: %s, status: %s",
		event.Payload.OrderID, event.Payload.Status)

	// Handle specific order status changes that might affect payments
	switch event.Payload.Status {
	case constant.OrderStatusCanceled:
		// If order is canceled, we might want to refund any completed payments
		c.logger.Infof("Order %s canceled, checking for payments to refund", event.Payload.OrderID)
		// Refund logic can be implemented here
	case constant.OrderStatusPaid:
		// Order marked as paid (from external payment processing)
		c.logger.Infof("Order %s marked as paid externally", event.Payload.OrderID)
	case constant.OrderStatusPending, constant.OrderStatusShipped, constant.OrderStatusDelivered:
		// no action needed
	default:
		c.logger.Infof("No payment action needed for order %s status: %s",
			event.Payload.OrderID, event.Payload.Status)
	}

	return nil
}

// processDeletedOrder handles order deletion events.
func (c *OrderLifecycleConsumer) processDeletedOrder(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var event OrderLifecycleEvent

	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order deleted event: %w", err)
	}

	c.logger.Infof("Handling order deleted event for order ID: %s", event.Payload.OrderID)

	// When an order is deleted, we should handle any related payments
	paymentRepo := ds.PaymentRepository()

	// Find payment for this order
	payment, err := paymentRepo.FindByOrderID(ctx, event.Payload.OrderID)
	if err != nil {
		return fmt.Errorf("failed to find payment for deleted order: %w", err)
	}

	if payment != nil {
		c.logger.Infof("Found payment %s for deleted order %s, current status: %s",
			payment.ID, event.Payload.OrderID, payment.Status)

		// Handle payment based on current status
		switch payment.Status {
		case constant.PaymentStatusCompleted:
			// Need to refund completed payments
			c.logger.Warnf(
				"Order deleted but payment %s is completed - refund needed",
				payment.ID,
			)
			// Refund logic would go here
		case constant.PaymentStatusPending, constant.PaymentStatusProcessing:
			// Cancel pending/processing payments
			c.logger.Infof("Canceling payment %s due to order deletion", payment.ID)

			if err := payment.UpdateStatus(constant.PaymentStatusFailed); err != nil {
				return fmt.Errorf("failed to cancel payment: %w", err)
			}

			if _, err := paymentRepo.Update(ctx, payment); err != nil {
				return fmt.Errorf("failed to update canceled payment: %w", err)
			}
		case constant.PaymentStatusFailed, constant.PaymentStatusRefunded:
			// No action needed
		}
	} else {
		c.logger.Infof("No payment found for deleted order %s", event.Payload.OrderID)
	}

	return nil
}
