package constant

// PaymentMethod represents the different payment methods available.
//
//nolint:recvcheck // ignore for marshalling graphql
type PaymentMethod string

const (
	// PaymentMethodCard represents the card payment method.
	PaymentMethodCard PaymentMethod = "card"
)

// PaymentGateway represents the different payment gateways available.
//
//nolint:recvcheck // ignore for marshalling graphql
type PaymentGateway string

const (
	// PaymentGatewayStripe represents the Stripe payment gateway.
	PaymentGatewayStripe PaymentGateway = "stripe"
	// PaymentGatewayMock represents the mock payment gateway.
	PaymentGatewayMock PaymentGateway = "mock"
)
