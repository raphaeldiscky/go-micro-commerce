// Package resolver provides GraphQL resolvers for the auth service.
package resolver

import (
	"github.com/go-playground/validator/v10"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root resolver for GraphQL queries and mutations.
type Resolver struct {
	authService    service.AuthService
	addressService service.AddressService
	validator      *validator.Validate
	logger         logger.Logger
}

// NewResolver creates a new GraphQL resolver instance with the required dependencies.
func NewResolver(
	authService service.AuthService,
	addressService service.AddressService,
	appLogger logger.Logger,
) *Resolver {
	return &Resolver{
		authService:    authService,
		addressService: addressService,
		validator:      validator.New(),
		logger:         appLogger,
	}
}
