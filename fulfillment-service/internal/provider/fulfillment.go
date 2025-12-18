package provider

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/oapi/handler"
)

// SetupFulfillment initializes the order-related routes and services.
func SetupFulfillment(
	ctx context.Context,
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	err := providers.KafkaAdmin.CreateTopic(
		kafka.FulfillmentLifecycleTopic,
		constant.FulfillmentLifecycleTopicNumPartitions,
		constant.FulfillmentLifecycleTopicReplicationFactor,
	)
	if err != nil {
		appLogger.Fatalf("failed to create Kafka topic: %v", err)
	}

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
		appLogger.Fatalf("failed to create Kafka async producer: %v", err)
	}

	fulfillmentLifecycleProducer := mq.NewFulfillmentLifecycleProducer(asyncProducer)

	fulfillmentService := service.NewFulfillmentService(
		providers.DataStore,
		appLogger,
		fulfillmentLifecycleProducer,
		providers.CourierClient,
	)
	providers.FulfillmentService = fulfillmentService
	apiHandler := handler.NewHandler(fulfillmentService)

	routes.SetupFulfillmentRoutes(e, apiHandler)
}
