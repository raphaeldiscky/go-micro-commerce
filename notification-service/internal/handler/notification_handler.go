// Package handler provides HTTP handlers for notification operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
)

// NotificationHandler handles HTTP requests for notification operations.
type NotificationHandler struct {
	notificationService      service.NotificationService
	notificationEventService service.NotificationEventService
}

// NewNotificationHandler creates a new instance of NotificationHandler.
func NewNotificationHandler(
	notificationService service.NotificationService,
	notificationEventService service.NotificationEventService,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService:      notificationService,
		notificationEventService: notificationEventService,
	}
}

// ListNotifications handles GET /notifications.
func (h *NotificationHandler) ListNotifications(c echo.Context) error {
	userID := echoutils.GetUserIDFromContext(c)

	limit := pageutils.ParseQueryInt64(
		c,
		"limit",
		pkgconstant.DefaultLimit,
		pkgconstant.DefaultMinLimit,
		pkgconstant.DefaultMaxLimit,
	)

	nextCursor := c.QueryParam("next_cursor")

	notifications, pagination, err := h.notificationService.ListNotifications(
		c.Request().Context(),
		userID,
		limit,
		nextCursor,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOKCursorPagination(c, notifications, pagination)
}

// ListUnreadNotifications handles GET /notifications/unread.
func (h *NotificationHandler) ListUnreadNotifications(c echo.Context) error {
	userID := echoutils.GetUserIDFromContext(c)

	limit := pageutils.ParseQueryInt64(
		c,
		"limit",
		pkgconstant.DefaultLimit,
		pkgconstant.DefaultMinLimit,
		pkgconstant.DefaultMaxLimit,
	)

	nextCursor := c.QueryParam("next_cursor")

	notifications, pagination, err := h.notificationService.ListUnreadNotifications(
		c.Request().Context(),
		userID,
		limit,
		nextCursor,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOKCursorPagination(c, notifications, pagination)
}

// GetNotification handles GET /notifications/:notificationID.
func (h *NotificationHandler) GetNotification(c echo.Context) error {
	notificationIDStr := c.Param("notificationID")

	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return err
	}

	userID := echoutils.GetUserIDFromContext(c)

	notification, err := h.notificationService.GetNotification(
		c.Request().Context(),
		notificationID,
		userID,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, notification)
}

// GetUnreadCount handles GET /notifications/unread/count.
func (h *NotificationHandler) GetUnreadCount(c echo.Context) error {
	userID := echoutils.GetUserIDFromContext(c)

	count, err := h.notificationService.GetUnreadCount(
		c.Request().Context(),
		userID,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, count)
}

// MarkAsRead handles PUT /notifications/:notificationID/read.
func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	notificationIDStr := c.Param("notificationID")

	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return err
	}

	userID := echoutils.GetUserIDFromContext(c)

	err = h.notificationService.MarkAsRead(
		c.Request().Context(),
		notificationID,
		userID,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOKPlain(c)
}

// MarkAllAsRead handles PUT /notifications/read-all.
func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	userID := echoutils.GetUserIDFromContext(c)

	err := h.notificationService.MarkAllAsRead(
		c.Request().Context(),
		userID,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOKPlain(c)
}

// CreateSystemNotification handles POST /notifications.
func (h *NotificationHandler) CreateSystemNotification(c echo.Context) error {
	var req dto.CreateNotificationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	notification, err := h.notificationEventService.CreateAndBroadcastNotification(
		c.Request().Context(),
		req.UserID,
		req.Type,
		req.Title,
		req.Message,
		req.Metadata,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, notification)
}
