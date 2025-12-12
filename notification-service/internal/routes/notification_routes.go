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
) {
	protected := e.Group("")
	protected.Use(middleware.AuthMiddleware)
	// GET / - List notifications with cursor pagination
	protected.GET("", notificationHandler.ListNotifications)
	// GET /unread - List unread notifications with cursor pagination
	protected.GET("/unread", notificationHandler.ListUnreadNotifications)
	// GET /unread/count - Get unread count
	protected.GET("/unread/count", notificationHandler.GetUnreadCount)
	// PUT /read-all - Mark all as read
	protected.PUT("/read-all", notificationHandler.MarkAllAsRead)
	// GET /:notificationID - Get single notification
	protected.GET("/:notificationID", notificationHandler.GetNotification)
	// PUT /:notificationID/read - Mark as read
	protected.PUT("/:notificationID/read", notificationHandler.MarkAsRead)

	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
	// POST / - Create system notification
	admin.POST("", notificationHandler.CreateSystemNotification)
}
