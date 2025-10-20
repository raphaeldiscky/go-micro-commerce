package constant

// PaymentMethod represents the different payment methods available.
type PaymentMethod string

const (
	// PaymentMethodCard represents the debit card payment method.
	PaymentMethodCard PaymentMethod = "card"
	// PaymentMethodBankTransfer represents the bank transfer payment method.
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

// PaymentGateway represents the different payment gateways available.
type PaymentGateway string

const (
	// PaymentGatewayStripe represents the Stripe payment gateway.
	PaymentGatewayStripe PaymentGateway = "stripe"
	// PaymentGatewayMock represents the mock payment gateway.
	PaymentGatewayMock PaymentGateway = "mock"
)
