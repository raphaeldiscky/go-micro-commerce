# Consul and Traefik Setup Guide

This guide explains how to set up Consul for service discovery and Traefik as an API Gateway with automatic service discovery in your Go microservices architecture.

## Architecture Overview

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    Client   │────│   Traefik   │────│  Services   │
│             │    │ (Load Bal.) │    │             │
└─────────────┘    └──────┬──────┘    └─────────────┘
                          │
                   ┌──────▼──────┐
                   │   Consul    │
                   │ (Discovery) │
                   └─────────────┘
```

**Flow:**

1. **Services** register themselves with **Consul** on startup
2. **Traefik** discovers services through **Consul Catalog**
3. **Clients** make requests to **Traefik**
4. **Traefik** routes requests to healthy service instances

## Current Implementation Analysis

### API Gateway Service

Your current API Gateway (`api-gateway`) acts as a **custom proxy** that:

- Uses Consul for service discovery via `ConsulServiceDiscovery`
- Implements circuit breakers and load balancing
- Proxies requests to backend services manually

### Proposed Architecture Enhancement

#### Option 1: Traefik as Primary Gateway (Recommended)

- **Traefik** handles all external traffic and load balancing
- **API Gateway** becomes an internal service for business logic
- Services register with Consul and get auto-discovered by Traefik

#### Option 2: Hybrid Approach

- **Traefik** for external traffic and SSL termination
- **API Gateway** for internal routing and business logic
- Both use Consul for service discovery

## Setup Instructions

### 1. Consul Configuration

Your current Consul setup in `api-gateway.yaml` is good. Here's an enhanced version:

```yaml
# Enhanced Consul configuration
consul:
  image: consul:1.15
  container_name: consul
  ports:
    - "8500:8500" # HTTP API
    - "8501:8501" # HTTPS API
    - "8502:8502" # gRPC API
    - "8600:8600/udp" # DNS
  environment:
    - CONSUL_BIND_INTERFACE=eth0
    - CONSUL_CLIENT_INTERFACE=eth0
  command: >
    consul agent
    -server
    -bootstrap-expect=1
    -ui
    -client=0.0.0.0
    -bind=0.0.0.0
    -data-dir=/consul/data
    -enable-script-checks=true
    -grpc-port=8502
    -log-level=INFO
    -enable-local-script-checks=true
    -connect=true
  volumes:
    - consul_data:/consul/data
    - ./consul/config:/consul/config:ro
  networks:
    - go-micro-commerce
  healthcheck:
    test: ["CMD", "consul", "members"]
    interval: 30s
    timeout: 10s
    retries: 3
```

### 2. Traefik Configuration

#### Enhanced Traefik Setup

```yaml
traefik:
  image: traefik:v3.5.0
  container_name: traefik
  ports:
    - "80:80" # HTTP
    - "443:443" # HTTPS (for production)
    - "9000:8080" # Dashboard
  command:
    # API Configuration
    - --api.dashboard=true
    - --api.insecure=true

    # Entry Points
    - --entrypoints.web.address=:80
    - --entrypoints.websecure.address=:443

    # Consul Catalog Provider
    - --providers.consulCatalog.endpoint.address=consul:8500
    - --providers.consulCatalog.exposedByDefault=false
    - --providers.consulCatalog.defaultRule=Host(`{{ .Name }}.localhost`)
    - --providers.consulCatalog.connectAware=true

    # File Provider for Static Configuration
    - --providers.file.directory=/etc/traefik/dynamic
    - --providers.file.watch=true

    # Metrics
    - --metrics.prometheus=true
    - --metrics.prometheus.addEntryPointsLabels=true
    - --metrics.prometheus.addServicesLabels=true

    # Tracing (optional)
    - --tracing.jaeger=true
    - --tracing.jaeger.samplingParam=1.0
    - --tracing.jaeger.localAgentHostPort=jaeger:6831

    # Logs
    - --log.level=INFO
    - --accesslog=true

  volumes:
    - ./traefik/config:/etc/traefik/dynamic:ro
    - /var/run/docker.sock:/var/run/docker.sock:ro # For Docker provider (optional)
  networks:
    - go-micro-commerce
  depends_on:
    consul:
      condition: service_healthy
  healthcheck:
    test: ["CMD", "traefik", "healthcheck", "--ping"]
    interval: 30s
    timeout: 10s
    retries: 3
  labels:
    - "traefik.enable=true"
    - "traefik.http.routers.traefik.rule=Host(`traefik.localhost`)"
    - "traefik.http.routers.traefik.service=api@internal"
```

### 3. Service Registration with Consul

Each microservice needs to register itself with Consul. Here's how to modify your services:

#### Example: Product Service Registration

Add to your service startup code:

```go
// pkg/consul/registration.go
package consul

import (
    "fmt"
    "net"
    "strconv"

    "github.com/hashicorp/consul/api"
)

type ServiceRegistration struct {
    client *api.Client
}

func NewServiceRegistration(consulAddr string) (*ServiceRegistration, error) {
    config := api.DefaultConfig()
    config.Address = consulAddr

    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }

    return &ServiceRegistration{client: client}, nil
}

func (s *ServiceRegistration) Register(serviceName, serviceID, address string, port int, tags []string) error {
    registration := &api.AgentServiceRegistration{
        ID:      serviceID,
        Name:    serviceName,
        Tags:    tags,
        Address: address,
        Port:    port,
        Check: &api.AgentServiceCheck{
            HTTP:                           fmt.Sprintf("http://%s:%d/health", address, port),
            Interval:                       "30s",
            Timeout:                        "10s",
            DeregisterCriticalServiceAfter: "60s",
        },
        Meta: map[string]string{
            "version": "1.0.0",
        },
    }

    return s.client.Agent().ServiceRegister(registration)
}

func (s *ServiceRegistration) Deregister(serviceID string) error {
    return s.client.Agent().ServiceDeregister(serviceID)
}
```

#### Service Integration Example

```go
// In your service main.go
func main() {
    // ... existing setup ...

    // Register with Consul
    consulReg, err := consul.NewServiceRegistration("consul:8500")
    if err != nil {
        log.Fatal(err)
    }

    serviceID := fmt.Sprintf("%s-%s", serviceName, generateUniqueID())
    tags := []string{
        "api",
        "v1",
        "traefik.enable=true",
        "traefik.http.routers.product-service.rule=Host(`product-service.localhost`)",
        "traefik.http.services.product-service.loadbalancer.server.port=8082",
    }

    err = consulReg.Register("product-service", serviceID, getLocalIP(), 8082, tags)
    if err != nil {
        log.Fatal(err)
    }

    // Deregister on shutdown
    defer consulReg.Deregister(serviceID)

    // ... start your server ...
}
```

### 4. Traefik Dynamic Configuration

Create Traefik configuration files:

```yaml
# traefik/config/api-routes.yml
http:
  routers:
    api-gateway:
      rule: "Host(`api.localhost`)"
      service: api-gateway-service
      entryPoints:
        - web
      middlewares:
        - rate-limit
        - auth-headers

  services:
    api-gateway-service:
      loadBalancer:
        servers:
          - url: "http://api-gateway:8080"
        healthCheck:
          path: /health
          interval: 30s
          timeout: 10s

  middlewares:
    rate-limit:
      rateLimit:
        burst: 100
        average: 10

    auth-headers:
      headers:
        customRequestHeaders:
          X-Forwarded-Proto: https
```

```yaml
# traefik/config/consul-services.yml
http:
  # This will be populated automatically by Consul Catalog provider
  # Services registered in Consul will appear here automatically
```

### 5. Modified Service Configuration

Update your `apps.yaml` to work with Traefik:

```yaml
services:
  api-gateway:
    image: ghcr.io/raphaeldiscky/go-micro-commerce/api-gateway
    container_name: api-gateway
    environment:
      - SERVER_PORT=8080
      - CONSUL_ADDRESS=consul:8500
      # ... other env vars
    networks:
      - go-micro-commerce
    labels:
      # Traefik labels for service discovery
      - "traefik.enable=true"
      - "traefik.http.routers.api-gateway.rule=Host(`api.localhost`)"
      - "traefik.http.routers.api-gateway.entrypoints=web"
      - "traefik.http.services.api-gateway.loadbalancer.server.port=8080"
      - "consul.tags=api,gateway,v1"
    depends_on:
      consul:
        condition: service_healthy

  product-service:
    image: ghcr.io/raphaeldiscky/go-micro-commerce/product-service
    container_name: product-service
    environment:
      - HTTP_SERVER_PORT=8082
      - CONSUL_ADDRESS=consul:8500
      # ... other env vars
    networks:
      - go-micro-commerce
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.product-service.rule=Host(`products.localhost`)"
      - "traefik.http.services.product-service.loadbalancer.server.port=8082"
      - "consul.tags=api,products,v1"
```

## Implementation Plan

### Phase 1: Basic Setup

1. ✅ Consul is already configured
2. ✅ Traefik is configured with Consul integration
3. 🔄 Create Traefik dynamic configuration files
4. 🔄 Test Consul and Traefik integration

### Phase 2: Service Registration

1. 🔄 Create service registration utility
2. 🔄 Modify each service to register with Consul
3. 🔄 Add health check endpoints to services
4. 🔄 Test service discovery

### Phase 3: API Gateway Enhancement

1. 🔄 Update API Gateway to work alongside Traefik
2. 🔄 Configure routing rules in Traefik
3. 🔄 Test end-to-end request flow
4. 🔄 Implement monitoring and observability

### Phase 4: Production Readiness

1. 🔄 Add SSL/TLS termination in Traefik
2. 🔄 Configure proper security headers
3. 🔄 Set up monitoring and alerting
4. 🔄 Load testing and performance optimization

## Testing the Setup

### 1. Start the Infrastructure

```bash
# Create the external network first
docker network create go-micro-commerce

# Start Consul and Traefik
docker-compose -f deployments/docker-compose/api-gateway.yaml up -d

# Verify Consul is running
curl http://localhost:8500/v1/status/leader

# Verify Traefik dashboard
open http://localhost:9000
```

### 2. Test Service Discovery

```bash
# Register a test service with Consul
curl -X PUT http://localhost:8500/v1/agent/service/register \
  -d '{
    "ID": "test-service-1",
    "Name": "test-service",
    "Tags": ["api", "v1"],
    "Address": "127.0.0.1",
    "Port": 8080,
    "Check": {
      "HTTP": "http://127.0.0.1:8080/health",
      "Interval": "30s"
    }
  }'

# Check if Traefik discovered the service
curl http://localhost:9000/api/http/services
```

### 3. Test Request Routing

```bash
# Test direct access to API Gateway
curl http://localhost:80 -H "Host: api.localhost"

# Test service routing through Traefik
curl http://localhost:80 -H "Host: products.localhost"
```

## Monitoring and Observability

### Consul Health Checks

- Services automatically register health checks
- Unhealthy services are removed from load balancing
- Consul UI shows service health status

### Traefik Metrics

- Prometheus metrics available at `/metrics`
- Request rate, response time, error rate tracking
- Dashboard integration with Grafana

### Logging

- Consul logs service registration/deregistration
- Traefik access logs for request tracing
- Application logs through structured logging

## Troubleshooting

### Common Issues

1. **Service Not Discovered**

   - Check Consul service registration
   - Verify Traefik Consul provider configuration
   - Check service health status

2. **Routing Not Working**

   - Verify Traefik router rules
   - Check service labels/tags
   - Test DNS resolution

3. **Health Check Failures**
   - Ensure `/health` endpoint exists
   - Check network connectivity
   - Verify health check configuration

### Debug Commands

```bash
# Check Consul services
curl http://localhost:8500/v1/catalog/services

# Check Traefik configuration
curl http://localhost:9000/api/http/routers

# Check service health
curl http://localhost:8500/v1/health/service/product-service

# Traefik logs
docker logs traefik

# Consul logs
docker logs consul
```

## Next Steps

1. **Implement the Phase 1 changes** in the configuration files
2. **Create the service registration utility** for your microservices
3. **Test the integration** with a single service first
4. **Gradually migrate** all services to use Consul registration
5. **Add monitoring and alerting** for production deployment

This setup provides a robust, scalable service discovery and load balancing solution that integrates well with your existing microservices architecture.
