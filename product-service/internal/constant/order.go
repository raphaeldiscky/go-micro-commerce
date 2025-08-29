package constant

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	// OrderStatusPending indicates that the order is pending (inventory reversed and need to be paid).
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusPaid indicates that the order has been paid.
	OrderStatusPaid OrderStatus = "paid"
	// OrderStatusShipped indicates that the order has been shipped.
	OrderStatusShipped OrderStatus = "shipped"
	// OrderStatusDelivered indicates that the order has been delivered.
	OrderStatusDelivered OrderStatus = "delivered"
	// OrderStatusCanceled indicates that the order has been canceled.
	OrderStatusCanceled OrderStatus = "canceled"
)
