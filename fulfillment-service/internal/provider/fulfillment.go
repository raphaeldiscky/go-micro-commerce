package provider

import (
	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/service"
)

// SetupFulfillment initializes the order-related routes and services.
func SetupFulfillment(
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	providers.KafkaAdmin.CreateTopic(
		kafka.FulfillmentLifecycleTopic,
		constant.FulfillmentLifecycleTopicNumPartitions,
		constant.FulfillmentLifecycleTopicReplicationFactor,
	)

	asyncProducer, err := kafka.NewAsyncProducer(&kafka.ProducerConfig{
		Brokers:        cfg.Kafka.Brokers,
		RetryMax:       cfg.Kafka.RetryMax,
		FlushFrequency: cfg.Kafka.FlushFrequency,
		ReturnSuccess:  cfg.Kafka.ReturnSuccess,
		ReturnErrors:   cfg.Kafka.ReturnErrors,
		Acks:           sarama.WaitForAll,
	})
	if err != nil {
		appLogger.Fatalf("failed to create Kafka async producer: %v", err)
	}

	fulfillmentLifecycleProducer := mq.NewFulfillmentLifecycleProducer(asyncProducer)

	fulfillmentService := service.NewFulfillmentService(
		providers.DataStore,
		appLogger,
		fulfillmentLifecycleProducer,
		providers.CarrierClient,
	)
	providers.FulfillmentService = fulfillmentService
	fulfillmentHandler := handler.NewFulfillmentHandler(fulfillmentService, appLogger)

	routes.SetupFulfillmentRoutes(e, fulfillmentHandler)
}
