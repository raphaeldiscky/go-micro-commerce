// Package handler provides HTTP handlers for notification operations.
package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/subscription"
)

// DebugHandler provides debugging endpoints for the notification service.
type DebugHandler struct {
	subscriptionManager *subscription.Manager
	logger              logger.Logger
}

// NewDebugHandler creates a new debug handler.
func NewDebugHandler(
	subscriptionManager *subscription.Manager,
	appLogger logger.Logger,
) *DebugHandler {
	return &DebugHandler{
		subscriptionManager: subscriptionManager,
		logger:              appLogger,
	}
}

// GetActiveSubscriptions returns information about active GraphQL subscriptions.
func (h *DebugHandler) GetActiveSubscriptions(c echo.Context) error {
	subscriptions := h.subscriptionManager.GetAllSubscriptions()

	// Convert to a more readable format
	result := make(map[string]int)
	totalSubscribers := 0

	for userID, count := range subscriptions {
		result[userID.String()] = count
		totalSubscribers += count
	}

	h.logger.Info("Debug: Active subscriptions requested",
		"total_users", len(subscriptions),
		"total_subscribers", totalSubscribers)

	return c.JSON(http.StatusOK, pkgdto.WebResponse[any, any]{
		Data: map[string]any{
			"total_users":       len(subscriptions),
			"total_subscribers": totalSubscribers,
			"subscriptions":     result,
		},
		Message: "active subscriptions retrieved",
	})
}
