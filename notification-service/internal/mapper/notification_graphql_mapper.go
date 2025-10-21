package mapper

import (
	"encoding/json"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/entity"
)

// MapToGraphQLNotification maps entity.Notification to graph.Notification.
func MapToGraphQLNotification(notif *entity.Notification) *graph.Notification {
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

// MapToGraphQLNewNotification maps entity.Notification to graph.NewNotification (for events).
func MapToGraphQLNewNotification(notif *entity.Notification) *graph.NewNotification {
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

// MapToGraphQLNotificationConnection maps notification list to GraphQL connection.
func MapToGraphQLNotificationConnection(
	notifications []*entity.Notification,
	nextCursor string,
	hasNextPage bool,
) *graph.NotificationConnection {
	edges := make([]*graph.NotificationEdge, len(notifications))

	for i, notif := range notifications {
		// Generate cursor from notification timestamp and ID
		cursorData := map[string]interface{}{
			"id":        notif.ID,
			"timestamp": notif.CreatedAt.Unix(),
		}

		cursorJSON, err := json.Marshal(cursorData)
		if err != nil {
			continue
		}

		edges[i] = &graph.NotificationEdge{
			Node:   MapToGraphQLNotification(notif),
			Cursor: string(cursorJSON),
		}
	}

	var endCursor *string

	if nextCursor != "" {
		endCursor = &nextCursor
	}

	return &graph.NotificationConnection{
		Edges: edges,
		PageInfo: &graph.PageInfo{
			HasNextPage: hasNextPage,
			EndCursor:   endCursor,
		},
	}
}
