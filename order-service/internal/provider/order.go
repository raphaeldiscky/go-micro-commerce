package provider

import (
	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/service"
)

// SetupOrder initializes the order-related routes and services.
func SetupOrder(cfg *config.Config, e *echo.Echo, appLogger logger.Logger, providers *Providers) {
	providers.KafkaAdmin.CreateTopic(
		constant.TopicOrderLifecycle,
		constant.TopicOrderLifecycleNumPartitions,
		constant.TopicOrderLifecycleReplicationFactor,
	)

	asyncProducer, err := mq.NewKafkaAsyncProducer(&mq.KafkaProducerConfig{
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

	orderLifecycleProducer := event.NewOrderLifecycleProducer(asyncProducer)

	productClient, err := client.NewProductClient(cfg.Client, cfg.Consul)
	if err != nil {
		appLogger.Warnf(
			"failed to create product client: %v. Order service will start without product client functionality.",
			err,
		)

		productClient = nil
	}

	orderService := service.NewOrderService(
		providers.DataStore,
		productClient,
		appLogger,
		orderLifecycleProducer,
	)
	orderHandler := handler.NewOrderHandler(orderService, appLogger)

	routes.SetupOrderRoutes(e, orderHandler)
}
