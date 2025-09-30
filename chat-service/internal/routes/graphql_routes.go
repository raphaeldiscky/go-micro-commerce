package routes

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph/resolver"
	chatmiddleware "github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/middleware"
)

// SetupGraphQLRoutes sets up all GraphQL routes.
func SetupGraphQLRoutes(
	e *echo.Echo,
	graphResolver *resolver.Resolver,
) {
	// Create GraphQL handler with context middleware
	graphHandler := handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{Resolvers: graphResolver}),
	)

	// Add middleware to extract client metadata from headers
	graphHandler.AroundOperations(middleware.GraphQLContextMiddleware())

	// GraphQL endpoint without auth (for introspection and public queries)
	// GET for introspection queries (needed by Apollo Router)
	e.GET("/graph", echo.WrapHandler(graphHandler))

	// POST for queries/mutations (public for introspection, use /graph/auth for protected)
	e.POST("/graph", echo.WrapHandler(graphHandler))

	// Protected GraphQL endpoint (requires authentication)
	e.POST("/graph/auth", echo.WrapHandler(graphHandler), chatmiddleware.AuthMiddleware)

	// GraphQL Playground (development only)
	playgroundHandler := playground.Handler("GraphQL Playground", "/graph")
	e.GET("/graph/playground", echo.WrapHandler(playgroundHandler))
}
