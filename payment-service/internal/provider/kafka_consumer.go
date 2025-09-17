package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mq/consumer"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/worker"
)

// SetupKafkaConsumers initializes the Kafka consumers for the payment service.
func SetupKafkaConsumers(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *worker.KafkaConsumer {
	var consumers []kafka.Consumer

	// Consumer for order lifecycle events (order created, updated, deleted)
	ordersConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		kafka.OrderLifecycleTopic,
		kafka.PaymentOrderEventsConsumerGroup,
		consumer.NewOrderLifecycleConsumer(
			appLogger,
			providers.DataStore,
			providers.PaymentService,
		).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create order lifecycle consumer: %v", err)

		return nil
	}

	// Consumer for payment request events from order service
	paymentRequestConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		kafka.PaymentRequestTopic,
		kafka.PaymentEventsConsumerGroup,
		consumer.NewPaymentRequestConsumer(appLogger, providers.DataStore).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create payment request consumer: %v", err)

		return nil
	}

	consumers = append(consumers, ordersConsumer, paymentRequestConsumer)

	appLogger.Infof("successfully created %d Kafka consumers", len(consumers))

	return worker.NewKafkaConsumer(cfg, appLogger, consumers)
}
