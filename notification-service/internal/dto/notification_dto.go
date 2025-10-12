package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
)

// NotificationResponse represents the response for notification operations.
type NotificationResponse struct {
	ID        uuid.UUID                 `json:"id"`
	UserID    uuid.UUID                 `json:"user_id"`
	Type      constant.NotificationType `json:"type"`
	Title     string                    `json:"title"`
	Message   string                    `json:"message"`
	Metadata  json.RawMessage           `json:"metadata,omitempty"`
	IsRead    bool                      `json:"is_read"`
	ReadAt    *time.Time                `json:"read_at,omitempty"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

// NotificationListResponse represents a cursor-paginated list of notifications.
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
}

// UnreadCountResponse represents the response for unread notification count.
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

// MarkAsReadRequest represents the request to mark a notification as read.
type MarkAsReadRequest struct {
	NotificationID uuid.UUID `json:"notification_id" validate:"required"`
}
