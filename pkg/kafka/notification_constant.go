package kafka

// Notification service topics.
const (
	NotificationRequestTopic = "notification.request"
)

// Notification service event types.
const (
	NotificationRequestedEventType = "NotificationRequested"
)

// Consumer groups for Notification Service (consuming from other services).
const (
	// ConsumerGroupNotificationUserEvents is the consumer group for user events.
	ConsumerGroupNotificationUserEvents = "notification-service.user-events" // For user lifecycle
	// ConsumerGroupNotificationOrderEvents is the consumer group for order events.
	ConsumerGroupNotificationOrderEvents = "notification-service.order-events" // For order lifecycle
)
