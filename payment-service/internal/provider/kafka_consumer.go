package provider

import (
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/event"
)

// SetupKafkaConsumers initializes the Kafka consumers for the order service.
func SetupKafkaConsumers(
	cfg *config.KafkaConfig,
	appLogger logger.Logger,
	providers *Providers,
) []mq.KafkaConsumer {
	var consumers []mq.KafkaConsumer

	productsConsumer, err := mq.NewConsumerKafka(
		cfg.Brokers,
		constant.TopicProductLifecycle,
		constant.ConsumerGroupOrderProductEvents,
		event.NewProductLifecycleConsumer(appLogger, providers.DataStore).Handler,
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
