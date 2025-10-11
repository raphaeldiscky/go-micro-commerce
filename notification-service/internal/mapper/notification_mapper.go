package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/entity"
)

// MapToNotificationResponse converts a notification entity to a response DTO.
func MapToNotificationResponse(notification *entity.Notification) *dto.NotificationResponse {
	if notification == nil {
		return nil
	}

	return &dto.NotificationResponse{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		Metadata:  notification.Metadata,
		IsRead:    notification.IsRead,
		ReadAt:    notification.ReadAt,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}
}

// MapToNotificationListResponse converts a slice of notification entities to a list response DTO.
func MapToNotificationListResponse(
	notifications []*entity.Notification,
) *dto.NotificationListResponse {
	notificationResponses := make([]dto.NotificationResponse, 0, len(notifications))

	for _, notification := range notifications {
		if resp := MapToNotificationResponse(notification); resp != nil {
			notificationResponses = append(notificationResponses, *resp)
		}
	}

	return &dto.NotificationListResponse{
		Notifications: notificationResponses,
	}
}
