package constant

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	// OrderStatusPending indicates that the order is pending.
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusConfirmed indicates that the order has been confirmed.
	OrderStatusConfirmed OrderStatus = "confirmed"
	// OrderStatusPaid indicates that the order has been paid.
	OrderStatusPaid OrderStatus = "paid"
	// OrderStatusShipped indicates that the order has been shipped.
	OrderStatusShipped OrderStatus = "shipped"
	// OrderStatusDelivered indicates that the order has been delivered.
	OrderStatusDelivered OrderStatus = "delivered"
	// OrderStatusCanceled indicates that the order has been canceled.
	OrderStatusCanceled OrderStatus = "canceled"
)
