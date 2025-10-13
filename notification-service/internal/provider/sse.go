package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
)

// SetupSSE initializes the SSE server routes for GraphQL subscriptions.
func SetupSSE(
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	notifificationService := service.NewNotificationService(
		providers.DataStore,
		appLogger,
	)
	// Initialize GraphQL resolver (reuse from HTTP setup)
	graphResolver := resolver.NewResolver(
		notifificationService,
		providers.SubscriptionManager,
		appLogger,
	)

	sseHandler := handler.NewNotificationSSEHandler(providers.SSEHub, appLogger)

	// Register only GraphQL SSE routes
	routes.SetupSSERoutes(e, sseHandler)
	routes.SetupGraphQLSSERoutes(e, cfg, graphResolver, appLogger)
}
