package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/server"
)

// KafkaConsumerWorker wraps the Kafka consumer server as a Worker.
type KafkaConsumerWorker struct {
	server *server.KafkaConsumerServer
	logger logger.Logger
}

// NewKafkaConsumerWorker creates a new Kafka consumer worker.
func NewKafkaConsumerWorker(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *KafkaConsumerWorker {
	return &KafkaConsumerWorker{
		server: server.NewKafkaConsumerServer(cfg, appLogger, providers),
		logger: appLogger,
	}
}

// Name returns the name of the worker.
func (w *KafkaConsumerWorker) Name() string {
	return "Kafka Consumer"
}

// Start starts the Kafka consumer server.
func (w *KafkaConsumerWorker) Start(ctx context.Context) error {
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

// Shutdown gracefully shuts down the Kafka consumer worker.
func (w *KafkaConsumerWorker) Shutdown(ctx context.Context) error {
	return w.server.Shutdown(ctx)
}
