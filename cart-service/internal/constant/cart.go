package constant

// CartStatus represents the status of a cart.
//
//nolint:recvcheck // ignore for marshalling graphql
type CartStatus string

const (
	// CartStatusActive indicates that the cart is active.
	CartStatusActive CartStatus = "active"
	// CartStatusCheckedOut indicates that the cart is being checked out.
	CartStatusCheckedOut CartStatus = "checked_out"
	// CartStatusArchived indicates that the cart has been archived.
	CartStatusArchived CartStatus = "archived"
)
