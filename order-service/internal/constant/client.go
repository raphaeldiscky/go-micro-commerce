package constant

import "time"

const (
	// ClientProductGRPCPort is the port for the product gRPC client.
	ClientProductGRPCPort = 50052
	// ClientFulfillmentGRPCPort is the port for the fulfillment gRPC client.
	ClientFulfillmentGRPCPort = 50055
	// ClientCartGRPCPort is the port for the cart gRPC client.
	ClientCartGRPCPort = 50056
	// ClientPaymentGRPCPort is the port for the payment gRPC client.
	ClientPaymentGRPCPort = 50057
)

const (
	// ProductClientTimeout is the timeout for product client requests.
	ProductClientTimeout time.Duration = 5 * time.Second
	// FulfillmentClientTimeout is the timeout for fulfillment client requests.
	FulfillmentClientTimeout time.Duration = 5 * time.Second
	// CartClientTimeout is the timeout for cart client requests.
	CartClientTimeout time.Duration = 5 * time.Second
	// PaymentClientTimeout is the timeout for payment client requests.
	PaymentClientTimeout time.Duration = 10 * time.Second
)
