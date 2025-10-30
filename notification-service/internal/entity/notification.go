package entity

import (
	"errors"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
)

// Notification represents a user-facing notification entity.
type Notification struct {
	ID        uuid.UUID                     `db:"id"`
	UserID    uuid.UUID                     `db:"user_id"`
	Type      constant.PushNotificationType `db:"type"`
	Title     string                        `db:"title"`
	Message   string                        `db:"message"`
	Metadata  sonic.NoCopyRawMessage        `db:"metadata"`
	IsRead    bool                          `db:"is_read"`
	ReadAt    *time.Time                    `db:"read_at"`
	CreatedAt time.Time                     `db:"created_at"`
	UpdatedAt time.Time                     `db:"updated_at"`
}

// NewPushNotification creates a new notification entity.
func NewPushNotification(
	userID uuid.UUID,
	notificationType constant.PushNotificationType,
	title string,
	message string,
	metadata map[string]any,
) (*Notification, error) {
	var metadataJSON sonic.NoCopyRawMessage
	if metadata != nil {
		data, err := sonic.Marshal(metadata)
		if err != nil {
			return nil, err
		}

		metadataJSON = data
	}

	notif := &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Metadata:  metadataJSON,
		IsRead:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := notif.validate(); err != nil {
		return nil, err
	}

	return notif, nil
}

// MarkAsRead marks the notification as read.
func (n *Notification) MarkAsRead() {
	n.IsRead = true
	now := time.Now()
	n.ReadAt = &now
	n.UpdatedAt = now
}

// Validate performs validation on the notification.
func (n *Notification) validate() error {
	switch n.Type {
	case
		constant.PushNotificationTypeNewMessage,
		constant.PushNotificationTypeNewProduct,
		constant.PushNotificationTypeOrderUpdate,
		constant.PushNotificationTypeOrderConfirmed,
		constant.PushNotificationTypeOrderShipped,
		constant.PushNotificationTypeOrderDelivered,
		constant.PushNotificationTypeOrderCancelled,
		constant.PushNotificationTypePaymentSuccess,
		constant.PushNotificationTypePaymentFailed,
		constant.PushNotificationTypePaymentTimeout,
		constant.PushNotificationTypeSystemAlert:
		// valid types → no error
		return nil
	default:
		return errors.New("invalid notification type")
	}
}

// GetMetadata unmarshals the metadata JSON into a map.
func (n *Notification) GetMetadata() (map[string]interface{}, error) {
	if n.Metadata == nil {
		return nil, errors.New("empty metadata")
	}

	var metadata map[string]interface{}
	if err := sonic.Unmarshal(n.Metadata, &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}
