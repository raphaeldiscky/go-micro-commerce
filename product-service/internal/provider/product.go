package provider

import (
	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/service"
)

// SetupProduct initializes the product-related routes and services.
func SetupProduct(cfg *config.Config, e *echo.Echo, appLogger logger.Logger, providers *Providers) {
	providers.KafkaAdmin.CreateTopic(
		constant.ProductLifecycleTopic,
		constant.ProductLifecycleTopicNumPartitions,
		constant.ProductLifecycleTopicReplicationFactor,
	)

	asyncProducer, err := mq.NewKafkaAsyncProducer(&mq.KafkaProducerConfig{
		Brokers:        cfg.Kafka.Brokers,
		RetryMax:       cfg.Kafka.RetryMax,
		FlushFrequency: cfg.Kafka.FlushFrequency,
		ReturnSuccess:  cfg.Kafka.ReturnSuccess,
		ReturnErrors:   true, // Enable error returns for better error handling
		Acks:           sarama.WaitForLocal,
	})
	if err != nil {
		appLogger.Fatalf("failed to create Kafka async producer: %v", err)
	}

	productCreatedProducer := event.NewProductCreatedProducer(asyncProducer)
	productUpdatedProducer := event.NewProductUpdatedProducer(asyncProducer)
	productDeletedProducer := event.NewProductDeletedProducer(asyncProducer)

	productService := service.NewProductService(
		providers.DataStore,
		productCreatedProducer,
		productUpdatedProducer,
		productDeletedProducer,
	)
	productHandler := handler.NewProductHandler(productService, appLogger)

	routes.SetupProductRoutes(e, productHandler)
}
