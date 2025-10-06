// Package routes provides the HTTP routes for the authentication service.
package routes

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/config"
	authmiddleware "github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/middleware"
)

// SetupGraphQLRoutes sets up all GraphQL routes.
func SetupGraphQLRoutes(
	e *echo.Echo,
	cfg *config.Config,
	graphResolver *resolver.Resolver,
) {
	// Create GraphQL handler with context middleware
	graphHandler := handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{Resolvers: graphResolver}),
	)

	// Add middleware to extract user headers from Apollo Router (forwarded from JWT claims)
	graphHandler.AroundOperations(middleware.GraphQLContextMiddleware())

	// GraphQL endpoint with optional auth middleware
	// GET for introspection queries (needed by Apollo Router)
	e.GET("/graph", graphQLEchoHandler(graphHandler))

	// POST for actual queries/mutations (public for register/login, protected for me query)
	e.POST("/graph", graphQLEchoHandler(graphHandler))

	// Protected GraphQL endpoint (requires authentication)
	e.POST("/graph/auth", graphQLEchoHandler(graphHandler), authmiddleware.AuthMiddleware)

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

// graphQLEchoHandler wraps GraphQL handler to pass Echo context through to resolvers.
func graphQLEchoHandler(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()

		// Add response writer to context so resolvers can set cookies
		ctx := context.WithValue(req.Context(), constant.CtxKeyResponseWriter, res.Writer)
		req = req.WithContext(ctx)

		h.ServeHTTP(res, req)

		return nil
	}
}
