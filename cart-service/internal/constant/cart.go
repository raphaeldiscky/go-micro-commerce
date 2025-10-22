package constant

// CartStatus represents the status of an cart.
type CartStatus string

const (
	// CartStatusPending indicates that the cart is pending.
	CartStatusPending CartStatus = "pending"
	// CartStatusProcessing indicates that the cart is being processed.
	CartStatusProcessing CartStatus = "processing"
	// CartStatusPaymentExpired indicates that the cart payment has expired.
	CartStatusPaymentExpired CartStatus = "payment_expired"
	// CartStatusPaid indicates that the cart has been paid.
	CartStatusPaid CartStatus = "paid"
	// CartStatusShipped indicates that the cart has been shipped.
	CartStatusShipped CartStatus = "shipped"
	// CartStatusDelivered indicates that the cart has been delivered.
	CartStatusDelivered CartStatus = "delivered"
	// CartStatusCanceled indicates that the cart has been canceled.
	CartStatusCanceled CartStatus = "canceled"
	// CartStatusFailed indicates that the cart has failed.
	CartStatusFailed CartStatus = "failed"
)
