// Package service provides business logic for notification operations.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/repository"
)

// NotificationService defines the interface for notification business operations.
type NotificationService interface {
	// ListNotifications retrieves notifications for a user with cursor pagination
	ListNotifications(
		ctx context.Context,
		userID uuid.UUID,
		limit int64,
		cursor string,
	) (*dto.NotificationListResponse, *pkgdto.CursorPagination, error)

	// ListUnreadNotifications retrieves unread notifications for a user with cursor pagination
	ListUnreadNotifications(
		ctx context.Context,
		userID uuid.UUID,
		limit int64,
		cursor string,
	) (*dto.NotificationListResponse, *pkgdto.CursorPagination, error)

	// GetNotification retrieves a notification by ID
	GetNotification(
		ctx context.Context,
		notificationID uuid.UUID,
		userID uuid.UUID,
	) (*dto.NotificationResponse, error)

	// GetUnreadCount retrieves the count of unread notifications for a user
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (*dto.UnreadCountResponse, error)

	// MarkAsRead marks a notification as read
	MarkAsRead(
		ctx context.Context,
		notificationID uuid.UUID,
		userID uuid.UUID,
	) error

	// MarkAllAsRead marks all notifications as read for a user
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error

	// DeleteNotification deletes a notification
	DeleteNotification(
		ctx context.Context,
		notificationID uuid.UUID,
		userID uuid.UUID,
	) error

	// DeleteAllNotifications deletes all notifications for a user
	DeleteAllNotifications(ctx context.Context, userID uuid.UUID) error

	// GetTotalCount retrieves the total count of notifications for a user
	GetTotalCount(ctx context.Context, userID uuid.UUID) (int64, error)
}

// notificationService implements the NotificationService interface.
type notificationService struct {
	dataStore repository.DataStore
	logger    logger.Logger
}

// NewNotificationService creates a new instance of notificationService.
func NewNotificationService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
) NotificationService {
	return &notificationService{
		dataStore: dataStore,
		logger:    appLogger,
	}
}

// ListNotifications retrieves notifications for a user with cursor pagination.
func (s *notificationService) ListNotifications(
	ctx context.Context,
	userID uuid.UUID,
	limit int64,
	cursor string,
) (*dto.NotificationListResponse, *pkgdto.CursorPagination, error) {
	notificationRepo := s.dataStore.NotificationRepository()

	var (
		cursorID        string
		cursorTimestamp int64
	)

	if cursor != "" {
		cursorData, err := pageutils.DecodeCursor(cursor)
		if err != nil {
			return nil, nil, httperror.NewBadRequestError("invalid cursor")
		}

		cursorID = cursorData.ID
		cursorTimestamp = cursorData.Timestamp
	}

	fetchLimit := limit + 1

	notifications, err := notificationRepo.FindByUserIDWithCursor(
		ctx,
		userID,
		fetchLimit,
		cursorID,
		cursorTimestamp,
	)
	if err != nil {
		s.logger.Error("Failed to find notifications with cursor",
			"user_id", userID,
			"error", err)

		return nil, nil, httperror.NewInternalServerError("failed to get notifications")
	}

	hasNext := len(notifications) > int(limit)
	if hasNext {
		notifications = notifications[:limit]
	}

	listResponse := mapper.MapToNotificationListResponse(notifications)

	var nextCursor string

	if hasNext && len(notifications) > 0 {
		lastNotif := notifications[len(notifications)-1]

		nextCursor, err = pageutils.GenerateNextCursor(
			lastNotif.ID.String(),
			lastNotif.CreatedAt.Unix(),
			"",
		)
		if err != nil {
			return nil, nil, httperror.NewInternalServerError("failed to generate cursor")
		}
	}

	pagination := pageutils.NewCursorPagination(nextCursor, "", hasNext, false, limit)

	return listResponse, pagination, nil
}

// ListUnreadNotifications retrieves unread notifications for a user with cursor pagination.
func (s *notificationService) ListUnreadNotifications(
	ctx context.Context,
	userID uuid.UUID,
	limit int64,
	cursor string,
) (*dto.NotificationListResponse, *pkgdto.CursorPagination, error) {
	notificationRepo := s.dataStore.NotificationRepository()

	var (
		cursorID        string
		cursorTimestamp int64
	)

	if cursor != "" {
		cursorData, err := pageutils.DecodeCursor(cursor)
		if err != nil {
			return nil, nil, httperror.NewBadRequestError("invalid cursor")
		}

		cursorID = cursorData.ID
		cursorTimestamp = cursorData.Timestamp
	}

	fetchLimit := limit + 1

	notifications, err := notificationRepo.FindUnreadByUserIDWithCursor(
		ctx,
		userID,
		fetchLimit,
		cursorID,
		cursorTimestamp,
	)
	if err != nil {
		s.logger.Error("Failed to find unread notifications with cursor",
			"user_id", userID,
			"error", err)

		return nil, nil, httperror.NewInternalServerError("failed to get unread notifications")
	}

	hasNext := len(notifications) > int(limit)
	if hasNext {
		notifications = notifications[:limit]
	}

	listResponse := mapper.MapToNotificationListResponse(notifications)

	var nextCursor string

	if hasNext && len(notifications) > 0 {
		lastNotif := notifications[len(notifications)-1]

		nextCursor, err = pageutils.GenerateNextCursor(
			lastNotif.ID.String(),
			lastNotif.CreatedAt.Unix(),
			"",
		)
		if err != nil {
			return nil, nil, httperror.NewInternalServerError("failed to generate cursor")
		}
	}

	pagination := pageutils.NewCursorPagination(nextCursor, "", hasNext, false, limit)

	return listResponse, pagination, nil
}

// GetNotification retrieves a notification by ID.
func (s *notificationService) GetNotification(
	ctx context.Context,
	notificationID uuid.UUID,
	userID uuid.UUID,
) (*dto.NotificationResponse, error) {
	notificationRepo := s.dataStore.NotificationRepository()

	notification, err := notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		s.logger.Error("Failed to find notification",
			"notification_id", notificationID,
			"error", err)

		return nil, httperror.NewInternalServerError("failed to get notification")
	}

	// Verify the notification belongs to the user
	if notification.UserID != userID {
		return nil, httperror.NewForbiddenError("access denied")
	}

	return mapper.MapToNotificationResponse(notification), nil
}

// GetUnreadCount retrieves the count of unread notifications for a user.
func (s *notificationService) GetUnreadCount(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.UnreadCountResponse, error) {
	notificationRepo := s.dataStore.NotificationRepository()

	count, err := notificationRepo.CountUnreadByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to count unread notifications",
			"user_id", userID,
			"error", err)

		return nil, httperror.NewInternalServerError("failed to count unread notifications")
	}

	return &dto.UnreadCountResponse{
		Count: count,
	}, nil
}

// MarkAsRead marks a notification as read.
func (s *notificationService) MarkAsRead(
	ctx context.Context,
	notificationID uuid.UUID,
	userID uuid.UUID,
) error {
	notificationRepo := s.dataStore.NotificationRepository()

	err := notificationRepo.MarkAsRead(ctx, notificationID, userID)
	if err != nil {
		s.logger.Error("Failed to mark notification as read",
			"notification_id", notificationID,
			"user_id", userID,
			"error", err)

		return httperror.NewInternalServerError("failed to mark notification as read")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user.
func (s *notificationService) MarkAllAsRead(
	ctx context.Context,
	userID uuid.UUID,
) error {
	notificationRepo := s.dataStore.NotificationRepository()

	err := notificationRepo.MarkAllAsRead(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to mark all notifications as read",
			"user_id", userID,
			"error", err)

		return httperror.NewInternalServerError("failed to mark all notifications as read")
	}

	return nil
}

// DeleteNotification deletes a notification.
func (s *notificationService) DeleteNotification(
	ctx context.Context,
	notificationID uuid.UUID,
	userID uuid.UUID,
) error {
	notificationRepo := s.dataStore.NotificationRepository()

	err := notificationRepo.Delete(ctx, notificationID, userID)
	if err != nil {
		s.logger.Error("Failed to delete notification",
			"notification_id", notificationID,
			"user_id", userID,
			"error", err)

		return httperror.NewInternalServerError("failed to delete notification")
	}

	return nil
}

// DeleteAllNotifications deletes all notifications for a user.
func (s *notificationService) DeleteAllNotifications(
	ctx context.Context,
	userID uuid.UUID,
) error {
	notificationRepo := s.dataStore.NotificationRepository()

	err := notificationRepo.DeleteAllByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to delete all notifications",
			"user_id", userID,
			"error", err)

		return httperror.NewInternalServerError("failed to delete all notifications")
	}

	return nil
}

// GetTotalCount retrieves the total count of notifications for a user.
func (s *notificationService) GetTotalCount(
	ctx context.Context,
	userID uuid.UUID,
) (int64, error) {
	notificationRepo := s.dataStore.NotificationRepository()

	count, err := notificationRepo.CountByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to count total notifications",
			"user_id", userID,
			"error", err)

		return 0, httperror.NewInternalServerError("failed to count notifications")
	}

	return count, nil
}
