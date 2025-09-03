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
	// PaymentMethodBankTransfer represents the bank transfer payment method.
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	// PaymentMethodApplePay represents the Apple Pay payment method.
	PaymentMethodApplePay PaymentMethod = "apple_pay"
	// PaymentMethodGooglePay represents the Google Pay payment method.
	PaymentMethodGooglePay PaymentMethod = "google_pay"
	// PaymentMethodPayPal represents the PayPal payment method.
	PaymentMethodPayPal PaymentMethod = "paypal"
)

// BankTransferStatus represents the status of a bank transfer.
type BankTransferStatus string

const (
	// BankTransferStatusPending indicates the transfer is pending.
	BankTransferStatusPending BankTransferStatus = "pending"
	// BankTransferStatusProcessing indicates the transfer is being processed.
	BankTransferStatusProcessing BankTransferStatus = "processing"
	// BankTransferStatusCompleted indicates the transfer is completed.
	BankTransferStatusCompleted BankTransferStatus = "completed"
	// BankTransferStatusFailed indicates the transfer failed.
	BankTransferStatusFailed BankTransferStatus = "failed"
	// BankTransferStatusCancelled indicates the transfer was canceled.
	BankTransferStatusCancelled BankTransferStatus = "canceled"
)

// PaymentGatewayStatus represents the status of a payment gateway transaction.
type PaymentGatewayStatus string

const (
	// PaymentGatewayStatusPending indicates the payment is pending.
	PaymentGatewayStatusPending PaymentGatewayStatus = "pending"
	// PaymentGatewayStatusAuthorized indicates the payment is authorized but not captured.
	PaymentGatewayStatusAuthorized PaymentGatewayStatus = "authorized"
	// PaymentGatewayStatusSucceeded indicates the payment succeeded.
	PaymentGatewayStatusSucceeded PaymentGatewayStatus = "succeeded"
	// PaymentGatewayStatusFailed indicates the payment failed.
	PaymentGatewayStatusFailed PaymentGatewayStatus = "failed"
	// PaymentGatewayStatusCancelled indicates the payment was canceled.
	PaymentGatewayStatusCancelled PaymentGatewayStatus = "canceled"
	// PaymentGatewayStatusRequiresAction indicates the payment requires additional action.
	PaymentGatewayStatusRequiresAction PaymentGatewayStatus = "requires_action"
)

// DigitalWalletType represents different digital wallet types.
type DigitalWalletType string

const (
	// DigitalWalletTypeApplePay represents Apple Pay.
	DigitalWalletTypeApplePay DigitalWalletType = "apple_pay"
	// DigitalWalletTypeGooglePay represents Google Pay.
	DigitalWalletTypeGooglePay DigitalWalletType = "google_pay"
	// DigitalWalletTypePayPal represents PayPal.
	DigitalWalletTypePayPal DigitalWalletType = "paypal"
)

// PaymentActionType represents types of actions required for payment completion.
type PaymentActionType string

const (
	// PaymentActionTypeRedirect requires user to be redirected to a URL.
	PaymentActionTypeRedirect PaymentActionType = "redirect"
	// PaymentActionType3DSecure requires 3D Secure authentication.
	PaymentActionType3DSecure PaymentActionType = "3d_secure"
	// PaymentActionTypeOTP requires OTP verification.
	PaymentActionTypeOTP PaymentActionType = "otp"
)

// RefundStatus represents the status of a refund.
type RefundStatus string

const (
	// RefundStatusPending indicates the refund is pending.
	RefundStatusPending RefundStatus = "pending"
	// RefundStatusProcessing indicates the refund is being processed.
	RefundStatusProcessing RefundStatus = "processing"
	// RefundStatusSucceeded indicates the refund succeeded.
	RefundStatusSucceeded RefundStatus = "succeeded"
	// RefundStatusFailed indicates the refund failed.
	RefundStatusFailed RefundStatus = "failed"
	// RefundStatusCancelled indicates the refund was canceled.
	RefundStatusCancelled RefundStatus = "canceled"
)
