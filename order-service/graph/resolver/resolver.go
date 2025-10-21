// Package resolver provides GraphQL resolvers for the order service.
package resolver

import (
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver provides GraphQL resolver dependencies.
type Resolver struct {
	orderService service.OrderService
}

// NewResolver creates a new Resolver with dependencies.
func NewResolver(orderService service.OrderService) *Resolver {
	return &Resolver{
		orderService: orderService,
	}
}
