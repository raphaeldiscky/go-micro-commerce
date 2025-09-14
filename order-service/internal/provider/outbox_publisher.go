package provider

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/worker"
)

// SetupOutboxPublisher initializes the outbox publisher service.
func SetupOutboxPublisher(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *worker.OutboxPublisher {
	err := providers.KafkaAdmin.CreateTopic(
		kafka.OrderLifecycleTopic,
		constant.OrderLifecycleTopicNumPartitions,
		constant.OrderLifecycleTopicReplicationFactor,
	)
	if err != nil {
		appLogger.Fatalf("failed to create Kafka topic: %v", err)
	}

	err = providers.KafkaAdmin.CreateTopic(
		kafka.OrderDLQTopic,
		constant.OrderDLQTopicNumPartitions,
		constant.OrderDLQTopicReplicationFactor,
	)
	if err != nil {
		appLogger.Fatalf("failed to create Kafka topic: %v", err)
	}

	err = providers.KafkaAdmin.CreateTopic(
		kafka.PaymentRequestTopic,
		constant.PaymentRequestTopicNumPartitions,
		constant.PaymentRequestTopicReplicationFactor,
	)
	if err != nil {
		appLogger.Fatalf("failed to create Kafka topic: %v", err)
	}

	err = providers.KafkaAdmin.CreateTopic(
		kafka.PaymentDLQTopic,
		constant.PaymentDLQTopicNumPartitions,
		constant.PaymentDLQTopicReplicationFactor,
	)
	if err != nil {
		appLogger.Fatalf("failed to create Kafka topic: %v", err)
	}

	err = providers.KafkaAdmin.CreateTopic(
		kafka.FulfillmentRequestTopic,
		constant.FulfillmentRequestTopicNumPartitions,
		constant.FulfillmentRequestTopicReplicationFactor,
	)
	if err != nil {
		appLogger.Fatalf("failed to create Kafka topic: %v", err)
	}

	err = providers.KafkaAdmin.CreateTopic(
		kafka.FulfillmentDLQTopic,
		constant.FulfillmentDLQTopicNumPartitions,
		constant.FulfillmentDLQTopicReplicationFactor,
	)
	if err != nil {
		appLogger.Fatalf("failed to create Kafka topic: %v", err)
	}

	err = providers.KafkaAdmin.CreateTopic(
		kafka.NotificationRequestTopic,
		constant.NotificationRequestTopicNumPartitions,
		constant.NotificationRequestTopicReplicationFactor,
	)
	if err != nil {
		appLogger.Fatalf("failed to create Kafka topic: %v", err)
	}

	err = providers.KafkaAdmin.CreateTopic(
		kafka.NotificationDLQTopic,
		constant.NotificationDLQTopicNumPartitions,
		constant.NotificationDLQTopicReplicationFactor,
	)
	if err != nil {
		appLogger.Fatalf("failed to create Kafka topic: %v", err)
	}

	registry := kafka.NewEventRegistry()
	// Create Kafka producer for outbox events
	asyncProducer, err := kafka.NewAsyncProducer(ctx, &kafka.ProducerConfig{
		Brokers:        cfg.Kafka.Brokers,
		RetryMax:       cfg.Kafka.RetryMax,
		RetryInterval:  cfg.Kafka.RetryInterval,
		FlushFrequency: cfg.Kafka.FlushFrequency,
		ReturnSuccess:  cfg.Kafka.ReturnSuccess,
		ReturnErrors:   cfg.Kafka.ReturnErrors,
		Acks:           sarama.WaitForAll,
	}, appLogger)
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

	providers.NotificationRequestProducer = notificationRequestProducer

	// DLQ
	orderDLQProducer := producer.NewOrderDLQProducer(asyncProducer)
	paymentDLQProducer := producer.NewPaymentDLQProducer(asyncProducer)
	fulfillmentDLQProducer := producer.NewFulfillmentDLQProducer(asyncProducer)
	notificationDLQProducer := producer.NewNotificationDLQProducer(asyncProducer)

	// Create outbox publisher
	outboxPublisher := worker.NewOutboxPublisher(
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
