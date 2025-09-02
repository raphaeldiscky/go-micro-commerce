package provider

import (
	"github.com/IBM/sarama"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// SetupOutboxPublisher initializes the outbox publisher service.
func SetupOutboxPublisher(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *service.OutboxPublisher {
	providers.KafkaAdmin.CreateTopic(
		event.PaymentLifecycleTopic,
		constant.PaymentLifecycleTopicNumPartitions,
		constant.PaymentLifecycleTopicReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		event.PaymentDLQTopic,
		constant.PaymentDLQTopicNumPartitions,
		constant.PaymentDLQTopicReplicationFactor,
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

	registry.Register(event.PaymentCreatedEventType, &mq.PaymentLifecycleEvent{})
	registry.Register(event.PaymentFailedEventType, &mq.PaymentLifecycleEvent{})
	registry.Register(event.PaymentCompletedEventType, &mq.PaymentLifecycleEvent{})

	orderLifecycleProducer := mq.NewPaymentLifecycleProducer(asyncProducer)
	orderDLQProducer := mq.NewPaymentDLQProducer(asyncProducer)

	// Create outbox publisher
	outboxPublisher := service.NewOutboxPublisher(
		providers.DataStore,
		appLogger,
		orderLifecycleProducer,
		orderDLQProducer,
		*cfg.OutboxPublisher,
		registry,
	)

	return outboxPublisher
}
