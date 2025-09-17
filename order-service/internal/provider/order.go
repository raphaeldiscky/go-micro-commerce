package provider

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/saga"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
)

// SetupOrder initializes the order-related routes and services.
func SetupOrder(
	ctx context.Context,
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	err := providers.KafkaAdmin.CreateTopic(
		kafka.OrderLifecycleTopic,
		constant.OrderLifecycleTopicNumPartitions,
		constant.OrderLifecycleTopicReplicationFactor,
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

	orderLifecycleProducer := producer.NewOrderLifecycleProducer(asyncProducer)
	paymentRequestProducer := producer.NewPaymentRequestProducer(asyncProducer)
	fulfillmentRequestProducer := producer.NewFulfillmentRequestProducer(
		asyncProducer,
	)

	productClient, err := client.NewProductClient(cfg)
	if err != nil {
		appLogger.Warnf(
			"failed to create product client: %v. Order service will start without product client functionality.",
			err,
		)

		productClient = nil
	}

	// Create task cancellation service
	taskCancellationService := asynq.NewTaskCancellationService(providers.AsynqInspector)

	// Create saga orchestrator
	sagaOrchestrator := saga.NewSagaOrchestrator(
		providers.DataStore,
		productClient,
		paymentRequestProducer,
		orderLifecycleProducer,
		fulfillmentRequestProducer,
		providers.FulfillmentClient,
		providers.PaymentClient,
		providers.AsynqClient,
		taskCancellationService,
		appLogger,
		cfg,
	)

	// Setup Temporal client
	temporalClient := SetupTemporal(cfg, appLogger, providers)
	providers.TemporalClient = temporalClient

	jobsScheduler := SetupJobScheduler(cfg, sagaOrchestrator, appLogger, providers)
	providers.JobScheduler = jobsScheduler
	orderService := service.NewOrderService(
		cfg,
		providers.DataStore,
		productClient,
		appLogger,
		orderLifecycleProducer,
		sagaOrchestrator,
		temporalClient,
	)
	providers.OrderService = orderService
	orderHandler := handler.NewOrderHandler(orderService, appLogger)

	routes.SetupOrderRoutes(e, orderHandler)
}
