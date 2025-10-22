package constant

import "time"

const (
	// ClientProductGRPCPort is the port for the product gRPC client.
	ClientProductGRPCPort = 50052
	// ClientFulfillmentGRPCPort is the port for the fulfillment gRPC client.
	ClientFulfillmentGRPCPort = 50055
)

const (
	// ProductClientTimeout is the timeout for product client requests.
	ProductClientTimeout time.Duration = 5 * time.Second
	// FulfillmentClientTimeout is the timeout for fulfillment client requests.
	FulfillmentClientTimeout time.Duration = 5 * time.Second
)
