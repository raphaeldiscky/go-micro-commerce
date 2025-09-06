package provider

import (
	"path/filepath"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/smtputils"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/mq/consumer"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
)

// SetupKafkaConsumers initializes the Kafka consumers for the notification service.
func SetupKafkaConsumers(
	cfg *config.KafkaConfig,
	appLogger logger.Logger,
	mailer smtputils.Mailer,
) []kafka.Consumer {
	var consumers []kafka.Consumer

	// Create template service with path to templates directory
	templatesPath := filepath.Join("internal", "templates")
	emailService := service.NewEmailService(templatesPath, mailer)

	userVerificationConsumer, err := kafka.NewConsumer(
		cfg.Brokers,
		constant.TopicUserVerification,
		constant.ConsumerGroupNotificationUserEvents,
		consumer.NewUserVerificationConsumer(emailService, appLogger).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create user verification lifecycle consumer: %v", err)
		// In a real app, you might want to panic here as the service cannot run.
		return nil
	}

	notificationRequestConsumer, err := kafka.NewConsumer(
		cfg.Brokers,
		kafka.NotificationRequestTopic,
		kafka.NotificationServiceConsumerGroup,
		consumer.NewNotificationRequestConsumer(emailService, appLogger).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create notification request consumer: %v", err)
		// In a real app, you might want to panic here as the service cannot run.
		return nil
	}

	consumers = append(consumers, userVerificationConsumer, notificationRequestConsumer)

	appLogger.Infof("successfully created %d Kafka consumers", len(consumers))

	return consumers
}
