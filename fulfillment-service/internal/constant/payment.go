package constant

// PaymentStatus represents the status of a payment transaction.
type PaymentStatus string

const (
	// PaymentStatusPending indicates that the payment is pending.
	PaymentStatusPending PaymentStatus = "pending"
	// PaymentStatusProcessing indicates that the payment is being processed.
	PaymentStatusProcessing PaymentStatus = "processing"
	// PaymentStatusCompleted indicates that the payment has been completed successfully.
	PaymentStatusCompleted PaymentStatus = "completed"
	// PaymentStatusFailed indicates that the payment has failed.
	PaymentStatusFailed PaymentStatus = "failed"
	// PaymentStatusRefunded indicates that the payment has been refunded.
	PaymentStatusRefunded PaymentStatus = "refunded"
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
