package routes

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	chathandler "github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/middleware"
)

// SetupChatRoutes sets up all chat routes.
func SetupChatRoutes(
	e *echo.Echo,
	h *chathandler.ChatHandler,
	connHandler *chathandler.ConnectionHandler,
	graphqlResolver *graph.Resolver,
) {
	v1 := e.Group("/v1")
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)

	// Connection management routes
	protected.POST("/connect", connHandler.RequestConnection)
	protected.GET("/nodes/health", connHandler.GetNodeHealth)
	protected.POST("/validate-ticket", connHandler.ValidateTicket)

	// Chat data management routes (read-only and CRUD operations)
	protected.GET("/conversations", h.GetUserConversations)
	protected.POST("/conversations", h.CreateConversation)
	protected.GET("/:conversationID", h.GetConversation)
	protected.GET("/:conversationID/messages", h.GetMessages)
	protected.POST("/:conversationID/join", h.JoinConversation)
	protected.GET("/:conversationID/participants", h.GetParticipants)
	protected.GET("/users/online", h.GetOnlineUsers)

	// GraphQL endpoints
	graphqlHandler := handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{Resolvers: graphqlResolver}),
	)

	// GraphQL endpoint with auth middleware
	e.POST("/graphql", echo.WrapHandler(graphqlHandler), middleware.AuthMiddleware)

	// Also support GET for introspection queries (playground needs this)
	e.GET("/graphql", echo.WrapHandler(graphqlHandler))

	// GraphQL Playground (development only)
	playgroundHandler := playground.Handler("GraphQL Playground", "/graphql")
	e.GET("/graphql/playground", echo.WrapHandler(playgroundHandler))

	// NOTE: The following operations are handled via WebSocket messages on /v1/ws:
	// - Send Message → WebSocket message type "chat"
	// - Update Presence → WebSocket message type "presence"
	// - Typing Indicator → WebSocket message type "typing"
	// - Delivery Receipt → WebSocket message type "delivery_receipt"
	// - Read Receipt → WebSocket message type "read_receipt"
}
