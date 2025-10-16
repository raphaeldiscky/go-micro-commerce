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
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)
	// GET /v1 - List notifications with cursor pagination
	protected.GET("", notificationHandler.ListNotifications)
	// GET /v1/unread - List unread notifications with cursor pagination
	protected.GET("/unread", notificationHandler.ListUnreadNotifications)
	// GET /v1/unread/count - Get unread count
	protected.GET("/unread/count", notificationHandler.GetUnreadCount)
	// PUT /v1/read-all - Mark all as read
	protected.PUT("/read-all", notificationHandler.MarkAllAsRead)
	// GET /v1/:notificationID - Get single notification
	protected.GET("/:notificationID", notificationHandler.GetNotification)
	// PUT /v1/:notificationID/read - Mark as read
	protected.PUT("/:notificationID/read", notificationHandler.MarkAsRead)

	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
	// POST /v1 - Create system notification
	admin.POST("", notificationHandler.CreateSystemNotification)
}
