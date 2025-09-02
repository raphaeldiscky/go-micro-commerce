// Package server provides a Kafka server implementation for consuming messages from Kafka topics.
package server

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/provider"
)

// KafkaConsumerServer represents a server for consuming messages from Kafka topics.
type KafkaConsumerServer struct {
	ctx       context.Context
	cancel    context.CancelFunc
	logger    logger.Logger
	consumers []kafka.Consumer
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
		logger:    appLogger,
		consumers: consumers,
	}
}

// Start begins the Kafka consumer server.
func (s *KafkaConsumerServer) Start() error {
	s.logger.Info("Running Kafka consumer server...")

	for _, consumer := range s.consumers {
		s.wg.Add(1)

		go func(c kafka.Consumer) {
			defer s.wg.Done()

			if err := c.Consume(s.ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					s.logger.Infof("Consumer for topic %s stopped gracefully", c.Topic())
				} else {
					s.logger.Errorf("Consumer error for topic %s: %v", c.Topic(), err)
				}
			} else {
				s.logger.Infof("Consumer for topic %s stopped", c.Topic())
			}
		}(consumer)
	}

	s.logger.Info("Kafka server is running...")

	return nil
}

// Shutdown gracefully stops the Kafka consumer server.
func (s *KafkaConsumerServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Attempting to shut down the Kafka server...")

	// Cancel the internal context to signal all consumers to stop
	s.cancel()

	// Wait for all consumers to finish gracefully or timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("All consumers stopped gracefully")
	case <-ctx.Done():
		s.logger.Warn("Shutdown timeout reached while waiting for consumers")
	}

	// Close all consumers
	var closeErrors []error

	for _, consumer := range s.consumers {
		if err := consumer.Close(); err != nil {
			s.logger.Errorf("Error closing consumer %v: %v", consumer.Topic(), err)
			closeErrors = append(closeErrors, err)
		}
	}

	if len(closeErrors) > 0 {
		s.logger.Errorf("Kafka server shutdown completed with %d errors", len(closeErrors))

		return fmt.Errorf("kafka server shutdown errors: %v", closeErrors)
	}

	s.logger.Info("Kafka server shut down completed successfully")

	return nil
}
