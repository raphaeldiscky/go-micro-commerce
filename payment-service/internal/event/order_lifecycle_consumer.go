package event

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/repository"
)

// OrderItemPayload holds the data for each item in the order.
type OrderItemPayload struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// OrderLifecyclePayload holds the data for the Order Lifecycle event.
type OrderLifecyclePayload struct {
	OrderID    uuid.UUID            `json:"order_id"`
	UserID     uuid.UUID            `json:"user_id"`
	Status     constant.OrderStatus `json:"status"`
	TotalPrice decimal.Decimal      `json:"total_price"`
	Items      []OrderItemPayload   `json:"items"`
}

// OrderLifecycleEvent is the envelope for all Order events.
type OrderLifecycleEvent struct {
	Metadata mq.KafkaMetadata      `json:"metadata"`
	Payload  OrderLifecyclePayload `json:"payload"`
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
	var meta struct {
		Metadata mq.KafkaMetadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	switch meta.Metadata.EventType {
	case constant.KafkaEventTypeOrderCreated:
		return c.handleCreatedOrder(ctx, body)
	case constant.KafkaEventTypeOrderUpdated:
		return c.handleUpdatedOrder(ctx, body)
	case constant.KafkaEventTypeOrderDeleted:
		return c.handleDeletedOrder(ctx, body)
	default:
		c.logger.Warnf("ignoring event type: %s", meta.Metadata.EventType)

		return nil
	}
}

// handleCreatedOrder.
func (c *OrderLifecycleConsumer) handleCreatedOrder(ctx context.Context, body []byte) error {
	var event OrderLifecycleEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal product created event: %w", err)
	}

	c.logger.Infof("Handling product created event for product ID: %s", event.Payload.OrderID)

	return c.datastore.Atomic(ctx, func(_ repository.DataStore) error {
		return nil
	})
}

// handleUpdatedOrder.
func (c *OrderLifecycleConsumer) handleUpdatedOrder(ctx context.Context, body []byte) error {
	var event OrderLifecycleEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal product updated event: %w", err)
	}

	c.logger.Infof("Handling product updated event for product ID: %s", event.Payload.OrderID)

	return c.datastore.Atomic(ctx, func(_ repository.DataStore) error {
		return nil
	})
}

// handleDeletedOrder.
func (c *OrderLifecycleConsumer) handleDeletedOrder(ctx context.Context, body []byte) error {
	var event OrderLifecycleEvent

	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal product deleted event: %w", err)
	}

	c.logger.Infof("Handling product deleted event for product ID: %s", event.Payload.OrderID)

	// Add your business logic here for handling the deleted product event.
	return c.datastore.Atomic(ctx, func(_ repository.DataStore) error {
		return nil
	})
}
