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

	userVerificationConsumer, err := mq.NewConsumerKafka(
		cfg.Brokers,
		constant.UserVerificationTopic,
		constant.ConsumerGroupNotificationUserEvents,
		event.NewUserVerificationConsumer(mailer, appLogger).Handler,
	)
	if err != nil {
		appLogger.Errorf("Failed to create user verification lifecycle consumer: %v", err)
		// In a real app, you might want to panic here as the service cannot run.
		return nil
	}

	consumers = append(consumers, userVerificationConsumer)

	// --- Add more consumers for different topics e.g. user.security here following the same pattern ---
	// example:
	// passwordResetHandler := event.NewPasswordResetConsumer(mailer)
	// passwordResetConsumer, err := mq.NewConsumerKafka(...)
	// consumers = append(consumers, passwordResetConsumer)

	appLogger.Infof("Successfully created %d Kafka consumers", len(consumers))

	return consumers
}
