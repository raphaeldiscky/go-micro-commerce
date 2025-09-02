package kafka

const (
	// OrderLifecycleTopic is the topic for order lifecycle events.
	OrderLifecycleTopic = "order.lifecycle"
)

const (
	// OrderCreatedEventType is when customer places an order (pending).
	// Needed by: inventory reservation, payment service, fraud detection.
	OrderCreatedEventType = "OrderCreated"
	// OrderProcessingEventType is when order is being processed (processing).
	// Needed by: shipping service, accounting, notification service.
	OrderProcessingEventType = "OrderProcessing"
	// OrderConfirmedEventType is when order is confirmed (confirmed).
	// Needed by: inventory service to update stock, notification service.
	OrderConfirmedEventType = "OrderConfirmed"
	// OrderPaidEventType is when payment succeeded (paid).
	// Needed by: shipping service, accounting, notification service.
	OrderPaidEventType = "OrderPaid"
	// OrderShippedEventType is when order handed to logistics (shipped).
	// Needed by: notification service, delivery tracking.
	OrderShippedEventType = "OrderShipped"
	// OrderDeliveredEventType is after customer received package (delivered).
	// Needed by: loyalty points, analytics.
	OrderDeliveredEventType = "OrderDelivered"
	// OrderCanceledEventType is when canceled by system or customer (canceled)
	// Needed by: inventory release, refund, analytics.
	OrderCanceledEventType = "OrderCanceled"
	// OrderFailedEventType is when order processing failed (failed).
	// Needed by: notification service, analytics.
	OrderFailedEventType = "OrderFailed"
)

const (
	// OrderProductEventsConsumerGroup is the consumer group for order product events.
	OrderProductEventsConsumerGroup = "order-service.product-events"
)
