package provider

import (
	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/service"
)

// SetupPayment initializes the order-related routes and services.
func SetupPayment(cfg *config.Config, e *echo.Echo, appLogger logger.Logger, providers *Providers) {
	providers.KafkaAdmin.CreateTopic(
		constant.TopicPaymentLifecycle,
		constant.TopicPaymentLifecycleNumPartitions,
		constant.TopicPaymentLifecycleReplicationFactor,
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

	orderLifecycleProducer := event.NewPaymentLifecycleProducer(asyncProducer)

	orderService := service.NewPaymentService(
		providers.DataStore,
		appLogger,
		orderLifecycleProducer,
	)
	orderHandler := handler.NewPaymentHandler(orderService, appLogger)

	routes.SetupPaymentRoutes(e, orderHandler)
}
