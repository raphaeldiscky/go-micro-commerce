package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event/payload"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// ProductCreatedEvent is the envelope for all product events.
type ProductCreatedEvent struct {
	Metadata event.Metadata                `json:"metadata"`
	Payload  payload.ProductCreatedPayload `json:"payload"`
}

// ProductUpdatedEvent is the envelope for all product events.
type ProductUpdatedEvent struct {
	Metadata event.Metadata                `json:"metadata"`
	Payload  payload.ProductUpdatedPayload `json:"payload"`
}

// ProductDeletedEvent is the envelope for all product events.
type ProductDeletedEvent struct {
	Metadata event.Metadata                `json:"metadata"`
	Payload  payload.ProductDeletedPayload `json:"payload"`
}

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
		Metadata event.Metadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	switch meta.Metadata.EventType {
	case event.ProductCreatedEventType:
		return c.handleCreatedProduct(ctx, body)
	case event.ProductUpdatedEventType:
		return c.handleUpdatedProduct(ctx, body)
	case event.ProductDeletedEventType:
		return c.handleDeletedProduct(ctx, body)
	default:
		c.logger.Warnf("ignoring event type: %s", meta.Metadata.EventType)

		return nil
	}
}

// handleCreatedProduct.
func (c *ProductLifecycleConsumer) handleCreatedProduct(ctx context.Context, body []byte) error {
	var evt ProductCreatedEvent
	if err := sonic.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal product created event: %w", err)
	}

	c.logger.Infof("Handling product created event for product ID: %s", evt.Payload.ProductID)

	// Add your business logic here for handling the created product event.
	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		product := &entity.Product{
			ID:        evt.Payload.ProductID,
			Name:      evt.Payload.Name,
			Price:     evt.Payload.Price,
			Quantity:  evt.Payload.Quantity,
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
	var evt ProductUpdatedEvent
	if err := sonic.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal product updated event: %w", err)
	}

	c.logger.Infof("Handling product updated event for product ID: %s", evt.Payload.ProductID)

	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		product, err := ds.ProductRepository().FindByID(ctx, evt.Payload.ProductID)
		if err != nil {
			return fmt.Errorf("failed to find product: %w", err)
		}

		// Update the product fields
		product.Name = evt.Payload.Name
		product.Price = evt.Payload.Price
		product.Quantity = evt.Payload.Quantity
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
	var evt ProductDeletedEvent

	if err := sonic.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal product deleted event: %w", err)
	}

	c.logger.Infof("Handling product deleted event for product ID: %s", evt.Payload.ProductID)

	// Add your business logic here for handling the deleted product event.
	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		if err := ds.ProductRepository().Delete(ctx, evt.Payload.ProductID); err != nil {
			return fmt.Errorf("failed to delete product: %w", err)
		}

		return nil
	})
}
