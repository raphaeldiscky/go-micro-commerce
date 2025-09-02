package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq"
)

// SetupKafkaConsumers initializes the Kafka consumers for the order service.
func SetupKafkaConsumers(
	cfg *config.KafkaConfig,
	appLogger logger.Logger,
	providers *Providers,
) []kafka.Consumer {
	var consumers []kafka.Consumer

	productsConsumer, err := kafka.NewConsumer(
		cfg.Brokers,
		event.ProductLifecycleTopic,
		event.OrderProductEventsConsumerGroup,
		mq.NewProductLifecycleConsumer(appLogger, providers.DataStore).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create product lifecycle consumer: %v", err)
		// In a real app, you might want to panic here as the service cannot run.
		return nil
	}

	consumers = append(consumers, productsConsumer)

	appLogger.Infof("successfully created %d Kafka consumers", len(consumers))

	return consumers
}
