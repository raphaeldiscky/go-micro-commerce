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
	// OrderPaymentPendingEventType is when order payment is pending and need to be paid (payment_pending).
	OrderPaymentPendingEventType = "OrderPaymentPending"
	// OrderPaymentExpiredEventType is when order payment expired (payment_expired).
	// Needed by: inventory release, no refund needed, analytics.
	OrderPaymentExpiredEventType = "OrderPaymentExpired"
	// OrderFailedEventType is when order processing failed (failed).
	// Needed by: notification service, analytics.
	OrderFailedEventType = "OrderFailed"
	// OrderCompletedEventType is when order is completed (completed).
	OrderCompletedEventType = "OrderCompleted"
)

const (
	// OrderProductEventsConsumerGroup is the consumer group for order product events.
	OrderProductEventsConsumerGroup = "order-service.product-events"
	// OrderPaymentEventsConsumerGroup is the consumer group for order payment events.
	OrderPaymentEventsConsumerGroup = "order-service.payment-events"
	// OrderFulfillmentEventsConsumerGroup is the consumer group for order fulfillment events.
	OrderFulfillmentEventsConsumerGroup = "order-service.fulfillment-events"
	// OrderCartEventsConsumerGroup is the consumer group for order cart events.
	OrderCartEventsConsumerGroup = "order-service.cart-events"
)
