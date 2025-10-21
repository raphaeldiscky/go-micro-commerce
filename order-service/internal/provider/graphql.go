package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/order-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
)

// SetupGraphQLResolver creates and configures the GraphQL resolver.
func SetupGraphQLResolver(
	orderService service.OrderService,
) *resolver.Resolver {
	return resolver.NewResolver(orderService)
}
