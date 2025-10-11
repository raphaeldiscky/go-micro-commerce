// Package routes provides the HTTP routes for the authentication service.
package routes

// import (
// 	"context"

// 	"github.com/99designs/gqlgen/graphql"
// 	"github.com/99designs/gqlgen/graphql/handler"
// 	"github.com/99designs/gqlgen/graphql/handler/transport"
// 	"github.com/99designs/gqlgen/graphql/playground"
// 	"github.com/labstack/echo/v4"
// 	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
// 	"github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"

// 	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph"
// 	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph/resolver"
// 	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
// )

// // SetupGraphQLRoutes sets up all GraphQL routes.
// func SetupGraphQLRoutes(
// 	e *echo.Echo,
// 	cfg *config.Config,
// 	graphResolver *resolver.Resolver,
// 	appLogger logger.Logger,
// ) {
// 	// Create GraphQL handler with context middleware and custom directives
// 	srv := handler.NewDefaultServer(
// 		graph.NewExecutableSchema(graph.Config{
// 			Resolvers: graphResolver,
// 			Directives: graph.DirectiveRoot{
// 				RequiresAuth: middleware.RequiresAuthDirective,
// 				RequiresRole: func(ctx context.Context, obj any, next graphql.Resolver, role graph.Role) (any, error) {
// 					// Convert graph.Role enum to string for middleware
// 					return middleware.RequiresRoleDirective(ctx, obj, next, string(role))
// 				},
// 			},
// 		}),
// 	)
// 	srv.AddTransport(transport.SSE{})
// 	srv.AddTransport(transport.Options{})
// 	srv.AddTransport(transport.GET{})
// 	srv.AddTransport(transport.POST{})

// 	// Add middleware to extract user headers from Apollo Router (forwarded from JWT claims)
// 	srv.AroundOperations(middleware.GraphQLContextMiddleware())

// 	// Add logging middleware to log GraphQL operations
// 	srv.AroundOperations(middleware.GraphQLLoggingMiddleware(appLogger))

// 	e.GET("/graph", echo.WrapHandler(srv))
// 	e.POST("/graph", echo.WrapHandler(srv))

// 	if cfg.App.Environment == "development" {
// 		playgroundHandler := playground.Handler("GraphQL Playground", "/graph")

// 		e.GET("/graph/playground", func(c echo.Context) error {
// 			// Relax CSP for GraphQL Playground
// 			c.Response().Header().Set("Content-Security-Policy",
// 				"default-src 'self' 'unsafe-inline' 'unsafe-eval' data: blob: https://cdn.jsdelivr.net https://unpkg.com;")
// 			c.Response().Header().Set("X-Frame-Options", "SAMEORIGIN")
// 			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
// 			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
// 			playgroundHandler.ServeHTTP(c.Response(), c.Request())

// 			return nil
// 		})
// 	}
// }
