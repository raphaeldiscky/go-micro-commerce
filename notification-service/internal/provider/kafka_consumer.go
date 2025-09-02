package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/smtputils"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/event"
)

// SetupKafkaConsumers initializes the Kafka consumers for the notification service.
func SetupKafkaConsumers(
	cfg *config.KafkaConfig,
	appLogger logger.Logger,
	mailer smtputils.Mailer,
) []kafka.Consumer {
	var consumers []kafka.Consumer

	userVerificationConsumer, err := kafka.NewConsumer(
		cfg.Brokers,
		constant.TopicUserVerification,
		constant.ConsumerGroupNotificationUserEvents,
		event.NewUserVerificationConsumer(mailer, appLogger).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create user verification lifecycle consumer: %v", err)
		// In a real app, you might want to panic here as the service cannot run.
		return nil
	}

	consumers = append(consumers, userVerificationConsumer)

	// --- Add more consumers for different topics e.g. user.security here following the same pattern ---
	// example:
	// passwordResetHandler := event.NewPasswordResetConsumer(mailer)
	// passwordResetConsumer, err := kafka.NewConsumer(...)
	// consumers = append(consumers, passwordResetConsumer)

	appLogger.Infof("successfully created %d Kafka consumers", len(consumers))

	return consumers
}
