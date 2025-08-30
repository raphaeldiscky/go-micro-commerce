# Traefik Routing and Health Check Configuration

## Overview

This configuration sets up Traefik as a reverse proxy/load balancer that routes traffic between:

- **API Gateway** (port 8080) - Main entry point for all API requests
- **Product Service** (port 8082) - Internal microservice
- Other services as needed

## Architecture Flow

```
Internet/Client → Traefik (port 80) → API Gateway (port 8080) → Product Service (port 8082)
                     ↓
                Direct access for testing
                     ↓
                Product Service (port 8082)
```

## Configuration Files

### 1. Static Configuration (api-routes.yml)

- Defines fixed routes for services
- Configures health checks and load balancing
- Sets up middleware for rate limiting and headers

### 2. Dynamic Configuration (dynamic-routes.yml)

- More advanced routing with priorities
- CORS headers and authentication middleware
- Better suited for production environments

### 3. Docker Compose Labels

- Alternative configuration using Docker labels
- Automatic service discovery
- Simpler for basic setups

## Health Check Details

### How Health Checks Work

1. **Traefik Health Checks**:

   - Traefik periodically calls `/health` on each service
   - Services returning HTTP 200 are considered healthy
   - Unhealthy services are removed from load balancing

2. **Docker Health Checks**:

   - Docker monitors container health independently
   - Uses `wget --spider` to check `/health` endpoints
   - Containers marked as unhealthy won't receive traffic

3. **Service Registration Health Checks**:
   - Consul also monitors service health
   - API Gateway can query healthy service instances
   - Enables dynamic service discovery

### Health Check Endpoints

- **API Gateway**: `http://api-gateway:8080/health`
- **Product Service**: `http://product-service:8082/health`

Both return JSON responses like:

```json
{
  "status": "healthy",
  "service": "product-service",
  "timestamp": "2025-08-05T10:30:00Z"
}
```

## Access Points

### External Access (through Traefik)

- **Main API**: `http://api.localhost` → API Gateway → Routes to services
- **Direct Product Service**: `http://products.localhost` → Product Service (for testing)
- **Traefik Dashboard**: `http://traefik.localhost:9000`

### Internal Access (container-to-container)

- **API Gateway**: `http://api-gateway:8080`
- **Product Service**: `http://product-service:8082`

## Starting the Services

The correct order is:

```bash
# 1. Start infrastructure (PostgreSQL, Redis, Kafka, Consul, Traefik)
task start_infra

# 2. Start applications
task start_apps

# 3. Start monitoring (optional)
task start_monitoring
```

## Testing Health Checks

```bash
# Test through Traefik
curl http://api.localhost/health
curl http://products.localhost/health

# Test directly (if ports are exposed)
curl http://localhost:8080/health
curl http://localhost:8082/health

# Check Traefik dashboard
open http://localhost:9000
```

## Troubleshooting

### Common Issues

1. **Service Dependencies**: Make sure infrastructure services are running first
2. **Network Issues**: Ensure all services are on the `go-micro-commerce` network
3. **Health Check Failures**: Check if services are responding on `/health`
4. **DNS Resolution**: Use container names (not localhost) for inter-service communication

### Useful Commands

```bash
# Check service health
docker ps --format "table {{.Names}}\t{{.Status}}"

# Check Traefik logs
docker logs traefik

# Check network connectivity
docker exec api-gateway ping product-service

# View Consul services
curl http://localhost:8500/v1/agent/services
```

## Production Considerations

1. **HTTPS**: Configure SSL certificates and redirect HTTP to HTTPS
2. **Authentication**: Implement proper JWT validation middleware
3. **Rate Limiting**: Adjust limits based on expected traffic
4. **Health Check Intervals**: Balance between responsiveness and resource usage
5. **Circuit Breakers**: Add circuit breaker patterns for service failures
6. **Monitoring**: Set up Prometheus metrics collection from health endpoints
