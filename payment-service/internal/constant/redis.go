package constant

import "time"

const (
	// CreateOrderTTL is the time-to-live for create order locks.
	CreateOrderTTL = 10 * time.Second
	// CreateOrderRetryInterval is the retry interval for creating orders.
	CreateOrderRetryInterval = 500 * time.Millisecond
	// CreateOrderRetryLimit is the maximum number of retries for creating orders.
	CreateOrderRetryLimit = 3
)
