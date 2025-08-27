package constant

// PaymentStatus represents the status of an order.
type PaymentStatus string

const (
	// PaymentStatusPending indicates that the order is pending.
	PaymentStatusPending PaymentStatus = "pending"
	// PaymentStatusConfirmed indicates that the order has been confirmed.
	PaymentStatusConfirmed PaymentStatus = "confirmed"
	// PaymentStatusPaid indicates that the order has been paid.
	PaymentStatusPaid PaymentStatus = "paid"
	// PaymentStatusShipped indicates that the order has been shipped.
	PaymentStatusShipped PaymentStatus = "shipped"
	// PaymentStatusDelivered indicates that the order has been delivered.
	PaymentStatusDelivered PaymentStatus = "delivered"
	// PaymentStatusCanceled indicates that the order has been canceled.
	PaymentStatusCanceled PaymentStatus = "canceled"
)

// PaymentMethod represents the different payment methods available.
type PaymentMethod string

const (
	// PaymentMethodCreditCard represents the credit card payment method.
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	// PaymentMethodDebitCard represents the debit card payment method.
	PaymentMethodDebitCard PaymentMethod = "debit_card"
	// PaymentMethodPayPal represents the PayPal payment method.
	PaymentMethodPayPal PaymentMethod = "paypal"
	// PaymentMethodBankTransfer represents the bank transfer payment method.
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)
