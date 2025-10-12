package entity

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
)

// Notification represents a user-facing notification entity.
type Notification struct {
	ID        uuid.UUID                 `db:"id"`
	UserID    uuid.UUID                 `db:"user_id"`
	Type      constant.NotificationType `db:"type"`
	Title     string                    `db:"title"`
	Message   string                    `db:"message"`
	Metadata  json.RawMessage           `db:"metadata"`
	IsRead    bool                      `db:"is_read"`
	ReadAt    *time.Time                `db:"read_at"`
	CreatedAt time.Time                 `db:"created_at"`
	UpdatedAt time.Time                 `db:"updated_at"`
}

// NewNotification creates a new notification entity.
func NewNotification(
	userID uuid.UUID,
	notificationType constant.NotificationType,
	title string,
	message string,
	metadata map[string]interface{},
) (*Notification, error) {
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}

		metadataJSON = data
	}

	return &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Metadata:  metadataJSON,
		IsRead:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// MarkAsRead marks the notification as read.
func (n *Notification) MarkAsRead() {
	n.IsRead = true
	now := time.Now()
	n.ReadAt = &now
	n.UpdatedAt = now
}

// GetMetadata unmarshals the metadata JSON into a map.
func (n *Notification) GetMetadata() (map[string]interface{}, error) {
	if n.Metadata == nil {
		return nil, errors.New("empty metadata")
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(n.Metadata, &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}
