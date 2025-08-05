# Service Integration Example: Product Service with Consul Registration

This example shows how to modify the product service to register itself with Consul for automatic service discovery by Traefik.

## Step 1: Add Consul Dependencies

Add to your service's `go.mod` file:

```go
require (
    github.com/hashicorp/consul/api v1.29.4
    go.uber.org/zap v1.27.0
)
```

## Step 2: Create Registration Utility

Since each service is a separate module, copy the registration utility or create a shared library:

```go
// internal/consul/registration.go
package consul

import (
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"

    "github.com/hashicorp/consul/api"
    "go.uber.org/zap"
)

// ServiceRegistration handles service registration with Consul.
type ServiceRegistration struct {
    client *api.Client
    logger *zap.Logger
}

// NewServiceRegistration creates a new service registration client.
func NewServiceRegistration(consulAddr string, logger *zap.Logger) (*ServiceRegistration, error) {
    config := api.DefaultConfig()
    config.Address = consulAddr

    client, err := api.NewClient(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create consul client: %w", err)
    }

    if logger == nil {
        logger = zap.NewNop()
    }

    return &ServiceRegistration{
        client: client,
        logger: logger,
    }, nil
}

// ServiceConfig holds configuration for service registration.
type ServiceConfig struct {
    ServiceName    string
    ServiceID      string
    Address        string
    Port           int
    Tags           []string
    Meta           map[string]string
    HealthCheckURL string
    CheckInterval  string
    CheckTimeout   string
}

// Register registers a service with Consul.
func (s *ServiceRegistration) Register(config ServiceConfig) error {
    // Generate service ID if not provided
    if config.ServiceID == "" {
        config.ServiceID = fmt.Sprintf("%s-%s", config.ServiceName, generateInstanceID())
    }

    // Set default health check URL if not provided
    if config.HealthCheckURL == "" {
        config.HealthCheckURL = fmt.Sprintf("http://%s:%d/health", config.Address, config.Port)
    }

    // Set default check intervals
    if config.CheckInterval == "" {
        config.CheckInterval = "30s"
    }
    if config.CheckTimeout == "" {
        config.CheckTimeout = "10s"
    }

    registration := &api.AgentServiceRegistration{
        ID:      config.ServiceID,
        Name:    config.ServiceName,
        Tags:    config.Tags,
        Address: config.Address,
        Port:    config.Port,
        Meta:    config.Meta,
        Check: &api.AgentServiceCheck{
            HTTP:                           config.HealthCheckURL,
            Interval:                       config.CheckInterval,
            Timeout:                        config.CheckTimeout,
            DeregisterCriticalServiceAfter: "60s",
        },
    }

    s.logger.Info("Registering service with Consul",
        zap.String("service_id", config.ServiceID),
        zap.String("service_name", config.ServiceName),
        zap.String("address", config.Address),
        zap.Int("port", config.Port),
        zap.Strings("tags", config.Tags),
    )

    if err := s.client.Agent().ServiceRegister(registration); err != nil {
        return fmt.Errorf("failed to register service: %w", err)
    }

    s.logger.Info("Service registered successfully",
        zap.String("service_id", config.ServiceID),
        zap.String("service_name", config.ServiceName),
    )

    return nil
}

// Deregister removes a service from Consul.
func (s *ServiceRegistration) Deregister(serviceID string) error {
    s.logger.Info("Deregistering service from Consul",
        zap.String("service_id", serviceID),
    )

    if err := s.client.Agent().ServiceDeregister(serviceID); err != nil {
        return fmt.Errorf("failed to deregister service: %w", err)
    }

    s.logger.Info("Service deregistered successfully",
        zap.String("service_id", serviceID),
    )

    return nil
}

// CreateTraefikTags creates Traefik-specific tags for service discovery.
func CreateTraefikTags(serviceName, host string, port int, additionalTags ...string) []string {
    tags := []string{
        "api",
        "v1",
        "traefik.enable=true",
        fmt.Sprintf("traefik.http.routers.%s.rule=Host(`%s`)", serviceName, host),
        fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%d", serviceName, port),
        fmt.Sprintf("traefik.http.routers.%s.entrypoints=web", serviceName),
    }

    tags = append(tags, additionalTags...)
    return tags
}

// Helper functions...
func generateInstanceID() string {
    hostname, _ := os.Hostname()
    if hostname == "" {
        hostname = "unknown"
    }
    return hostname
}
```

## Step 3: Modify Service Main Function

Update your service's `main.go` to register with Consul:

```go
// cmd/main.go
package main

import (
    "context"
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

    "your-service/internal/config"
    "your-service/internal/consul"
    // ... other imports
)

func main() {
    // Initialize logger
    logger, err := zap.NewProduction()
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    defer logger.Sync()

    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        logger.Fatal("Failed to load configuration", zap.Error(err))
    }

    // Initialize Consul registration
    consulAddr := os.Getenv("CONSUL_ADDRESS")
    if consulAddr == "" {
        consulAddr = "localhost:8500"
    }

    consulReg, err := consul.NewServiceRegistration(consulAddr, logger)
    if err != nil {
        logger.Fatal("Failed to create consul registration", zap.Error(err))
    }

    // Service configuration
    serviceName := "product-service"
    serviceAddress := os.Getenv("SERVICE_ADDRESS")
    if serviceAddress == "" {
        serviceAddress = "localhost"
    }
    servicePort := cfg.HTTPServer.Port

    // Create Traefik tags for automatic routing
    tags := consul.CreateTraefikTags(
        serviceName,
        "products.localhost", // This will be the domain for accessing the service
        servicePort,
        "products", // Additional tag
        "microservice",
    )

    // Register service with Consul
    serviceConfig := consul.ServiceConfig{
        ServiceName: serviceName,
        Address:     serviceAddress,
        Port:        servicePort,
        Tags:        tags,
        Meta: map[string]string{
            "version":     "1.0.0",
            "environment": cfg.App.Environment,
        },
    }

    if err := consulReg.Register(serviceConfig); err != nil {
        logger.Fatal("Failed to register service with Consul", zap.Error(err))
    }

    // Deregister on shutdown
    defer func() {
        if err := consulReg.Deregister(serviceConfig.ServiceID); err != nil {
            logger.Error("Failed to deregister service", zap.Error(err))
        }
    }()

    // Create Echo instance
    e := echo.New()

    // Add middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())

    // Health check endpoint (required for Consul health checks)
    e.GET("/health", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "status":  "healthy",
            "service": serviceName,
            "version": "1.0.0",
        })
    })

    // Your API routes
    setupRoutes(e, cfg)

    // Start server
    go func() {
        addr := fmt.Sprintf(":%d", servicePort)
        if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
            logger.Fatal("Failed to start server", zap.Error(err))
        }
    }()

    logger.Info("Service started",
        zap.String("service", serviceName),
        zap.String("address", serviceAddress),
        zap.Int("port", servicePort),
    )

    // Wait for interrupt signal to gracefully shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("Shutting down service...")

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := e.Shutdown(ctx); err != nil {
        logger.Fatal("Failed to shutdown server", zap.Error(err))
    }

    logger.Info("Service stopped")
}

func setupRoutes(e *echo.Echo, cfg *config.Config) {
    // Your existing routes setup
    api := e.Group("/api/v1")

    // Product routes
    products := api.Group("/products")
    products.GET("", getProducts)
    products.GET("/:id", getProduct)
    products.POST("", createProduct)
    products.PUT("/:id", updateProduct)
    products.DELETE("/:id", deleteProduct)
}
```

## Step 4: Update Docker Compose Configuration

Update your service in `apps.yaml`:

```yaml
services:
  product-service:
    image: ghcr.io/raphaeldiscky/go-micro-template/product-service
    container_name: product-service
    environment:
      - HTTP_SERVER_PORT=8082
      - CONSUL_ADDRESS=consul:8500
      - SERVICE_ADDRESS=product-service # Container name for internal communication
      - DB_HOST=postgres-product
      - DB_PORT=5432
      # ... other environment variables
    networks:
      - go-micro-template
    depends_on:
      consul:
        condition: service_healthy
      postgres-product:
        condition: service_healthy
    # No need for port mapping - Traefik will handle routing
    # ports:
    #   - "8082:8082"

    # Traefik will discover this service automatically through Consul
    # But you can also add Docker labels as backup:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.product-service.rule=Host(`products.localhost`)"
      - "traefik.http.services.product-service.loadbalancer.server.port=8082"
    restart: unless-stopped
```

## Step 5: Test the Integration

1. **Start the infrastructure:**

   ```bash
   # Create network
   docker network create go-micro-template

   # Start Consul and Traefik
   docker-compose -f deployments/docker-compose/api-gateway-enhanced.yaml up -d

   # Start your services
   docker-compose -f deployments/docker-compose/apps.yaml up -d product-service
   ```

2. **Verify service registration:**

   ```bash
   # Check Consul services
   curl http://localhost:8500/v1/catalog/services

   # Check specific service health
   curl http://localhost:8500/v1/health/service/product-service
   ```

3. **Test Traefik routing:**

   ```bash
   # Access through Traefik (should route to your service)
   curl http://localhost:80/api/v1/products -H "Host: products.localhost"

   # Check Traefik dashboard
   open http://localhost:9000
   ```

## Step 6: Access Patterns

After setup, you can access your services in multiple ways:

1. **Through Traefik (recommended for external access):**

   ```bash
   curl http://localhost:80/api/v1/products -H "Host: products.localhost"
   ```

2. **Direct access (for internal service-to-service communication):**

   ```bash
   curl http://product-service:8082/api/v1/products
   ```

3. **Through API Gateway (for complex routing/business logic):**
   ```bash
   curl http://localhost:80/api/v1/products -H "Host: api.localhost"
   ```

## Benefits of This Setup

1. **Automatic Service Discovery:** New services are automatically discovered by Traefik
2. **Health Monitoring:** Unhealthy services are automatically removed from load balancing
3. **Load Balancing:** Multiple instances of the same service are automatically load balanced
4. **No Port Management:** Services don't need to expose ports - Traefik handles routing
5. **Flexible Routing:** Can route based on host, path, headers, etc.
6. **Monitoring:** Built-in metrics and health checks

This integration provides a robust, scalable foundation for your microservices architecture with minimal configuration overhead.
