package provider

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/service"
)

// SetupProduct initializes the product-related routes and services.
func SetupProduct(
	ctx context.Context,
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	// If ProductService is not initialized, initialize it
	if providers.ProductService == nil {
		InitializeProductService(ctx, cfg, appLogger, providers)
	}

	// Set up HTTP routes
	productHandler := handler.NewProductHandler(providers.ProductService, appLogger)
	routes.SetupProductRoutes(e, productHandler)
}

// InitializeProductService initializes only the ProductService without HTTP routes.
// This is used to ensure ProductService is available for gRPC server without race conditions.
func InitializeProductService(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) {
	err := providers.KafkaAdmin.CreateTopic(
		kafka.ProductLifecycleTopic,
		constant.ProductLifecycleTopicNumPartitions,
		constant.ProductLifecycleTopicReplicationFactor,
	)
	if err != nil {
		appLogger.Fatalf("failed to create Kafka topic: %v", err)
	}

	asyncProducer, err := kafka.NewAsyncProducer(ctx, &kafka.ProducerConfig{
		Brokers:        cfg.Kafka.Brokers,
		RetryMax:       cfg.Kafka.RetryMax,
		RetryTicker:    cfg.Kafka.RetryTicker,
		FlushFrequency: cfg.Kafka.FlushFrequency,
		ReturnSuccess:  cfg.Kafka.ReturnSuccess,
		ReturnErrors:   cfg.Kafka.ReturnErrors,
		Acks:           sarama.WaitForAll,
	})
	if err != nil {
		appLogger.Fatalf("failed to create Kafka async producer: %v", err)
	}

	productCreatedProducer := mq.NewProductCreatedProducer(asyncProducer)
	productUpdatedProducer := mq.NewProductUpdatedProducer(asyncProducer)
	productDeletedProducer := mq.NewProductDeletedProducer(asyncProducer)

	productService := service.NewProductService(
		providers.DataStore,
		appLogger,
		productCreatedProducer,
		productUpdatedProducer,
		productDeletedProducer,
	)

	providers.ProductService = productService
}
