// Package handler provides HTTP handlers for notification operations.
package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/sse"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
)

// NotificationSSEHandler handles SSE connections for real-time notifications.
type NotificationSSEHandler struct {
	hub    *sse.Hub
	logger logger.Logger
}

// NewNotificationSSEHandler creates a new instance of NotificationSSEHandler.
func NewNotificationSSEHandler(
	hub *sse.Hub,
	appLogger logger.Logger,
) *NotificationSSEHandler {
	return &NotificationSSEHandler{
		hub:    hub,
		logger: appLogger,
	}
}

// StreamNotifications handles GET /notifications/stream.
// Establishes an SSE connection for the authenticated user.
func (h *NotificationSSEHandler) StreamNotifications(c echo.Context) error {
	userID := echoutils.GetUserIDFromContext(c)

	// Create SSE connection
	conn := sse.NewConnection(userID, c, h.logger)

	// Register connection with hub
	h.hub.Register(conn)

	// Ensure unregister on disconnect
	defer h.hub.Unregister(conn)

	h.logger.Info("User connected to notification stream",
		"user_id", userID,
		"connection_id", conn.ID())

	// Start write pump (blocks until connection closes)
	conn.WritePump(c.Request().Context())

	h.logger.Info("User disconnected from notification stream",
		"user_id", userID,
		"connection_id", conn.ID())

	return nil
}
