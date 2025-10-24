package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// SetupGraphQLResolver creates and configures the GraphQL resolver.
func SetupGraphQLResolver(
	paymentService service.PaymentService,
) *resolver.Resolver {
	return resolver.NewResolver(paymentService)
}
