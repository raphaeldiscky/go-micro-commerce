// Package routes provides the HTTP routes for the authentication service.
package routes

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgmiddleware "github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
)

// SetupGraphQLRoutes sets up all GraphQL routes.
func SetupGraphQLRoutes(
	e *echo.Echo,
	cfg *config.Config,
	graphResolver *resolver.Resolver,
	appLogger logger.Logger,
) {
	executableSchema := graph.NewExecutableSchema(graph.Config{
		Resolvers: graphResolver,
		Directives: graph.DirectiveRoot{
			RequiresAuth: pkgmiddleware.RequiresAuthDirective,
			RequiresRole: func(ctx context.Context, obj any, next graphql.Resolver, role graph.Role) (any, error) {
				// Convert graph.Role enum to string for middleware
				return pkgmiddleware.RequiresRoleDirective(ctx, obj, next, string(role))
			},
		},
	})

	// Create GraphQL handler with context middleware
	srv := handler.NewDefaultServer(executableSchema)

	// Add middleware to extract user headers from Apollo Router (forwarded from JWT claims)
	srv.AroundOperations(pkgmiddleware.GraphQLContextMiddleware())
	srv.AroundOperations(pkgmiddleware.GraphQLLoggingMiddleware(appLogger))

	e.GET("/graph", echo.WrapHandler(srv))
	e.POST("/graph", echo.WrapHandler(srv))

	// SSE Subscription Server (separate from main GraphQL server)
	// Use handler.New() instead of NewDefaultServer() to support SSE transport
	// NewDefaultServer() is deprecated and doesn't include SSE transport
	sseSrv := handler.New(executableSchema)
	// SSE transport handles both GET and POST methods when Accept: text/event-stream is present
	sseSrv.AddTransport(transport.SSE{
		KeepAlivePingInterval: constant.SubscriptionKeepAlivePingInterval,
	})
	// Options transport for CORS preflight requests
	sseSrv.AddTransport(transport.Options{})
	// DO NOT add GET/POST transports - they conflict with SSE and cause "transport not supported"
	// The SSE transport itself handles both GET and POST methods for the SSE protocol

	sseSrv.AroundOperations(pkgmiddleware.GraphQLContextMiddleware())
	sseSrv.AroundOperations(pkgmiddleware.GraphQLLoggingMiddleware(appLogger))

	// SSE endpoint supports both GET and POST methods
	// GET: subscription query in URL parameters
	// POST: subscription query in request body (used by graphql-sse client)
	e.GET("/graph/subscriptions/sse", echo.WrapHandler(sseSrv))
	e.POST("/graph/subscriptions/sse", echo.WrapHandler(sseSrv))

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
