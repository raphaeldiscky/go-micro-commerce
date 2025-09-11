package provider

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/worker"
)

// SetupOutboxPublisher initializes the outbox publisher service.
func SetupOutboxPublisher(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *worker.OutboxPublisher {
	err := providers.KafkaAdmin.CreateTopic(
		kafka.PaymentLifecycleTopic,
		constant.PaymentLifecycleTopicNumPartitions,
		constant.PaymentLifecycleTopicReplicationFactor,
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
	})
	if err != nil {
		appLogger.Fatalf("failed to create outbox Kafka producer: %v", err)
	}

	registry.Register(kafka.PaymentCreatedEventType, &producer.PaymentLifecycleEvent{})
	registry.Register(kafka.PaymentFailedEventType, &producer.PaymentLifecycleEvent{})
	registry.Register(kafka.PaymentCompletedEventType, &producer.PaymentLifecycleEvent{})

	orderLifecycleProducer := producer.NewPaymentLifecycleProducer(asyncProducer)
	orderDLQProducer := producer.NewPaymentDLQProducer(asyncProducer)

	// Create outbox publisher
	outboxPublisher := worker.NewOutboxPublisher(
		providers.DataStore,
		appLogger,
		orderLifecycleProducer,
		orderDLQProducer,
		*cfg.OutboxPublisher,
		registry,
	)

	return outboxPublisher
}
