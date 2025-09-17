package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/mq/consumer"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/worker"
)

// SetupKafkaConsumers initializes the Kafka consumers for the search service.
func SetupKafkaConsumers(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *worker.KafkaConsumer {
	var consumers []kafka.Consumer

	// Create product lifecycle consumer
	productConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		kafka.ProductLifecycleTopic,
		kafka.SearchProductEventsConsumerGroup,
		consumer.NewSearchEventsConsumer(providers.DataStore, appLogger).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create product lifecycle consumer: %v", err)

		return nil
	}

	consumers = append(consumers, productConsumer)

	appLogger.Infof("successfully created %d Kafka consumers for search service", len(consumers))

	return worker.NewKafkaConsumer(appLogger, consumers)
}
