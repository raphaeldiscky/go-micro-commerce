package provider

import (
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

	// Payment Lifecycle Consumer
	paymentLifecycleConsumer, err := kafka.NewConsumer(
		cfg.Brokers,
		kafka.PaymentLifecycleTopic,
		"order-service.payment-events", // Consumer group for order service consuming payment events
		mq.NewPaymentLifecycleConsumer(appLogger, providers.DataStore).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create payment lifecycle consumer: %v", err)

		return nil
	}

	// Fulfillment Lifecycle Consumer
	fulfillmentLifecycleConsumer, err := kafka.NewConsumer(
		cfg.Brokers,
		kafka.FulfillmentLifecycleTopic,
		kafka.OrderFulfillmentEventsConsumerGroup, // Consumer group for order service consuming fulfillment events
		mq.NewFulfillmentLifecycleConsumer(appLogger, providers.DataStore).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create fulfillment lifecycle consumer: %v", err)

		return nil
	}

	consumers = append(consumers, paymentLifecycleConsumer, fulfillmentLifecycleConsumer)

	appLogger.Infof("successfully created %d Kafka consumers", len(consumers))

	return consumers
}
