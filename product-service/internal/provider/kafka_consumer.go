package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/event"
)

// SetupKafkaConsumers initializes the Kafka consumers for the product service.
func SetupKafkaConsumers(
	cfg *config.KafkaConfig,
	appLogger logger.Logger,
	providers *Providers,
) []mq.KafkaConsumer {
	var consumers []mq.KafkaConsumer

	ordersConsumer, err := mq.NewConsumerKafka(
		cfg.Brokers,
		constant.TopicOrderLifecycle,
		constant.ConsumerGroupProductOrderEvents,
		event.NewOrderLifecycleConsumer(appLogger, providers.DataStore).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create order lifecycle consumer: %v", err)
		// In a real app, you might want to panic here as the service cannot run.
		return nil
	}

	consumers = append(consumers, ordersConsumer)

	appLogger.Infof("successfully created %d Kafka consumers", len(consumers))

	return consumers
}
