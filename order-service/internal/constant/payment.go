package constant

// PaymentGateway represents the different payment gateways available.
//
//nolint:recvcheck // ignore for marshalling graphql
type PaymentGateway string

const (
	// PaymentGatewayStripe represents the Stripe payment gateway.
	PaymentGatewayStripe PaymentGateway = "stripe"
	// PaymentGatewayXendit represents the Xendit payment gateway.
	PaymentGatewayXendit PaymentGateway = "xendit"
	// PaymentGatewayMock represents the mock payment gateway.
	PaymentGatewayMock PaymentGateway = "mock"
)
