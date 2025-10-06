package routes

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"

	pkgmiddleware "github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/middleware"
)

// SetupGraphQLRoutes sets up all GraphQL routes.
func SetupGraphQLRoutes(
	e *echo.Echo,
	cfg *config.Config,
	graphResolver *resolver.Resolver,
) {
	executableSchema := graph.NewExecutableSchema(graph.Config{Resolvers: graphResolver})

	// Create GraphQL handler with context middleware
	graphHandler := handler.NewDefaultServer(executableSchema)

	// Add middleware to extract client metadata from headers
	graphHandler.AroundOperations(pkgmiddleware.GraphQLContextMiddleware())

	// GraphQL endpoint without auth (for introspection and public queries)
	// GET for introspection queries (needed by Apollo Router)
	e.GET("/graph", echo.WrapHandler(graphHandler))

	// POST for queries/mutations (public for introspection, use /graph/auth for protected)
	e.POST("/graph", echo.WrapHandler(graphHandler))
	// Protected GraphQL endpoint (requires authentication)
	e.POST("/graph/auth", echo.WrapHandler(graphHandler), middleware.AuthMiddleware)

	// WebSocket handler for GraphQL subscriptions with graphql-transport-ws protocol
	wsHandler := handler.New(executableSchema)

	// Configure WebSocket transport (graphql-transport-ws protocol)
	wsHandler.AddTransport(transport.Websocket{
		KeepAlivePingInterval: constant.GraphQLKeepAlivePingInterval,
	})

	// Add context middleware for subscriptions
	wsHandler.AroundOperations(pkgmiddleware.GraphQLContextMiddleware())

	// WebSocket subscriptions endpoint (protected with auth)
	e.GET("/graph/subscriptions", echo.WrapHandler(wsHandler), middleware.AuthMiddleware)

	if cfg.App.Environment == "development" {
		playgroundHandler := playground.Handler("GraphQL Playground", "/graph")

		e.GET("/graph/playground", func(c echo.Context) error {
			// Relax CSP for GraphQL Playground
			c.Response().Header().Set("Content-Security-Policy",
				"default-src 'self' 'unsafe-inline' 'unsafe-eval' data: blob: https://cdn.jsdelivr.net https://unpkg.com;")
			c.Response().Header().Set("X-Frame-Options", "SAMEORIGIN")
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			playgroundHandler.ServeHTTP(c.Response(), c.Request())

			return nil
		})
	}
}
