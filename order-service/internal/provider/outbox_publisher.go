package provider

import (
	"github.com/IBM/sarama"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
)

// SetupOutboxPublisher initializes the outbox publisher service.
func SetupOutboxPublisher(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *service.OutboxPublisher {
	providers.KafkaAdmin.CreateTopic(
		constant.TopicOrderLifecycle,
		constant.TopicOrderLifecycleNumPartitions,
		constant.TopicOrderLifecycleReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		constant.TopicOrderDLQ,
		constant.TopicOrderDLQNumPartitions,
		constant.TopicOrderDLQReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		constant.TopicPaymentRequest,
		constant.TopicPaymentRequestNumPartitions,
		constant.TopicPaymentRequestReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		constant.TopicPaymentDLQ,
		constant.TopicPaymentDLQNumPartitions,
		constant.TopicPaymentDLQReplicationFactor,
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

	registry.Register(constant.KafkaEventTypeOrderCreated, &mq.OrderLifecycleEvent{})
	registry.Register(constant.KafkaEventTypeOrderCanceled, &mq.OrderLifecycleEvent{})
	registry.Register(constant.KafkaEventTypePaymentRequested, &mq.PaymentRequestEvent{})

	// Producers
	orderLifecycleProducer := mq.NewOrderLifecycleProducer(asyncProducer)
	paymentRequestProducer := mq.NewPaymentRequestProducer(asyncProducer)

	// DLQ
	orderDLQProducer := mq.NewOrderDLQProducer(asyncProducer)
	paymentDLQProducer := mq.NewPaymentDLQProducer(asyncProducer)

	// Create outbox publisher
	outboxPublisher := service.NewOutboxPublisher(
		providers.DataStore,
		appLogger,
		orderLifecycleProducer,
		orderDLQProducer,
		paymentRequestProducer,
		paymentDLQProducer,
		*cfg.OutboxPublisher,
		registry,
	)

	return outboxPublisher
}
