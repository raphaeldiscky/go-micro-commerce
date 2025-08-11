package provider

import (
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/smtputils"

	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/event"
)

// SetupKafkaConsumers initializes the Kafka consumers for the notification service.
func SetupKafkaConsumers(
	cfg *config.KafkaConfig,
	appLogger logger.Logger,
	mailer smtputils.Mailer,
) []mq.KafkaConsumer {
	var consumers []mq.KafkaConsumer

	// 1. Create the business logic handler for email verification
	emailVerificationHandler := event.NewEmailVerificationConsumer(mailer)

	// 2. Create the generic Kafka consumer, injecting the business logic handler
	emailConsumer, err := mq.NewConsumerKafka(
		cfg.Brokers,
		constant.UserVerificationTopic,
		constant.ConsumerGroupNotificationUserEvents,
		emailVerificationHandler.Handler,
	)
	if err != nil {
		appLogger.Errorf("Failed to create email verification consumer: %v", err)
		// In a real app, you might want to panic here as the service cannot run.
		return nil
	}

	consumers = append(consumers, emailConsumer)

	// --- Add more consumers here following the same pattern ---
	// example:
	// passwordResetHandler := event.NewPasswordResetConsumer(mailer)
	// passwordResetConsumer, err := mq.NewConsumerKafka(...)
	// consumers = append(consumers, passwordResetConsumer)

	appLogger.Infof("Successfully created %d Kafka consumers", len(consumers))

	return consumers
}
