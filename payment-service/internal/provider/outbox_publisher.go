package provider

import (
	"github.com/IBM/sarama"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/service"
)

// SetupOutboxPublisher initializes the outbox publisher service.
func SetupOutboxPublisher(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *service.OutboxPublisher {
	providers.KafkaAdmin.CreateTopic(
		constant.TopicPaymentLifecycle,
		constant.TopicPaymentLifecycleNumPartitions,
		constant.TopicPaymentLifecycleReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		constant.TopicPaymentDLQ,
		constant.TopicPaymentDLQNumPartitions,
		constant.TopicPaymentDLQReplicationFactor,
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

	registry.Register(constant.KafkaEventTypePaymentCreated, &event.PaymentLifecycleEvent{})
	registry.Register(constant.KafkaEventTypePaymentCanceled, &event.PaymentLifecycleEvent{})

	orderLifecycleProducer := event.NewPaymentLifecycleProducer(asyncProducer)
	orderDLQProducer := event.NewPaymentDLQProducer(asyncProducer)

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
