package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/consumer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/worker"
)

// SetupKafkaConsumers initializes the Kafka consumers for the order service.
func SetupKafkaConsumers(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *worker.KafkaConsumer {
	var consumers []kafka.Consumer

	// Payment Lifecycle Consumer
	paymentLifecycleConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		kafka.PaymentLifecycleTopic,
		kafka.OrderPaymentEventsConsumerGroup, // Consumer group for order service consuming payment events
		consumer.NewPaymentLifecycleConsumer(
			appLogger,
			providers.DataStore,
			providers.PaymentClient,
		).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create payment lifecycle consumer: %v", err)

		return nil
	}

	// Fulfillment Lifecycle Consumer
	fulfillmentLifecycleConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		kafka.FulfillmentLifecycleTopic,
		kafka.OrderFulfillmentEventsConsumerGroup, // Consumer group for order service consuming fulfillment events
		consumer.NewFulfillmentLifecycleConsumer(
			appLogger,
			providers.DataStore,
			providers.FulfillmentClient,
		).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create fulfillment lifecycle consumer: %v", err)

		return nil
	}

	consumers = append(consumers, paymentLifecycleConsumer, fulfillmentLifecycleConsumer)

	appLogger.Infof("successfully created %d Kafka consumers", len(consumers))

	return worker.NewKafkaConsumer(appLogger, consumers)
}
