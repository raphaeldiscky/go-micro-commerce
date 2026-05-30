package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/worker"
)

// kafkaConsumerRunner wraps the Kafka consumer as a Runner.
type kafkaConsumerRunner struct {
	consumer *worker.KafkaConsumer
}

// newKafkaConsumerRunner creates a new Kafka consumer runner.
func newKafkaConsumerRunner(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *kafkaConsumerRunner {
	return &kafkaConsumerRunner{
		consumer: provider.SetupKafkaConsumers(cfg, appLogger, providers),
	}
}

// Name returns the name of the runner.
func (r *kafkaConsumerRunner) Name() string {
	return "Kafka Consumer"
}

// Start starts the Kafka consumer.
func (r *kafkaConsumerRunner) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if err := r.consumer.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil // Context canceled, normal shutdown
	case err := <-errChan:
		return err // Consumer error
	}
}

// Shutdown gracefully shuts down the Kafka consumer.
func (r *kafkaConsumerRunner) Shutdown(ctx context.Context) error {
	return r.consumer.Shutdown(ctx)
}

// newKafkaConsumerCmd runs the Kafka consumer role.
func newKafkaConsumerCmd() *cobra.Command {
	return roleCmd(
		"kafka-consumer",
		"Run the Kafka event consumer",
		func(app *appContext) ([]Runner, func()) {
			runner := newKafkaConsumerRunner(app.cfg, app.logger, app.providers)

			return []Runner{runner}, nil
		},
	)
}
