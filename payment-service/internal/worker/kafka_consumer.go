// Package worker provides a Kafka worker implementation for consuming messages from Kafka topics.
package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
)

// KafkaConsumer represents a worker for consuming messages from Kafka topics.
type KafkaConsumer struct {
	ctx       context.Context
	cancel    context.CancelFunc
	logger    logger.Logger
	config    *config.Config
	consumers []kafka.Consumer
	wg        sync.WaitGroup
}

// NewKafkaConsumer creates a new Kafka consumer worker.
func NewKafkaConsumer(
	cfg *config.Config,
	appLogger logger.Logger,
	consumers []kafka.Consumer,
) *KafkaConsumer {
	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaConsumer{
		ctx:       ctx,
		cancel:    cancel,
		logger:    appLogger,
		config:    cfg,
		consumers: consumers,
	}
}

// Start begins the Kafka consumer worker.
func (s *KafkaConsumer) Start() error {
	s.logger.Info("Running Kafka consumer worker...")

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

	s.logger.Info("Kafka worker is running...")

	return nil
}

// Shutdown gracefully stops the Kafka consumer worker.
func (s *KafkaConsumer) Shutdown(ctx context.Context) error {
	s.logger.Info("Attempting to shut down the Kafka worker...")

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
		s.logger.Errorf("Kafka worker shutdown completed with %d errors", len(closeErrors))

		return fmt.Errorf("kafka worker shutdown errors: %v", closeErrors)
	}

	s.logger.Info("Kafka worker shut down completed successfully")

	return nil
}
