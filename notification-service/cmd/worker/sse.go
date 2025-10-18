package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/server"
)

// SSEWorker wraps the SSE server as a Worker.
type SSEWorker struct {
	server *server.SSEServer
}

// NewSSEWorker creates a new SSE worker.
func NewSSEWorker(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *SSEWorker {
	return &SSEWorker{
		server: server.NewSSEServer(cfg, appLogger, providers),
	}
}

// Name returns the name of the worker.
func (w *SSEWorker) Name() string {
	return "SSE Server"
}

// Start starts the SSE server.
func (w *SSEWorker) Start(ctx context.Context) error {
	// Start server in goroutine
	errChan := make(chan error, 1)

	go func() {
		if err := w.server.Start(); err != nil {
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

// Shutdown gracefully shuts down the SSE worker.
func (w *SSEWorker) Shutdown(ctx context.Context) error {
	return w.server.Shutdown(ctx)
}
