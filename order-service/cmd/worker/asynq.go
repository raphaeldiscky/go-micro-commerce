package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/provider"
)

// AsynqWorker wraps the asynq server as a Worker.
type AsynqWorker struct {
	asynqProvider *provider.AsynqProvider
	logger        logger.Logger
}

// NewAsynqWorker creates a new asynq worker.
func NewAsynqWorker(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) (*AsynqWorker, error) {
	asynqProvider, err := provider.SetupAsynq(cfg, providers, appLogger)
	if err != nil {
		return nil, err
	}

	return &AsynqWorker{
		asynqProvider: asynqProvider,
		logger:        appLogger,
	}, nil
}

// Name returns the name of the worker.
func (w *AsynqWorker) Name() string {
	return "Asynq Worker"
}

// Start starts the asynq worker.
func (w *AsynqWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting asynq worker...")

	// Start server in goroutine
	errChan := make(chan error, 1)

	go func() {
		if err := w.asynqProvider.Server.Start(w.asynqProvider.Mux); err != nil {
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

// Shutdown gracefully shuts down the asynq worker.
func (w *AsynqWorker) Shutdown(_ context.Context) error {
	w.logger.Info("Stopping asynq worker...")
	w.asynqProvider.Server.Stop()
	w.logger.Info("Asynq worker stopped successfully")

	return nil
}
