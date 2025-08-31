package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/event"
)

// SetupKafkaConsumers initializes the Kafka consumers for the payment service.
func SetupKafkaConsumers(
	cfg *config.KafkaConfig,
	appLogger logger.Logger,
	providers *Providers,
) []mq.KafkaConsumer {
	var consumers []mq.KafkaConsumer

	// Consumer for order lifecycle events (order created, updated, deleted)
	ordersConsumer, err := mq.NewConsumerKafka(
		cfg.Brokers,
		constant.TopicOrderLifecycle,
		constant.ConsumerGroupPaymentOrderEvents,
		event.NewOrderLifecycleConsumer(appLogger, providers.DataStore).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create order lifecycle consumer: %v", err)

		return nil
	}

	// Consumer for payment request events from order service
	paymentRequestConsumer, err := mq.NewConsumerKafka(
		cfg.Brokers,
		constant.TopicPaymentRequest,
		constant.ConsumerGroupPaymentEvents,
		event.NewPaymentRequestConsumer(appLogger, providers.DataStore).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create payment request consumer: %v", err)

		return nil
	}

	consumers = append(consumers, ordersConsumer, paymentRequestConsumer)

	appLogger.Infof("successfully created %d Kafka consumers", len(consumers))

	return consumers
}
