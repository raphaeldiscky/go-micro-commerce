// Package routes provides the HTTP routes for the authentication service.
package routes

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/graph/resolver"
	authmiddleware "github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/middleware"
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

	// GraphQL endpoint with optional auth middleware
	// GET for introspection queries (needed by Apollo Router)
	e.GET("/graph", echo.WrapHandler(graphHandler))

	// POST for actual queries/mutations (public for register/login, protected for me query)
	e.POST("/graph", echo.WrapHandler(graphHandler))

	// Protected GraphQL endpoint (requires authentication)
	e.POST("/graph/auth", echo.WrapHandler(graphHandler), authmiddleware.AuthMiddleware)

	// GraphQL Playground (development only)
	playgroundHandler := playground.Handler("GraphQL Playground", "/graph")
	e.GET("/graph/playground", echo.WrapHandler(playgroundHandler))
}
