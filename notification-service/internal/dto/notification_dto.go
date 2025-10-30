package dto

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
)

// NotificationResponse represents the response for notification operations.
type NotificationResponse struct {
	ID        uuid.UUID                     `json:"id"`
	UserID    uuid.UUID                     `json:"user_id"`
	Type      constant.PushNotificationType `json:"type"`
	Title     string                        `json:"title"`
	Message   string                        `json:"message"`
	Metadata  sonic.NoCopyRawMessage        `json:"metadata,omitempty"`
	IsRead    bool                          `json:"is_read"`
	ReadAt    *time.Time                    `json:"read_at,omitempty"`
	CreatedAt time.Time                     `json:"created_at"`
	UpdatedAt time.Time                     `json:"updated_at"`
}

// UnreadCountResponse represents the response for unread notification count.
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

// MarkAsReadRequest represents the request to mark a notification as read.
type MarkAsReadRequest struct {
	NotificationID uuid.UUID `json:"notification_id" validate:"required"`
}

// CreateNotificationRequest represents the request to create a system notification.
type CreateNotificationRequest struct {
	UserID   *uuid.UUID                    `json:"user_id,omitempty"` // nil = broadcast to all users
	Type     constant.PushNotificationType `json:"type"               validate:"required"`
	Title    string                        `json:"title"              validate:"required,min=1,max=255"`
	Message  string                        `json:"message"            validate:"required,min=1"`
	Metadata map[string]any                `json:"metadata,omitempty"`
}
