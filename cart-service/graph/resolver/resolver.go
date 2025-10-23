// Package resolver provides GraphQL resolvers for the cart service.
package resolver

import (
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver provides GraphQL resolver dependencies.
type Resolver struct {
	cartService            service.CartService
	checkoutSessionService service.CheckoutSessionService
}

// NewResolver creates a new Resolver with dependencies.
func NewResolver(
	cartService service.CartService,
	checkoutSessionService service.CheckoutSessionService,
) *Resolver {
	return &Resolver{
		cartService:            cartService,
		checkoutSessionService: checkoutSessionService,
	}
}
