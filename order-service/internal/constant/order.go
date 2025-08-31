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
)
