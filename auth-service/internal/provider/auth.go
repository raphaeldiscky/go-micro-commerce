package provider

import (
	"log"

	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/service"
)

// SetupAuth initializes the authentication-related components.
func SetupAuth(cfg *config.Config, e *echo.Echo, appLogger logger.Logger, providers *Providers) {
	providers.KafkaAdmin.CreateTopic(
		kafka.UserVerificationTopic,
		constant.UserVerificationTopicNumPartitions,
		constant.UserVerificationTopicReplicationFactor,
	)

	asyncProducer, err := kafka.NewAsyncProducer(&kafka.ProducerConfig{
		Brokers:        cfg.Kafka.Brokers,
		RetryMax:       cfg.Kafka.RetryMax,
		FlushFrequency: cfg.Kafka.FlushFrequency,
		ReturnSuccess:  cfg.Kafka.ReturnSuccess,
		ReturnErrors:   true, // Enable error returns for better error handling
		Acks:           sarama.WaitForAll,
	})
	if err != nil {
		log.Fatalf("failed to create Kafka async producer: %v", err)
	}

	emailVerificationRequestedProducer := mq.NewEmailVerificationRequestedProducer(asyncProducer)
	userVerifiedProducer := mq.NewUserVerifiedProducer(asyncProducer)

	authService := service.NewAuthService(
		providers.DataStore,
		providers.JWTUtils,
		providers.BcryptHasher,
		appLogger,
		emailVerificationRequestedProducer,
		userVerifiedProducer,
	)
	authHandler := handler.NewAuthHandler(authService, appLogger)

	routes.SetupAuthRoutes(e, authHandler)
}
