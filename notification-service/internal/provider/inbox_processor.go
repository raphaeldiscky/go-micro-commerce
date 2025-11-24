package provider

import (
	"path/filepath"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/mailutils"

	pkgconfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/worker"
)

// SetupInboxProcessor initializes the inbox processor service.
func SetupInboxProcessor(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *worker.InboxProcessor {
	mailer, err := mailutils.NewMailer(&pkgconfig.MailConfig{
		Provider:       cfg.Mail.Provider,
		Host:           cfg.Mail.Host,
		FromEmail:      cfg.Mail.FromEmail,
		Port:           cfg.Mail.Port,
		SendGridAPIKey: cfg.Mail.SendGridAPIKey,
	})
	if err != nil {
		appLogger.Fatal("failed to create mailer", "error", err)
	}
	// Create template service with path to templates directory
	templatesPath := filepath.Join("internal", "template")
	emailService := service.NewEmailService(templatesPath, mailer)

	// Create notification event service with all dependencies
	notificationEventService := service.NewNotificationEventService(
		emailService,
		providers.DataStore.NotificationRepository(),
		providers.SSEHub,
		providers.EventBus,
		providers.InstanceID,
		appLogger,
	)

	// Create inbox processor
	inboxProcessor := worker.NewInboxProcessor(
		providers.DataStore,
		appLogger,
		notificationEventService,
		*cfg.InboxProcessor,
	)

	return inboxProcessor
}
