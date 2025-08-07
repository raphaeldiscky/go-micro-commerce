// Package main implements the API Gateway for the microservices architecture.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/auth"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/metrics"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/ratelimit"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/tracing"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/monitoring"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/service"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize metrics
	metricsInstance := metrics.NewMetrics()

	// Initialize services
	discoveryService, err := service.NewConsulServiceDiscovery(cfg.ServiceDiscovery)
	if err != nil {
		logger.Fatal("Failed to initialize service discovery", zap.Error(err))
	}

	// Initialize circuit breaker
	circuitBreaker := service.NewCircuitBreaker()

	// Initialize load balancer
	loadBalancer := service.NewLoadBalancer()

	// Create Echo instance
	e := echo.New()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(tracing.Middleware())
	e.Use(metricsInstance.Middleware())
	e.Use(ratelimit.Middleware(*cfg.RateLimit))

	// Initialize monitoring handler
	monitoringHandler := monitoring.NewHandler(logger, "1.0.0")
	monitoringHandler.RegisterRoutes(e)

	// Metrics endpoint (Prometheus format)
	e.GET("/metrics", metrics.Handler())

	// Initialize API Gateway
	gw := gateway.New(gateway.Config{
		Logger:           logger,
		ServiceDiscovery: discoveryService,
		CircuitBreaker:   circuitBreaker,
		LoadBalancer:     loadBalancer,
		Metrics:          metricsInstance,
		Config:           cfg,
	})

	// Setup routes
	setupRoutes(e, gw, cfg)

	// Start server
	go func() {
		if err := e.Start(fmt.Sprintf(":%d", cfg.HTTPServer.Port)); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("API Gateway started", zap.Int("port", cfg.HTTPServer.Port))

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API Gateway...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Fatal("Failed to shutdown server", zap.Error(err))
	}

	logger.Info("API Gateway stopped")
}

func setupRoutes(e *echo.Echo, gw *gateway.Gateway, cfg *config.Config) {
	// Health endpoints for services (no auth required)
	healthGroup := e.Group("/services")
	healthGroup.GET("/product-service/health", gw.ProxyToService("product-service", "/health"))
	healthGroup.GET("/auth-service/health", gw.ProxyToService("auth-service", "/health"))
	healthGroup.GET("/order-service/health", gw.ProxyToService("order-service", "/health"))
	healthGroup.GET(
		"/notification-service/health",
		gw.ProxyToService("notification-service", "/health"),
	)

	// API version 1
	v1 := e.Group("/api/v1")

	// Authentication routes (no auth required)
	authGroup := v1.Group("/auth")
	authGroup.POST("/register", gw.ProxyToService("auth-service", "/api/v1/auth/register"))
	authGroup.POST("/login", gw.ProxyToService("auth-service", "/api/v1/auth/login"))
	authGroup.POST("/verify-email", gw.ProxyToService("auth-service", "/api/v1/auth/verify-email"))
	authGroup.POST(
		"/resend-verification",
		gw.ProxyToService("auth-service", "/api/v1/auth/resend-verification"),
	)

	// Protected routes
	protected := v1.Group("")
	protected.Use(auth.JWTMiddleware(cfg.JWT))

	// User routes
	userGroup := protected.Group("/users")
	userGroup.GET("/:id", gw.ProxyToService("auth-service", "/api/v1/users/:id"))
	userGroup.PUT("/:id", gw.ProxyToService("auth-service", "/api/v1/users/:id"))

	// Product routes
	productGroup := protected.Group("/products")
	productGroup.GET("", gw.ProxyToService("product-service", "/api/v1/products"))
	productGroup.GET("/:id", gw.ProxyToService("product-service", "/api/v1/products/:id"))
	productGroup.POST("", gw.ProxyToService("product-service", "/api/v1/products"))
	productGroup.PUT("/:id", gw.ProxyToService("product-service", "/api/v1/products/:id"))
	productGroup.DELETE("/:id", gw.ProxyToService("product-service", "/api/v1/products/:id"))

	// Order routes
	orderGroup := protected.Group("/orders")
	orderGroup.GET("", gw.ProxyToService("order-service", "/api/v1/orders"))
	orderGroup.GET("/:id", gw.ProxyToService("order-service", "/api/v1/orders/:id"))
	orderGroup.POST("", gw.ProxyToService("order-service", "/api/v1/orders"))
	orderGroup.PUT("/:id", gw.ProxyToService("order-service", "/api/v1/orders/:id"))

	// Notification routes (admin only)
	notificationGroup := protected.Group("/notifications")
	// Add admin middleware here when implemented
	notificationGroup.GET("", gw.ProxyToService("notification-service", "/api/v1/notifications"))
	notificationGroup.POST(
		"/send",
		gw.ProxyToService("notification-service", "/api/v1/notifications/send"),
	)
}
