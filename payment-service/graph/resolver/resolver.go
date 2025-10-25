// Package resolver provides GraphQL resolvers for the payment service.
package resolver

import (
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver provides GraphQL resolver dependencies.
type Resolver struct {
	paymentService service.PaymentService
}

// NewResolver creates a new Resolver with dependencies.
func NewResolver(paymentService service.PaymentService) *Resolver {
	return &Resolver{
		paymentService: paymentService,
	}
}
