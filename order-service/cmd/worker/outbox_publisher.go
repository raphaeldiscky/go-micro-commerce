package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/provider"
)

func runOutboxPublisherWorker(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) {
	// Initialize outbox publisher
	outboxPublisher := provider.SetupOutboxPublisher(cfg, appLogger, providers)

	appLogger.Info("Starting outbox publisher worker")

	// Start the outbox publisher
	outboxPublisher.Start(ctx)

	appLogger.Info("Outbox publisher worker stopped")
}
