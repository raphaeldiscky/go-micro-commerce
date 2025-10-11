// Package resolver provides GraphQL resolvers for the chat service.
package resolver

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/subscription"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root resolver for GraphQL queries and mutations.
type Resolver struct {
	chatService         service.ChatService
	connectionService   service.ConnectionService
	subscriptionManager *subscription.Manager
	hub                 *websocket.ChatHub
	logger              logger.Logger
}

// NewResolver creates a new GraphQL resolver instance with the required dependencies.
func NewResolver(
	chatService service.ChatService,
	connectionService service.ConnectionService,
	subscriptionManager *subscription.Manager,
	hub *websocket.ChatHub,
	appLogger logger.Logger,
) *Resolver {
	return &Resolver{
		chatService:         chatService,
		connectionService:   connectionService,
		subscriptionManager: subscriptionManager,
		hub:                 hub,
		logger:              appLogger,
	}
}
