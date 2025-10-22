package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
)

// MapToGraphQLNotificationFromDTO maps dto.NotificationResponse to graph.Notification.
func MapToGraphQLNotificationFromDTO(notif *dto.NotificationResponse) *graph.Notification {
	var metadata *string

	if notif.Metadata != nil {
		metadataStr := string(notif.Metadata)
		metadata = &metadataStr
	}

	return &graph.Notification{
		ID:        notif.ID,
		UserID:    notif.UserID,
		Type:      notif.Type,
		Title:     notif.Title,
		Message:   notif.Message,
		Metadata:  metadata,
		IsRead:    notif.IsRead,
		ReadAt:    notif.ReadAt,
		CreatedAt: notif.CreatedAt,
		UpdatedAt: notif.UpdatedAt,
	}
}

// MapToGraphQLNewNotificationFromDTO maps dto.NotificationResponse to graph.NewNotification (for events).
func MapToGraphQLNewNotificationFromDTO(notif *dto.NotificationResponse) *graph.NewNotification {
	var metadata *string

	if notif.Metadata != nil {
		metadataStr := string(notif.Metadata)
		metadata = &metadataStr
	}

	return &graph.NewNotification{
		ID:        notif.ID,
		UserID:    notif.UserID,
		Type:      notif.Type,
		Title:     notif.Title,
		Message:   notif.Message,
		Metadata:  metadata,
		IsRead:    notif.IsRead,
		CreatedAt: notif.CreatedAt,
	}
}
