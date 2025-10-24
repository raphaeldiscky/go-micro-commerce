package constant

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
