// Package server provides a Kafka server implementation for consuming messages from Kafka topics.
package server

import (
	"context"
	"sync"
	"time"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/provider"
)

// KafkaConsumerServer represents a server for consuming messages from Kafka topics.
type KafkaConsumerServer struct {
	ctx       context.Context
	cancel    context.CancelFunc
	appLogger logger.Logger
	consumers []mq.KafkaConsumer
	wg        sync.WaitGroup
}

// NewKafkaConsumerServer creates a new Kafka consumer server.
func NewKafkaConsumerServer(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *KafkaConsumerServer {
	ctx, cancel := context.WithCancel(context.Background())
	consumers := provider.SetupKafkaConsumers(cfg.Kafka, appLogger, providers)

	return &KafkaConsumerServer{
		ctx:       ctx,
		cancel:    cancel,
		appLogger: appLogger,
		consumers: consumers,
	}
}

// Start begins the Kafka consumer server.
func (s *KafkaConsumerServer) Start() error {
	s.appLogger.Info("Running Kafka consumer server...")

	for _, consumer := range s.consumers {
		s.wg.Add(1)

		go func(c mq.KafkaConsumer) {
			defer s.wg.Done()

			if err := c.Consume(s.ctx); err != nil {
				s.appLogger.Errorf("Consumer error for topic %s: %v", c.Topic(), err)
			}
		}(consumer)
	}

	s.appLogger.Info("Kafka server is running...")

	return nil
}

// Shutdown gracefully stops the Kafka consumer server.
func (s *KafkaConsumerServer) Shutdown() {
	s.appLogger.Info("Attempting to shut down the Kafka server...")

	// Cancel the context to signal all consumers to stop
	s.cancel()

	// Wait for all consumers to finish gracefully
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// Wait for graceful shutdown or timeout
	select {
	case <-done:
		s.appLogger.Info("All consumers stopped gracefully")
	case <-time.After(30 * time.Second):
		s.appLogger.Warn("Timeout waiting for consumers to stop")
	}

	// Close all consumers
	for _, consumer := range s.consumers {
		if err := consumer.Close(); err != nil {
			s.appLogger.Errorf("Error closing consumer %v: %v", consumer.Topic(), err)
		}
	}

	s.appLogger.Info("Kafka server shut down completed")
}
