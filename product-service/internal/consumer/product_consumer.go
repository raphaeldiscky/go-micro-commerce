// Package consumer provides Kafka consumer implementations for the product service.
package consumer

import (
	"context"
	"log"

	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/event"
)

// ProductEventConsumer handles product-related events.
type ProductEventConsumer struct {
	// Add any dependencies you need (e.g., repository, logger)
}

// NewProductEventConsumer creates a new ProductEventConsumer.
func NewProductEventConsumer() *ProductEventConsumer {
	return &ProductEventConsumer{}
}

// HandleProductCreated handles ProductCreated events.
func (c *ProductEventConsumer) HandleProductCreated(evt event.ProductCreatedEvent, headers map[string]string) error {
	log.Printf("Handling ProductCreated event for product %s: %s",
		evt.Payload.ProductID, evt.Payload.Name)

	// Add your business logic here
	// e.g., update search index, send notifications, etc.

	return nil
}

// HandleProductUpdated handles ProductUpdated events.
func (c *ProductEventConsumer) HandleProductUpdated(evt event.ProductUpdatedEvent, headers map[string]string) error {
	log.Printf("Handling ProductUpdated event for product %s: %s",
		evt.Payload.ProductID, evt.Payload.Name)

	// Add your business logic here

	return nil
}

// HandleProductDeleted handles ProductDeleted events.
func (c *ProductEventConsumer) HandleProductDeleted(evt event.ProductDeletedEvent, headers map[string]string) error {
	log.Printf("Handling ProductDeleted event for product %s", evt.Payload.ProductID)

	// Add your business logic here

	return nil
}

// SetupConsumer sets up the Kafka consumer with event handlers.
func (c *ProductEventConsumer) SetupConsumer(cfg *mq.KafkaConsumerConfig, topics constant.ProductTopics) (*mq.ConsumerKafka, error) {
	// Create multi-event handler
	multiHandler := mq.NewMultiEventHandler()

	// Register typed handlers
	multiHandler.RegisterHandler(
		constant.KafkaEventTypeProductCreated,
		mq.CreateTypedHandler(c.HandleProductCreated),
	)

	multiHandler.RegisterHandler(
		constant.KafkaEventTypeProductUpdated,
		mq.CreateTypedHandler(c.HandleProductUpdated),
	)

	multiHandler.RegisterHandler(
		constant.KafkaEventTypeProductDeleted,
		mq.CreateTypedHandler(c.HandleProductDeleted),
	)

	// Create consumer
	return mq.NewConsumerKafka(
		cfg,
		"product-service-consumer", // consumer group ID
		topics.ProductLifecycle,    // topic
		multiHandler.Handle,        // handler function
	)
}

// Example of how to start the consumer
func StartProductConsumer(ctx context.Context, cfg *mq.KafkaConsumerConfig, topics constant.ProductTopics) error {
	consumer := NewProductEventConsumer()

	kafkaConsumer, err := consumer.SetupConsumer(cfg, topics)
	if err != nil {
		return err
	}
	defer kafkaConsumer.Close()

	log.Println("Starting product event consumer...")
	return kafkaConsumer.Start(ctx)
}
