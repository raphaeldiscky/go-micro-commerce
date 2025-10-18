package subscription

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/sse"
)

// Event type constants for SSE and Redis pub/sub.
const (
	TypeNotificationCreated = "notification_created"
	TypeNotificationRead    = "notification_read"
	TypeNotificationDeleted = "notification_deleted"
)

// NotificationCreatedEvent represents a new notification event for cross-instance delivery.
type NotificationCreatedEvent struct {
	UserID  uuid.UUID    `json:"user_id"`
	Message *sse.Message `json:"message"`
}

// NotificationReadEvent represents a notification read event.
type NotificationReadEvent struct {
	UserID         uuid.UUID `json:"user_id"`
	NotificationID uuid.UUID `json:"notification_id"`
}

// NotificationDeletedEvent represents a notification deleted event.
type NotificationDeletedEvent struct {
	UserID         uuid.UUID `json:"user_id"`
	NotificationID uuid.UUID `json:"notification_id"`
}
