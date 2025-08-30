package event

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/repository"
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

// OrderLifecycleConsumer handles the logic for processing order lifecycle events.
type OrderLifecycleConsumer struct {
	logger    logger.Logger
	datastore repository.DataStore
}

// NewOrderLifecycleConsumer creates a new consumer for order lifecycle events.
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
	case constant.KafkaEventTypeOrderCanceled:
		return c.handleCanceledOrder(ctx, body)
	default:
		c.logger.Warnf("ignoring event type: %s", meta.Metadata.EventType)

		return nil
	}
}

// handleCreatedOrder.
func (c *OrderLifecycleConsumer) handleCreatedOrder(ctx context.Context, body []byte) error {
	var event OrderLifecycleEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order created event: %w", err)
	}

	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		productRepo := ds.ProductRepository()

		productIDs := []uuid.UUID{}
		for _, item := range event.Payload.Items {
			productIDs = append(productIDs, item.ProductID)
		}

		products, err := productRepo.FindByIDsForUpdate(ctx, productIDs)
		if err != nil {
			return err
		}

		if len(products) != len(productIDs) {
			return fmt.Errorf("not all products found for update")
		}

		c.logger.Infof("reducing quantities for products: %v", productIDs)

		for i, product := range products {
			product.Quantity -= event.Payload.Items[i].Quantity
		}

		if err := productRepo.BulkUpdateQuantity(ctx, products); err != nil {
			return err
		}

		return nil
	})
}

// handleCanceledOrder.
func (c *OrderLifecycleConsumer) handleCanceledOrder(ctx context.Context, body []byte) error {
	var event OrderLifecycleEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order created event: %w", err)
	}

	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		productRepo := ds.ProductRepository()

		productIDs := []uuid.UUID{}
		for _, item := range event.Payload.Items {
			productIDs = append(productIDs, item.ProductID)
		}

		products, err := productRepo.FindByIDsForUpdate(ctx, productIDs)
		if err != nil {
			return err
		}

		if len(products) != len(productIDs) {
			return fmt.Errorf("not all products found for update")
		}

		c.logger.Infof("adding quantities for products: %v", productIDs)

		for i, product := range products {
			product.Quantity += event.Payload.Items[i].Quantity
		}

		if err := productRepo.BulkUpdateQuantity(ctx, products); err != nil {
			return err
		}

		return nil
	})
}
