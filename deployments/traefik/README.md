# Traefik API Gateway Setup

This directory contains the Traefik configuration for the microservices API gateway.

## 📁 Directory Structure

```
deployments/traefik/
├── config/
│   └── dynamic.yml         # Dynamic configuration (middleware, routing rules)
├── letsencrypt/           # SSL certificates storage
└── traefik.yml           # Static configuration (entry points, providers)
```

## 🚀 Quick Start

1. **Start with Traefik:**

   ```bash
   # Use the enhanced script with Traefik
   ./scripts/microservices-traefik.sh start
   ```

2. **Access Services:**
   - **Traefik Dashboard:** http://localhost:8080
   - **Product API:** http://localhost/api/v1/products
   - **Seller API:** http://localhost/api/v1/sellers
   - **Prometheus:** http://localhost:9091
   - **Grafana:** http://localhost:3001
   - **Kafka UI:** http://localhost:8081

## 🔧 Configuration Files

### `traefik.yml` (Static Configuration)

- Entry points (HTTP/HTTPS/gRPC)
- Service discovery via Docker labels
- SSL/TLS configuration with Let's Encrypt
- Logging and metrics configuration

### `config/dynamic.yml` (Dynamic Configuration)

- Middleware definitions (CORS, rate limiting, auth)
- Advanced routing rules
- Load balancing configuration
- Circuit breaker and retry policies

## 🌐 Routing Rules

Traefik automatically discovers services through Docker labels:

```yaml
# Product Service
- "traefik.http.routers.product-service.rule=Host(`localhost`) && PathPrefix(`/api/v1/products`)"
- "traefik.http.routers.product-service.middlewares=product-stripprefix"

# Seller Service
- "traefik.http.routers.seller-service.rule=Host(`localhost`) && PathPrefix(`/api/v1/sellers`)"
- "traefik.http.routers.seller-service.middlewares=seller-stripprefix"
```

## 🛡️ Security Features

### Middleware Stack

- **CORS Headers:** Cross-origin resource sharing
- **Rate Limiting:** 100 requests/minute with burst of 200
- **Circuit Breaker:** Automatic failover on high error rates
- **Retry Logic:** 3 attempts for failed requests
- **Compression:** Gzip compression for responses

### SSL/TLS

- Automatic HTTPS redirection
- Let's Encrypt integration for free SSL certificates
- TLS 1.2/1.3 support with secure cipher suites

## 📊 Monitoring Integration

### Prometheus Metrics

Traefik exposes detailed metrics at `/metrics`:

- Request duration and counts
- Response status codes
- Service health status
- Entry point statistics

### Grafana Dashboards

Pre-configured dashboards for:

- Request rates and latency
- Error rates by service
- Traffic distribution
- Service availability

## 🔧 Advanced Configuration

### Custom Middleware

Add custom middleware in `config/dynamic.yml`:

```yaml
http:
  middlewares:
    custom-auth:
      basicAuth:
        users:
          - "admin:$2y$10$..."

    custom-headers:
      headers:
        customRequestHeaders:
          X-Custom-Header: "value"
```

### Load Balancing

Configure load balancing strategies:

```yaml
http:
  services:
    product-service:
      loadBalancer:
        sticky:
          cookie:
            name: "server"
        healthCheck:
          path: "/health"
          interval: "30s"
```

## 🐳 Docker Integration

Services are automatically discovered via Docker labels. Key labels:

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.service-name.rule=Host(`localhost`) && PathPrefix(`/api/path`)"
  - "traefik.http.services.service-name.loadbalancer.server.port=8080"
  - "traefik.http.middlewares.service-stripprefix.stripprefix.prefixes=/api/v1/service"
```

## 🚀 Deployment

### Development

```bash
# Start all services with Traefik
./scripts/microservices-traefik.sh start

# Check service health
./scripts/microservices-traefik.sh health

# View Traefik logs
./scripts/microservices-traefik.sh logs traefik
```

### Production

1. Update `traefik.yml`:

   ```yaml
   api:
     insecure: false # Disable insecure API

   certificatesResolvers:
     letsencrypt:
       acme:
         email: your-email@example.com
         # Remove staging server line
   ```

2. Set proper domain names in routing rules
3. Configure firewall rules for ports 80/443
4. Set up monitoring alerts

## 🔍 Troubleshooting

### Common Issues

1. **Service Not Reachable:**

   - Check if service has `traefik.enable=true` label
   - Verify routing rules in Traefik dashboard
   - Ensure service is on the same Docker network

2. **SSL Certificate Issues:**

   - Check Let's Encrypt rate limits
   - Verify domain DNS configuration
   - Review Traefik logs for ACME errors

3. **High Latency:**
   - Check service health endpoints
   - Review Prometheus metrics
   - Verify middleware configuration

### Useful Commands

```bash
# View Traefik configuration
curl http://localhost:8080/api/rawdata

# Check service discovery
curl http://localhost:8080/api/http/services

# View router configuration
curl http://localhost:8080/api/http/routers

# Monitor logs in real-time
./scripts/microservices-traefik.sh logs traefik
```

## 📚 References

- [Traefik Documentation](https://doc.traefik.io/traefik/)
- [Docker Provider](https://doc.traefik.io/traefik/providers/docker/)
- [Let's Encrypt](https://doc.traefik.io/traefik/https/acme/)
- [Middleware](https://doc.traefik.io/traefik/middlewares/overview/)
