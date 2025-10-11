// Package provider provides HTTP client and server utilities.
package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
)

// SetupHTTP initializes the HTTP server routes and middleware.
func SetupHTTP(
	_ *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	// Initialize notification service
	notificationService := service.NewNotificationService(
		providers.DataStore,
		appLogger,
	)

	// Initialize handlers
	notificationHandler := handler.NewNotificationHandler(notificationService)
	sseHandler := handler.NewNotificationSSEHandler(providers.SSEHub, appLogger)
	appHandler := handler.NewAppHandler()

	// Register routes
	routes.SetupAppRoutes(e, appHandler)
	routes.SetupNotificationRoutes(e, notificationHandler, sseHandler)
}
