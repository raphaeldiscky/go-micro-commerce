// Package resolver provides GraphQL resolvers for the chat service.
package resolver

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root resolver for GraphQL queries and mutations.
type Resolver struct {
	chatService service.ChatService
	logger      logger.Logger
}

// NewResolver creates a new GraphQL resolver instance with the required dependencies.
func NewResolver(chatService service.ChatService, appLogger logger.Logger) *Resolver {
	return &Resolver{
		chatService: chatService,
		logger:      appLogger,
	}
}
