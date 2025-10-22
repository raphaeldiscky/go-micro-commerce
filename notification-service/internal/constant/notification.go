package constant

// PushNotificationType represents the type of notification.
//
//nolint:recvcheck // Mixed receivers required: Unmarshal uses pointer, Marshal uses value.
type PushNotificationType string

// Notification type constants.
const (
	PushNotificationTypeNewMessage     PushNotificationType = "new_message"
	PushNotificationTypeNewProduct     PushNotificationType = "new_product"
	PushNotificationTypeOrderUpdate    PushNotificationType = "order_update"
	PushNotificationTypeOrderConfirmed PushNotificationType = "order_confirmed"
	PushNotificationTypeOrderShipped   PushNotificationType = "order_shipped"
	PushNotificationTypeOrderDelivered PushNotificationType = "order_delivered"
	PushNotificationTypeOrderCancelled PushNotificationType = "order_cancelled"
	PushNotificationTypePaymentSuccess PushNotificationType = "payment_success"
	PushNotificationTypeSystemAlert    PushNotificationType = "system_alert"
)

// IsValid checks if the notification type is valid.
func (e PushNotificationType) IsValid() bool {
	switch e {
	case PushNotificationTypeNewMessage,
		PushNotificationTypeNewProduct,
		PushNotificationTypeOrderUpdate,
		PushNotificationTypeOrderConfirmed,
		PushNotificationTypeOrderShipped,
		PushNotificationTypeOrderDelivered,
		PushNotificationTypeOrderCancelled,
		PushNotificationTypePaymentSuccess,
		PushNotificationTypeSystemAlert:
		return true
	}

	return false
}

// String returns the string representation (lowercase with underscore).
func (e PushNotificationType) String() string {
	return string(e)
}
