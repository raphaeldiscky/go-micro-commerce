package constant

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	// OrderStatusPending indicates that the order is pending.
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusPaid indicates that the order has been paid.
	OrderStatusPaid OrderStatus = "paid"
	// OrderStatusConfirmed indicates that the order has been confirmed after payment.
	OrderStatusConfirmed OrderStatus = "confirmed"
	// OrderStatusProcessing indicates that the order is being processed.
	OrderStatusProcessing OrderStatus = "processing"
	// OrderStatusShipped indicates that the order has been shipped.
	OrderStatusShipped OrderStatus = "shipped"
	// OrderStatusDelivered indicates that the order has been delivered.
	OrderStatusDelivered OrderStatus = "delivered"
	// OrderStatusCanceled indicates that the order has been canceled.
	OrderStatusCanceled OrderStatus = "canceled"
	// OrderStatusFailed indicates that the order has failed.
	OrderStatusFailed OrderStatus = "failed"
)

// CancelOrderReason represents the reason for canceling an order.
type CancelOrderReason string

const (
	// CancelOrderReasonPaymentTimeout for payment timeout cancellation.
	CancelOrderReasonPaymentTimeout CancelOrderReason = "payment_timeout"
	// CancelOrderReasonUserCancellation for user-initiated cancellation.
	CancelOrderReasonUserCancellation CancelOrderReason = "user_cancellation"
	// CancelOrderReasonSystemError for system error cancellation.
	CancelOrderReasonSystemError CancelOrderReason = "system_error"
)
