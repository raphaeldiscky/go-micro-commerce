package constant

// NotificationType represents the type of notification.
type NotificationType string

// Notification type constants.
const (
	NotificationTypeNewMessage     NotificationType = "new_message"
	NotificationTypeNewProduct     NotificationType = "new_product"
	NotificationTypeOrderUpdate    NotificationType = "order_update"
	NotificationTypeOrderConfirmed NotificationType = "order_confirmed"
	NotificationTypeOrderShipped   NotificationType = "order_shipped"
	NotificationTypeOrderDelivered NotificationType = "order_delivered"
	NotificationTypeOrderCancelled NotificationType = "order_cancelled"
	NotificationTypePaymentSuccess NotificationType = "payment_success"
	NotificationTypeSystemAlert    NotificationType = "system_alert"
)
