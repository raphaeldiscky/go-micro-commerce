package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mq/consumer"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/worker"
)

// SetupKafkaConsumers initializes the Kafka consumers for the cart service.
func SetupKafkaConsumers(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *worker.KafkaConsumer {
	var consumers []kafka.Consumer

	// Consumer for order lifecycle events (order created)
	orderLifecycleConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		kafka.OrderLifecycleTopic,
		kafka.CartOrderEventsConsumerGroup,
		consumer.NewOrderLifecycleConsumer(
			appLogger,
			providers.DataStore,
		).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create order lifecycle consumer: %v", err)

		return nil
	}

	consumers = append(consumers, orderLifecycleConsumer)

	appLogger.Infof("Registered %d Kafka consumers", len(consumers))

	return worker.NewKafkaConsumer(appLogger, consumers)
}
