package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
)

// SetupGraphQLResolver creates and configures the GraphQL resolver.
func SetupGraphQLResolver(
	cartService service.CartService,
	checkoutSessionService service.CheckoutSessionService,
) *resolver.Resolver {
	return resolver.NewResolver(cartService, checkoutSessionService)
}
