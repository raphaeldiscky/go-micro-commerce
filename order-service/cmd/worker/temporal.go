package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"go.temporal.io/sdk/worker"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/provider"
)

// TemporalWorker wraps the temporal worker.
type TemporalWorker struct {
	logger         logger.Logger
	temporalClient *client.TemporalClient
}

// NewTemporalWorker creates a new Temporal worker.
func NewTemporalWorker(
	appLogger logger.Logger,
	providers *provider.Providers,
) *TemporalWorker {
	return &TemporalWorker{
		temporalClient: providers.TemporalClient,
		logger:         appLogger,
	}
}

// Name returns the worker name.
func (w *TemporalWorker) Name() string {
	return "Temporal Worker"
}

// Start starts the temporal worker.
func (w *TemporalWorker) Start(_ context.Context) error {
	w.logger.Info("Starting Temporal worker")

	if w.temporalClient == nil {
		w.logger.Warn("Temporal client not available, skipping worker start")

		return nil
	}

	// Start listening to the Task Queue
	err := w.temporalClient.Worker.Run(worker.InterruptCh())
	if err != nil {
		w.logger.Errorf("Failed to start Temporal worker: %v", err)

		return err
	}

	return nil
}

// Shutdown gracefully shuts down the temporal worker.
func (w *TemporalWorker) Shutdown(_ context.Context) error {
	w.logger.Info("Shutting down Temporal worker")

	if w.temporalClient != nil {
		// Close the temporal client
		w.temporalClient.Close()
	}

	w.logger.Info("Temporal worker shutdown completed")

	return nil
}
