package provider

import (
	"github.com/IBM/sarama"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/service"
)

// SetupOutboxPublisher initializes the outbox publisher service.
func SetupOutboxPublisher(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *service.OutboxPublisher {
	providers.KafkaAdmin.CreateTopic(
		kafka.FulfillmentLifecycleTopic,
		constant.FulfillmentLifecycleTopicNumPartitions,
		constant.FulfillmentLifecycleTopicReplicationFactor,
	)
	providers.KafkaAdmin.CreateTopic(
		kafka.FulfillmentDLQTopic,
		constant.FulfillmentDLQTopicNumPartitions,
		constant.FulfillmentDLQTopicReplicationFactor,
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

	registry.Register(kafka.FulfillmentCreatedEventType, &mq.FulfillmentLifecycleEvent{})
	registry.Register(kafka.FulfillmentShippedEventType, &mq.FulfillmentLifecycleEvent{})
	registry.Register(kafka.FulfillmentDeliveredEventType, &mq.FulfillmentLifecycleEvent{})

	fulfillmentLifecycleProducer := mq.NewFulfillmentLifecycleProducer(asyncProducer)
	fulfillmentDLQProducer := mq.NewFulfillmentDLQProducer(asyncProducer)

	// Create outbox publisher
	outboxPublisher := service.NewOutboxPublisher(
		providers.DataStore,
		appLogger,
		fulfillmentLifecycleProducer,
		fulfillmentDLQProducer,
		*cfg.OutboxPublisher,
		registry,
	)

	return outboxPublisher
}
