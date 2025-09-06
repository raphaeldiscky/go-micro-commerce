package provider

import (
	"github.com/IBM/sarama"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
)

// SetupOutboxPublisher initializes the outbox publisher service.
func SetupOutboxPublisher(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *service.OutboxPublisher {
	providers.KafkaAdmin.CreateTopic(
		kafka.OrderLifecycleTopic,
		constant.OrderLifecycleTopicNumPartitions,
		constant.OrderLifecycleTopicReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		kafka.OrderDLQTopic,
		constant.OrderDLQTopicNumPartitions,
		constant.OrderDLQTopicReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		kafka.PaymentRequestTopic,
		constant.PaymentRequestTopicNumPartitions,
		constant.PaymentRequestTopicReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		kafka.PaymentDLQTopic,
		constant.PaymentDLQTopicNumPartitions,
		constant.PaymentDLQTopicReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		kafka.FulfillmentRequestTopic,
		constant.FulfillmentRequestTopicNumPartitions,
		constant.FulfillmentRequestTopicReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		kafka.FulfillmentDLQTopic,
		constant.FulfillmentDLQTopicNumPartitions,
		constant.FulfillmentDLQTopicReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		kafka.NotificationRequestTopic,
		constant.NotificationRequestTopicNumPartitions,
		constant.NotificationRequestTopicReplicationFactor,
	)

	providers.KafkaAdmin.CreateTopic(
		kafka.NotificationDLQTopic,
		constant.NotificationDLQTopicNumPartitions,
		constant.NotificationDLQTopicReplicationFactor,
	)

	registry := kafka.NewEventRegistry()
	// Create Kafka producer for outbox events
	asyncProducer, err := kafka.NewAsyncProducer(&kafka.ProducerConfig{
		Brokers:        cfg.Kafka.Brokers,
		RetryMax:       cfg.Kafka.RetryMax,
		FlushFrequency: cfg.Kafka.FlushFrequency,
		ReturnSuccess:  cfg.Kafka.ReturnSuccess,
		ReturnErrors:   cfg.Kafka.ReturnErrors,
		Acks:           sarama.WaitForAll,
	})
	if err != nil {
		appLogger.Fatalf("failed to create outbox Kafka producer: %v", err)
	}

	registry.Register(kafka.OrderCreatedEventType, &producer.OrderLifecycleEvent{})
	registry.Register(kafka.OrderCanceledEventType, &producer.OrderLifecycleEvent{})
	registry.Register(kafka.PaymentRequestedEventType, &producer.PaymentRequestEvent{})
	registry.Register(kafka.FulfillmentRequestedEventType, &producer.FulfillmentRequestEvent{})
	registry.Register(kafka.NotificationRequestedEventType, &producer.NotificationRequestEvent{})

	// Producers
	orderLifecycleProducer := producer.NewOrderLifecycleProducer(asyncProducer)
	paymentRequestProducer := producer.NewPaymentRequestProducer(asyncProducer)
	fulfillmentRequestProducer := producer.NewFulfillmentRequestProducer(asyncProducer)
	notificationRequestProducer := producer.NewNotificationRequestProducer(asyncProducer)

	// DLQ
	orderDLQProducer := producer.NewOrderDLQProducer(asyncProducer)
	paymentDLQProducer := producer.NewPaymentDLQProducer(asyncProducer)
	fulfillmentDLQProducer := producer.NewFulfillmentDLQProducer(asyncProducer)
	notificationDLQProducer := producer.NewNotificationDLQProducer(asyncProducer)

	// Create outbox publisher
	outboxPublisher := service.NewOutboxPublisher(
		providers.DataStore,
		appLogger,
		orderLifecycleProducer,
		orderDLQProducer,
		paymentRequestProducer,
		paymentDLQProducer,
		fulfillmentRequestProducer,
		fulfillmentDLQProducer,
		notificationRequestProducer,
		notificationDLQProducer,
		*cfg.OutboxPublisher,
		registry,
	)

	return outboxPublisher
}
