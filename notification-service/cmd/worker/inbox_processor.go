package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/worker"
)

// InboxProcessorWorker wraps the Kafka consumer server as a Worker.
type InboxProcessorWorker struct {
	consumer *worker.InboxProcessor
}

// NewInboxProcessorWorker creates a new Kafka consumer worker.
func NewInboxProcessorWorker(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *InboxProcessorWorker {
	return &InboxProcessorWorker{
		consumer: provider.SetupInboxProcessor(cfg, appLogger, providers),
	}
}

// Name returns the name of the worker.
func (w *InboxProcessorWorker) Name() string {
	return "Inbox Processor"
}

// Start starts the Kafka consumer server.
func (w *InboxProcessorWorker) Start(ctx context.Context) error {
	// Start server in goroutine
	errChan := make(chan error, 1)

	go func() {
		if err := w.consumer.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		return nil // Context canceled, normal shutdown
	case err := <-errChan:
		return err // Server error
	}
}

// Shutdown gracefully shuts down the Kafka consumer worker.
func (w *InboxProcessorWorker) Shutdown(ctx context.Context) error {
	return w.consumer.Shutdown(ctx)
}
