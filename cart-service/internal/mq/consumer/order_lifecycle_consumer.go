// Package consumer provides the event definitions and handlers for the cart service.
package consumer

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/repository"
)

// OrderLifecycleEvent is the envelope for all Order events.
type OrderLifecycleEvent struct {
	Metadata kafkaevent.Metadata              `json:"metadata"`
	Payload  kafkaevent.OrderLifecyclePayload `json:"payload"`
}

// OrderLifecycleConsumer handles the logic for processing order lifecycle events.
type OrderLifecycleConsumer struct {
	logger    logger.Logger
	dataStore repository.DataStore
}

// NewOrderLifecycleConsumer creates a new consumer for order lifecycle events.
func NewOrderLifecycleConsumer(
	appLogger logger.Logger,
	ds repository.DataStore,
) *OrderLifecycleConsumer {
	return &OrderLifecycleConsumer{
		logger:    appLogger,
		dataStore: ds,
	}
}

// Handler is the method that implements mq.KafkaHandler. It contains the business logic.
func (c *OrderLifecycleConsumer) Handler(ctx context.Context, body []byte) error {
	// First, extract metadata to understand the event
	var meta struct {
		Metadata kafkaevent.Metadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	c.logger.Infof(
		"Received order lifecycle event: type=%s, id=%s",
		meta.Metadata.EventType,
		meta.Metadata.EventID,
	)

	// Process the event based on type
	switch meta.Metadata.EventType {
	case kafka.OrderCreatedEventType:
		return c.processCreatedOrder(ctx, body)
	default:
		c.logger.Infof("Ignoring event type: %s", meta.Metadata.EventType)
		return nil
	}
}

// processCreatedOrder handles order created events to update checkout session status.
func (c *OrderLifecycleConsumer) processCreatedOrder(
	ctx context.Context,
	body []byte,
) error {
	var evt OrderLifecycleEvent
	if err := sonic.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal order created event: %w", err)
	}

	c.logger.Infof(
		"Processing order created event: order_id=%s, checkout_session_id=%s",
		evt.Payload.OrderID,
		evt.Payload.CheckoutSessionID,
	)

	// Update checkout session status to order_placed
	return c.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		checkoutSessionRepo := ds.CheckoutSessionRepository()

		// Get checkout session
		session, err := checkoutSessionRepo.GetByID(ctx, evt.Payload.CheckoutSessionID)
		if err != nil {
			c.logger.Errorf("Failed to get checkout session: %v", err)
			return fmt.Errorf("failed to get checkout session: %w", err)
		}

		if session == nil {
			c.logger.Warnf("Checkout session not found: %s", evt.Payload.CheckoutSessionID)
			// Don't return error - this is not a fatal condition
			return nil
		}

		// Update status to order_placed (idempotent operation)
		if err = session.UpdateStatus(constant.CheckoutSessionStatusOrderPlaced); err != nil {
			c.logger.Errorf("Failed to update checkout session status: %v", err)
			return fmt.Errorf("failed to update checkout session status: %w", err)
		}

		// Save updated session
		_, err = checkoutSessionRepo.Update(ctx, session)
		if err != nil {
			c.logger.Errorf("Failed to save checkout session: %v", err)
			return fmt.Errorf("failed to save checkout session: %w", err)
		}

		c.logger.Infof(
			"Successfully updated checkout session %s to order_placed",
			evt.Payload.CheckoutSessionID,
		)

		return nil
	})
}
