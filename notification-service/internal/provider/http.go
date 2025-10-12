// Package provider provides HTTP client and server utilities.
package provider

import (
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/smtputils"

	pkgconfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
)

// SetupHTTP initializes the HTTP server routes and middleware.
func SetupHTTP(
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	// Initialize notification service
	notificationService := service.NewNotificationService(
		providers.DataStore,
		appLogger,
	)

	// Initialize email service for notification event service
	mailer := smtputils.NewMailer(&pkgconfig.SMTPConfig{
		Host:  cfg.SMTP.Host,
		Email: cfg.SMTP.Email,
		Port:  cfg.SMTP.Port,
	})
	templatesPath := filepath.Join("internal", "template")
	emailService := service.NewEmailService(templatesPath, mailer)

	// Initialize notification event service
	notificationEventService := service.NewNotificationEventService(
		emailService,
		providers.DataStore.NotificationRepository(),
		providers.SSEHub,
		providers.EventBus,
		providers.InstanceID,
		providers.Sharder,
		appLogger,
	)

	// Initialize handlers
	notificationHandler := handler.NewNotificationHandler(
		notificationService,
		notificationEventService,
	)
	sseHandler := handler.NewNotificationSSEHandler(providers.SSEHub, appLogger)
	appHandler := handler.NewAppHandler()

	// Initialize GraphQL resolver
	graphResolver := resolver.NewResolver(
		notificationService,
		providers.SubscriptionManager,
		appLogger,
	)

	// Register routes
	routes.SetupAppRoutes(e, appHandler)
	routes.SetupNotificationRoutes(e, notificationHandler, sseHandler)
	routes.SetupGraphQLRoutes(e, cfg, graphResolver, appLogger)
}
