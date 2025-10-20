package constant

// PaymentMethod represents the different payment methods available.
type PaymentMethod string

const (
	// PaymentMethodCard represents the card payment method.
	PaymentMethodCard PaymentMethod = "card"
)

// PaymentGateway represents the different payment gateways available.
type PaymentGateway string

const (
	// PaymentGatewayStripe represents the Stripe payment gateway.
	PaymentGatewayStripe PaymentGateway = "stripe"
	// PaymentGatewayMock represents the mock payment gateway.
	PaymentGatewayMock PaymentGateway = "mock"
)
