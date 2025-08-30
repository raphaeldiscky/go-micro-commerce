package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/server"
)

// GRPCWorker wraps the GRPC server as a Worker.
type GRPCWorker struct {
	server *server.GRPCServer
	logger logger.Logger
}

// NewGRPCWorker creates a new GRPC worker.
func NewGRPCWorker(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *GRPCWorker {
	return &GRPCWorker{
		server: server.NewGRPCServer(providers.ProductService, appLogger, cfg),
		logger: appLogger,
	}
}

// Name returns the name of the worker.
func (w *GRPCWorker) Name() string {
	return "gRPC Server"
}

// Start starts the GRPC server.
func (w *GRPCWorker) Start(ctx context.Context) error {
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

// Shutdown gracefully shuts down the GRPC worker.
func (w *GRPCWorker) Shutdown(ctx context.Context) error {
	return w.server.Shutdown(ctx)
}
