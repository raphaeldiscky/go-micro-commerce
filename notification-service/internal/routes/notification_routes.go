// Package routes provides the HTTP routes for the notification service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/middleware"
)

// SetupNotificationRoutes sets up all notification routes.
func SetupNotificationRoutes(
	e *echo.Echo,
	notificationHandler *handler.NotificationHandler,
	sseHandler *handler.NotificationSSEHandler,
) {
	v1 := e.Group("/v1")

	// SSE stream endpoint (requires authentication)
	v1.GET("/notifications/stream", sseHandler.StreamNotifications, middleware.AuthMiddleware)

	// REST API endpoints (requires authentication)
	api := v1.Group("/notifications", middleware.AuthMiddleware)
	{
		// GET /api/v1/notifications - List notifications with cursor pagination
		api.GET("", notificationHandler.ListNotifications)

		// GET /api/v1/notifications/unread - List unread notifications with cursor pagination
		api.GET("/unread", notificationHandler.ListUnreadNotifications)

		// GET /api/v1/notifications/unread/count - Get unread count
		api.GET("/unread/count", notificationHandler.GetUnreadCount)

		// PUT /api/v1/notifications/read-all - Mark all as read
		api.PUT("/read-all", notificationHandler.MarkAllAsRead)

		// DELETE /api/v1/notifications/all - Delete all notifications
		api.DELETE("/all", notificationHandler.DeleteAllNotifications)

		// GET /api/v1/notifications/:notificationID - Get single notification
		api.GET("/:notificationID", notificationHandler.GetNotification)

		// PUT /api/v1/notifications/:notificationID/read - Mark as read
		api.PUT("/:notificationID/read", notificationHandler.MarkAsRead)

		// DELETE /api/v1/notifications/:notificationID - Delete notification
		api.DELETE("/:notificationID", notificationHandler.DeleteNotification)
	}
}
