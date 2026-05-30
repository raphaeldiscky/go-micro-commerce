package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/worker"
)

// outboxPublisherRunner wraps the outbox publisher as a Runner.
type outboxPublisherRunner struct {
	publisher *worker.OutboxPublisher
	logger    logger.Logger
	cancel    context.CancelFunc
}

// newOutboxPublisherRunner creates a new outbox publisher runner.
func newOutboxPublisherRunner(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *outboxPublisherRunner {
	return &outboxPublisherRunner{
		publisher: provider.SetupOutboxPublisher(ctx, cfg, appLogger, providers),
		logger:    appLogger,
	}
}

// Name returns the name of the runner.
func (r *outboxPublisherRunner) Name() string {
	return "Outbox Publisher"
}

// Start starts the outbox publisher.
func (r *outboxPublisherRunner) Start(ctx context.Context) error {
	publisherCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	r.publisher.Start(publisherCtx)

	<-ctx.Done()

	return nil
}

// Shutdown gracefully shuts down the outbox publisher.
func (r *outboxPublisherRunner) Shutdown(_ context.Context) error {
	r.logger.Info("Stopping outbox publisher...")

	if r.cancel != nil {
		r.cancel()
	}

	r.logger.Info("Outbox publisher stopped successfully")

	return nil
}
