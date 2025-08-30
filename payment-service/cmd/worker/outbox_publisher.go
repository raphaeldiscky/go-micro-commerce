package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// OutboxPublisherWorker wraps the outbox publisher as a Worker.
type OutboxPublisherWorker struct {
	publisher *service.OutboxPublisher
	logger    logger.Logger
	cancel    context.CancelFunc
}

// NewOutboxPublisherWorker creates a new outbox publisher worker.
func NewOutboxPublisherWorker(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *OutboxPublisherWorker {
	return &OutboxPublisherWorker{
		publisher: provider.SetupOutboxPublisher(cfg, appLogger, providers),
		logger:    appLogger,
	}
}

// Name returns the name of the worker.
func (w *OutboxPublisherWorker) Name() string {
	return "Outbox Publisher"
}

// Start starts the outbox publisher.
func (w *OutboxPublisherWorker) Start(ctx context.Context) error {
	// Create cancellable context for the publisher
	publisherCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	// Start publisher
	w.publisher.Start(publisherCtx)

	// Wait for context cancellation
	<-ctx.Done()

	return nil
}

// Shutdown gracefully shuts down the outbox publisher.
func (w *OutboxPublisherWorker) Shutdown(_ context.Context) error {
	w.logger.Info("Stopping outbox publisher...")

	// Cancel the publisher context to stop processing loops
	if w.cancel != nil {
		w.cancel()
	}

	w.logger.Info("Outbox publisher stopped successfully")

	return nil
}
