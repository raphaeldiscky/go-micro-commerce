package routes

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgmiddleware "github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	chatconstant "github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
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

	// Add middleware to extract client metadata from headers
	srv.AroundOperations(pkgmiddleware.GraphQLContextMiddleware())

	// Add logging middleware to log GraphQL operations
	srv.AroundOperations(pkgmiddleware.GraphQLLoggingMiddleware(appLogger))

	// GraphQL endpoint without auth (for introspection and public queries)
	e.GET("/graph", graphQLEchoHandler(srv))
	e.POST("/graph", graphQLEchoHandler(srv))

	// WebSocket handler for GraphQL subscriptions with graphql-transport-ws protocol
	wsSrv := handler.New(executableSchema)

	// Configure the WebSocket transport
	// The Upgrader must allow connections from API Gateway proxy
	wsSrv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: chatconstant.GraphQLKeepAlivePingInterval,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true // Allow all origins for proxying through API Gateway
			},
			ReadBufferSize:  chatconstant.WsServerReadBufferSize,
			WriteBufferSize: chatconstant.WsServerWriteBufferSize,
		},
	})
	wsSrv.AddTransport(transport.Options{})
	wsSrv.AddTransport(transport.GET{})
	wsSrv.AddTransport(transport.POST{})
	// Add context middleware for subscriptions
	wsSrv.AroundOperations(pkgmiddleware.GraphQLContextMiddleware())

	// Add logging middleware for subscriptions
	wsSrv.AroundOperations(pkgmiddleware.GraphQLLoggingMiddleware(appLogger))

	// WebSocket subscriptions endpoint
	// Auth is handled by API Gateway which validates JWT and forwards X-User-* headers
	// The WebSocketAuthMiddleware extracts these headers from HTTP upgrade request and adds to context
	e.GET(
		"/graph/subscriptions/ws",
		echo.WrapHandler(wsSrv),
		pkgmiddleware.WebSocketAuthMiddleware(),
	)

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
