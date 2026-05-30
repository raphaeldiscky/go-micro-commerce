package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"go.temporal.io/sdk/worker"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/provider"
)

// temporalRunner wraps the Temporal worker as a Runner.
type temporalRunner struct {
	logger         logger.Logger
	temporalClient *client.TemporalClient
}

// newTemporalRunner creates a new Temporal runner.
func newTemporalRunner(
	appLogger logger.Logger,
	providers *provider.Providers,
) *temporalRunner {
	return &temporalRunner{
		temporalClient: providers.TemporalClient,
		logger:         appLogger,
	}
}

// Name returns the name of the runner.
func (r *temporalRunner) Name() string {
	return "Temporal Worker"
}

// Start starts the Temporal worker.
func (r *temporalRunner) Start(_ context.Context) error {
	r.logger.Info("Starting Temporal worker")

	if r.temporalClient == nil {
		r.logger.Warn("Temporal client not available, skipping worker start")

		return nil
	}

	err := r.temporalClient.Worker.Run(worker.InterruptCh())
	if err != nil {
		r.logger.Errorf("Failed to start Temporal worker: %v", err)

		return err
	}

	return nil
}

// Shutdown gracefully shuts down the Temporal worker.
func (r *temporalRunner) Shutdown(_ context.Context) error {
	r.logger.Info("Shutting down Temporal worker")

	if r.temporalClient != nil {
		r.temporalClient.Close()
	}

	r.logger.Info("Temporal worker shutdown completed")

	return nil
}
