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
