package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/provider"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/server"
)

// HTTPWorker wraps the HTTP server as a Worker.
type HTTPWorker struct {
	server *server.HTTPServer
	logger logger.Logger
}

// NewHTTPWorker creates a new HTTP worker.
func NewHTTPWorker(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
	gw *gateway.Gateway,
) *HTTPWorker {
	return &HTTPWorker{
		server: server.NewHTTPServer(cfg, appLogger, providers, gw),
		logger: appLogger,
	}
}

// Name returns the name of the worker.
func (w *HTTPWorker) Name() string {
	return "HTTP Server"
}

// Start starts the HTTP server.
func (w *HTTPWorker) Start(ctx context.Context) error {
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

// Shutdown gracefully shuts down the HTTP worker.
func (w *HTTPWorker) Shutdown(ctx context.Context) error {
	return w.server.Shutdown(ctx)
}
