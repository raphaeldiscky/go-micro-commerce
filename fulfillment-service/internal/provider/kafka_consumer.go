package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/mq"
)

// SetupKafkaConsumers initializes the Kafka consumers for the fulfillment service.
func SetupKafkaConsumers(
	cfg *config.KafkaConfig,
	appLogger logger.Logger,
	providers *Providers,
) []kafka.Consumer {
	var consumers []kafka.Consumer

	// Consumer for order lifecycle events (order created, updated, deleted)
	ordersConsumer, err := kafka.NewConsumer(
		cfg.Brokers,
		kafka.OrderLifecycleTopic,
		kafka.FulfillmentOrderEventsConsumerGroup,
		mq.NewOrderLifecycleConsumer(appLogger, providers.DataStore).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create order lifecycle consumer: %v", err)

		return nil
	}

	consumers = append(consumers, ordersConsumer)

	// Consumer for fulfillment request events
	fulfillmentRequestConsumer, err := kafka.NewConsumer(
		cfg.Brokers,
		kafka.FulfillmentRequestTopic,
		kafka.FulfillmentEventsConsumerGroup,
		mq.NewFulfillmentRequestConsumer(
			appLogger,
			providers.DataStore,
			providers.CarrierClient,
		).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create fulfillment request consumer: %v", err)

		return nil
	}

	consumers = append(consumers, fulfillmentRequestConsumer)

	appLogger.Infof("successfully created %d Kafka consumers", len(consumers))

	return consumers
}
