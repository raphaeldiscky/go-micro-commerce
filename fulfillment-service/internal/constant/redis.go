package constant

import "time"

const (
	// CreatePaymentTTL is the time-to-live for create order locks.
	CreatePaymentTTL = 10 * time.Second
	// CreatePaymentRetryInterval is the retry interval for creating orders.
	CreatePaymentRetryInterval = 500 * time.Millisecond
	// CreatePaymentRetryLimit is the maximum number of retries for creating orders.
	CreatePaymentRetryLimit = 3
)
