package event

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// ProductLifecycleConsumer handles the logic for processing product created events.
type ProductLifecycleConsumer struct {
	logger    logger.Logger
	datastore repository.DataStore
}

// NewProductLifecycleConsumer creates a new consumer for product lifecycle events.
func NewProductLifecycleConsumer(
	appLogger logger.Logger,
	ds repository.DataStore,
) *ProductLifecycleConsumer {
	return &ProductLifecycleConsumer{
		logger:    appLogger,
		datastore: ds,
	}
}

// Handler is the method that implements mq.KafkaHandler. It contains the business logic.
func (c *ProductLifecycleConsumer) Handler(ctx context.Context, body []byte) error {
	var meta struct {
		Metadata mq.KafkaMetadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	switch meta.Metadata.EventType {
	case constant.KafkaEventTypeProductCreated:
		return c.handleCreatedProduct(ctx, body)
	case constant.KafkaEventTypeProductUpdated:
		return c.handleUpdatedProduct(ctx, body)
	case constant.KafkaEventTypeProductDeleted:
		return c.handleDeletedProduct(ctx, body)
	default:
		c.logger.Warnf("ignoring event type: %s", meta.Metadata.EventType)

		return nil
	}
}

// handleCreatedProduct.
func (c *ProductLifecycleConsumer) handleCreatedProduct(ctx context.Context, body []byte) error {
	var event ProductCreatedEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal product created event: %w", err)
	}

	c.logger.Infof("Handling product created event for product ID: %s", event.Payload.ProductID)

	// Add your business logic here for handling the created product event.
	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		product := &entity.Product{
			ID:        event.Payload.ProductID,
			Name:      event.Payload.Name,
			Price:     event.Payload.Price,
			Quantity:  event.Payload.Quantity,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		_, err := ds.ProductRepository().Create(ctx, product)
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}

		return nil
	})
}

// handleUpdatedProduct.
func (c *ProductLifecycleConsumer) handleUpdatedProduct(ctx context.Context, body []byte) error {
	var event ProductUpdatedEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal product updated event: %w", err)
	}

	c.logger.Infof("Handling product updated event for product ID: %s", event.Payload.ProductID)

	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		product, err := ds.ProductRepository().FindByID(ctx, event.Payload.ProductID)
		if err != nil {
			return fmt.Errorf("failed to find product: %w", err)
		}

		// Update the product fields
		product.Name = event.Payload.Name
		product.Price = event.Payload.Price
		product.Quantity = event.Payload.Quantity
		product.UpdatedAt = time.Now().UTC()

		_, err = ds.ProductRepository().Update(ctx, product)
		if err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		return nil
	})
}

// handleDeletedProduct.
func (c *ProductLifecycleConsumer) handleDeletedProduct(ctx context.Context, body []byte) error {
	var event ProductDeletedEvent

	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal product deleted event: %w", err)
	}

	c.logger.Infof("Handling product deleted event for product ID: %s", event.Payload.ProductID)

	// Add your business logic here for handling the deleted product event.
	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		if err := ds.ProductRepository().Delete(ctx, event.Payload.ProductID); err != nil {
			return fmt.Errorf("failed to delete product: %w", err)
		}

		return nil
	})
}
