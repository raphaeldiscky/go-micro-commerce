package provider

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
)

// SetupCheckoutSession initializes the checkout-session-related routes and services.
func SetupCheckoutSession(
	ctx context.Context,
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	tel *telemetry.Telemetry,
	providers *Providers,
) {
	err := providers.KafkaAdmin.CreateTopic(
		kafka.NotificationRequestTopic,
		constant.NotificationRequestedTopicNumPartitions,
		constant.NotificationRequestedTopicReplicationFactor,
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

	notificationRequestedProducer := producer.NewNotificationRequestProducer(
		asyncProducer,
	)

	providers.NotificationRequestProducer = notificationRequestedProducer
	taskCancellationService := asynq.NewTaskCancellationService(providers.AsynqInspector)

	checkoutSessionService := service.NewCheckoutSessionService(
		providers.DataStore,
		appLogger,
		providers.ProductClient,
		providers.AsynqClient,
		taskCancellationService,
	)
	providers.CheckoutSessionService = checkoutSessionService
	checkoutSessionHandler := handler.NewCheckoutSessionHandler(
		checkoutSessionService,
		tel,
	)

	routes.SetupCheckoutSessionRoutes(e, checkoutSessionHandler)
}
