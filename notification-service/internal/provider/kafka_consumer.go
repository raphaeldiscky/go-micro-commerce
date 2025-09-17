package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/mq/consumer"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/worker"
)

// SetupKafkaConsumers initializes the Kafka consumers for the notification service.
func SetupKafkaConsumers(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *worker.KafkaConsumer {
	var consumers []kafka.Consumer

	userVerificationConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		kafka.UserVerificationTopic,
		kafka.ConsumerGroupNotificationUserEvents,
		consumer.NewUserVerificationConsumer(providers.DataStore, appLogger).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create user verification lifecycle consumer: %v", err)
		// In a real app, you might want to panic here as the service cannot run.
		return nil
	}

	notificationRequestConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		kafka.NotificationRequestTopic,
		kafka.ConsumerGroupNotificationOrderEvents,
		consumer.NewNotificationRequestConsumer(providers.DataStore, appLogger).Handler,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("failed to create notification request consumer: %v", err)
		// In a real app, you might want to panic here as the service cannot run.
		return nil
	}

	consumers = append(consumers, userVerificationConsumer, notificationRequestConsumer)

	appLogger.Infof("successfully created %d Kafka consumers", len(consumers))

	return worker.NewKafkaConsumer(appLogger, consumers)
}
