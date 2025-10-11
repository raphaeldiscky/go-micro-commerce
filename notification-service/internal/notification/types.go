// Package notification provides notification-specific types for cross-instance messaging.
package notification

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/sse"
)

// Event type constants for SSE and Redis pub/sub.
const (
	TypeNotificationCreated = "notification_created"
	TypeNotificationRead    = "notification_read"
	TypeNotificationDeleted = "notification_deleted"
)

// CreatedEvent represents a new notification event for cross-instance delivery.
type CreatedEvent struct {
	UserID  uuid.UUID    `json:"user_id"`
	Message *sse.Message `json:"message"`
}

// ReadEvent represents a notification read event.
type ReadEvent struct {
	UserID         uuid.UUID `json:"user_id"`
	NotificationID uuid.UUID `json:"notification_id"`
}

// DeletedEvent represents a notification deleted event.
type DeletedEvent struct {
	UserID         uuid.UUID `json:"user_id"`
	NotificationID uuid.UUID `json:"notification_id"`
}

// MarshalEvent marshals an event payload to JSON.
func MarshalEvent(payload any) ([]byte, error) {
	return json.Marshal(payload)
}

// UnmarshalCreatedEvent unmarshals a notification created event.
func UnmarshalCreatedEvent(data []byte) (*CreatedEvent, error) {
	var event CreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// UnmarshalReadEvent unmarshals a notification read event.
func UnmarshalReadEvent(data []byte) (*ReadEvent, error) {
	var event ReadEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// UnmarshalDeletedEvent unmarshals a notification deleted event.
func UnmarshalDeletedEvent(data []byte) (*DeletedEvent, error) {
	var event DeletedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return &event, nil
}
