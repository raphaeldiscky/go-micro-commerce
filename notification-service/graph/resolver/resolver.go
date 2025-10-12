// Package resolver provides GraphQL resolvers for the notification service.
package resolver

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/subscription"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver provides GraphQL resolver dependencies.
type Resolver struct {
	notificationService service.NotificationService
	subscriptionManager *subscription.Manager
	logger              logger.Logger
}

// NewResolver creates a new Resolver with dependencies.
func NewResolver(
	notificationService service.NotificationService,
	subscriptionManager *subscription.Manager,
	appLogger logger.Logger,
) *Resolver {
	return &Resolver{
		notificationService: notificationService,
		subscriptionManager: subscriptionManager,
		logger:              appLogger,
	}
}
