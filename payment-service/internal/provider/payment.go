package provider

import (
	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// SetupPayment initializes the order-related routes and services.
func SetupPayment(cfg *config.Config, e *echo.Echo, appLogger logger.Logger, providers *Providers) {
	providers.KafkaAdmin.CreateTopic(
		kafka.PaymentLifecycleTopic,
		constant.PaymentLifecycleTopicNumPartitions,
		constant.PaymentLifecycleTopicReplicationFactor,
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

	orderLifecycleProducer := mq.NewPaymentLifecycleProducer(asyncProducer)

	orderService := service.NewPaymentService(
		providers.DataStore,
		appLogger,
		orderLifecycleProducer,
	)
	orderHandler := handler.NewPaymentHandler(orderService, appLogger)

	routes.SetupPaymentRoutes(e, orderHandler)
}
