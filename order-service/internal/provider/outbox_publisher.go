package provider

import (
	"github.com/IBM/sarama"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/service"
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

	registry := mq.NewEventRegistry()
	// Create Kafka producer for outbox events
	asyncProducer, err := mq.NewKafkaAsyncProducer(&mq.KafkaProducerConfig{
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

	registry.Register(constant.KafkaEventTypeOrderCreated, &event.OrderLifecycleEvent{})
	registry.Register(constant.KafkaEventTypeOrderCanceled, &event.OrderLifecycleEvent{})

	orderLifecycleProducer := event.NewOrderLifecycleProducer(asyncProducer)
	orderDLQProducer := event.NewOrderDLQProducer(asyncProducer)

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
