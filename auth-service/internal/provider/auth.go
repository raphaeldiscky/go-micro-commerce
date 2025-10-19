package provider

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/service"
)

// SetupAuth initializes the authentication-related components.
func SetupAuth(
	ctx context.Context,
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	err := providers.KafkaAdmin.CreateTopic(
		kafka.UserVerificationTopic,
		constant.UserVerificationTopicNumPartitions,
		constant.UserVerificationTopicReplicationFactor,
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

	emailVerificationRequestedProducer := mq.NewEmailVerificationRequestedProducer(asyncProducer)
	userVerifiedProducer := mq.NewUserVerifiedProducer(asyncProducer)

	authService := service.NewAuthService(
		providers.DataStore,
		providers.JWTUtils,
		providers.BcryptHasher,
		cfg.Auth,
		appLogger,
		emailVerificationRequestedProducer,
		userVerifiedProducer,
	)
	addressService := service.NewAddressService(
		providers.DataStore,
		appLogger,
	)
	authHandler := handler.NewAuthHandler(authService)
	jwksHandler := handler.NewJWKSHandler(providers.JWTUtils)

	// Setup REST routes
	routes.SetupAuthRoutes(e, authHandler, jwksHandler)

	// Setup GraphQL routes
	graphqlResolver := resolver.NewResolver(authService, addressService, appLogger)
	routes.SetupGraphQLRoutes(e, cfg, graphqlResolver, appLogger)
}
