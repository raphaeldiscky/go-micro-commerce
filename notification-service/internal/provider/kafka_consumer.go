package provider

import (
	"path/filepath"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/smtputils"

	pkgconfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/mq/consumer"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/worker"
)

// SetupKafkaConsumers initializes the Kafka consumers for the notification service.
func SetupKafkaConsumers(
	cfg *config.Config,
	appLogger logger.Logger,
) *worker.KafkaConsumer {
	var consumers []kafka.Consumer

	mailer := smtputils.NewMailer(&pkgconfig.SMTPConfig{
		Host:  cfg.SMTP.Host,
		Email: cfg.SMTP.Email,
		Port:  cfg.SMTP.Port,
	})
	// Create template service with path to templates directory
	templatesPath := filepath.Join("internal", "template")
	emailService := service.NewEmailService(templatesPath, mailer)

	userVerificationConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
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
		cfg.Kafka.Brokers,
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

	return worker.NewKafkaConsumer(cfg, appLogger, consumers)
}
