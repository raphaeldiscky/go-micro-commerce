package constant

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	// OrderStatusPending indicates that the order is pending.
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusProcessing indicates that the order is being processed.
	OrderStatusProcessing OrderStatus = "processing"
	// OrderStatusPaymentPending indicates that the order payment is pending.
	OrderStatusPaymentPending OrderStatus = "payment_pending"
	// OrderStatusPaymentExpired indicates that the order payment has expired.
	OrderStatusPaymentExpired OrderStatus = "payment_expired"
	// OrderStatusPaid indicates that the order has been paid.
	OrderStatusPaid OrderStatus = "paid"
	// OrderStatusShipped indicates that the order has been shipped.
	OrderStatusShipped OrderStatus = "shipped"
	// OrderStatusDelivered indicates that the order has been delivered.
	OrderStatusDelivered OrderStatus = "delivered"
	// OrderStatusCanceled indicates that the order has been canceled.
	OrderStatusCanceled OrderStatus = "canceled"
	// OrderStatusFailed indicates that the order has failed.
	OrderStatusFailed OrderStatus = "failed"
)
