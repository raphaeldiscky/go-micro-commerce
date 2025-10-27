package provider

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/worker"
)

// SetupOutboxPublisher initializes the outbox publisher service.
func SetupOutboxPublisher(
	_ context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *worker.OutboxPublisher {
	registry := kafka.NewEventRegistry()

	registry.Register(
		kafka.NotificationRequestedEventType,
		&producer.NotificationRequestEvent{},
	)

	outboxPublisher := worker.NewOutboxPublisher(
		providers.DataStore,
		appLogger,
		providers.NotificationRequestProducer,
		*cfg.OutboxPublisher,
		registry,
	)

	return outboxPublisher
}
