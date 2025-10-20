package constant

// PaymentStatus represents the status of a payment transaction.
type PaymentStatus string

const (
	// PaymentStatusPending indicates that the payment is pending.
	PaymentStatusPending PaymentStatus = "pending"
	// PaymentStatusProcessing indicates that the payment is being processed.
	PaymentStatusProcessing PaymentStatus = "processing"
	// PaymentStatusTimeout indicates that the payment has timed out.
	PaymentStatusTimeout PaymentStatus = "timeout"
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
	// PaymentMethodCard represents the card payment method.
	PaymentMethodCard PaymentMethod = "card"
	// PaymentMethodBankTransfer represents the bank transfer payment method.
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

// PaymentGateway represents the different payment gateways available.
type PaymentGateway string

const (
	// PaymentGatewayStripe represents the Stripe payment gateway.
	PaymentGatewayStripe PaymentGateway = "stripe"
	// PaymentGatewayMock represents the Mock payment gateway for testing.
	PaymentGatewayMock PaymentGateway = "mock"
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
	// PaymentGatewayStatusCanceled indicates the payment was canceled.
	PaymentGatewayStatusCanceled PaymentGatewayStatus = "canceled"
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
	// RefundStatusCanceled indicates the refund was canceled.
	RefundStatusCanceled RefundStatus = "canceled"
)
