package constant

import "time"

// CheckoutSessionStatus represents the status of a checkout session.
//
//nolint:recvcheck // ignore for marshalling graphql
type CheckoutSessionStatus string

const (
	// CheckoutSessionStatusPending indicates that the checkout session is pending.
	CheckoutSessionStatusPending CheckoutSessionStatus = "pending"
	// CheckoutSessionStatusOrderPlaced indicates that the checkout session has been placed as an order.
	CheckoutSessionStatusOrderPlaced CheckoutSessionStatus = "order_placed"
	// CheckoutSessionStatusCanceled indicates that the checkout session has been canceled.
	CheckoutSessionStatusCanceled CheckoutSessionStatus = "canceled"
)

const (
	// CreateCheckoutSessionTTL is the TTL for creating a checkout session.
	CreateCheckoutSessionTTL = 10 * time.Second
	// CreateCheckoutSessionRetryLimit is the retry limit for creating a checkout session.
	CreateCheckoutSessionRetryLimit = 3
	// CreateCheckoutSessionRetryInterval is the retry interval for creating a checkout session.
	CreateCheckoutSessionRetryInterval = 500 * time.Millisecond

	// PlaceOrderTTL is the TTL for placing an order.
	PlaceOrderTTL = 10 * time.Second
	// PlaceOrderRetryLimit is the retry limit for placing an order.
	PlaceOrderRetryLimit = 3
	// PlaceOrderRetryInterval is the retry interval for placing an order.
	PlaceOrderRetryInterval = 500 * time.Millisecond
)

const (
	// CheckoutSessionExpirationTime is the expiration time for checkout sessions.
	CheckoutSessionExpirationTime = 30 * time.Minute
)
