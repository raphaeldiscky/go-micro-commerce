package constant

import "time"

const (
	// ProductClientTimeout is the timeout for product client requests.
	ProductClientTimeout time.Duration = 5 * time.Second
	// FulfillmentClientTimeout is the timeout for fulfillment client requests.
	FulfillmentClientTimeout time.Duration = 5 * time.Second
)
