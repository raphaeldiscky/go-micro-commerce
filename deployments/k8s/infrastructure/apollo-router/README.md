# Apollo Router Deployment for Kubernetes

This directory contains the Kubernetes deployment configuration for the Apollo Router using the official Apollo Router Helm chart with Federation v2.

## Overview

The GraphQL Gateway uses Apollo Router v2.7.0 to federate 6 microservice subgraphs:

- **auth-service** (port 8081)
- **chat-service** (port 8088)
- **notification-service** (port 8086)
- **order-service** (port 8083)
- **cart-service** (port 8089)
- **payment-service** (port 8084)

## Architecture

### Deployment Method

- **Helm Chart**: Official Apollo Router Helm chart (`apollo/router`)
- **Configuration**: ConfigMaps for router config, Rhai scripts, and supergraph schema
- **Scaling**: HorizontalPodAutoscaler (2-10 replicas based on CPU/Memory)
- **Observability**: OpenTelemetry tracing, Prometheus metrics, Grafana dashboards

### Components

1. **ConfigMaps**
   - `apollo-router-config`: Router configuration (CORS, headers, telemetry, rate limiting)
   - `apollo-router-rhai`: Rhai script for Set-Cookie header propagation
   - `apollo-supergraph-schema`: Pre-composed supergraph schema

2. **Helm Chart**
   - `values.yaml`: Apollo Router Helm chart configuration
   - Replica count, resources, health checks, security context

3. **Schema Composition Job**
   - `job-schema-composition.yaml`: K8s Job using Rover CLI
   - Fetches schemas from subgraph services
   - Composes supergraph schema
   - Updates ConfigMap with new schema

4. **Observability**
   - `servicemonitor.yaml`: Prometheus ServiceMonitor for metrics
   - GraphQL-specific metrics on port 9091
   - Health checks on port 4088
   - OTEL tracing to `otel-collector:4317`

## Files

```
apollo-router/
├── README.md                           # This file
├── values.yaml                         # Apollo Router Helm values
├── configmap-router.yaml               # Router configuration
├── configmap-rhai.yaml                 # Rhai script for cookie propagation
├── configmap-supergraph-schema.yaml    # Supergraph schema (generated)
├── create-supergraph-configmap.sh      # Script to generate schema ConfigMap
├── supergraph.yaml                     # Supergraph composition config (K8s DNS)
├── job-schema-composition.yaml         # Schema composition Job + RBAC
└── servicemonitor.yaml                 # Prometheus ServiceMonitor
```

## Deployment

### Prerequisites

1. Kubernetes cluster (Kind, MicroK8s, or any K8s cluster)
2. Helm 3.x installed
3. Kubectl configured
4. All subgraph services deployed and healthy

### Deploy with Tilt

The Apollo Router is automatically deployed when running `tilt up`:

```bash
tilt up
```

The Tiltfile:

1. Adds Apollo Helm repository
2. Deploys ConfigMaps
3. Deploys Apollo Router via Helm
4. Waits for all subgraph services
5. Port-forwards 4000 (GraphQL) and 9091 (metrics)

### Manual Deployment

If deploying without Tilt:

```bash
# Add Apollo Helm repository
helm repo add apollo https://helm.apollographql.com
helm repo update

# Create namespace (optional)
kubectl create namespace default

# Apply ConfigMaps
kubectl apply -f configmap-router.yaml
kubectl apply -f configmap-rhai.yaml

# Generate and apply supergraph schema ConfigMap
./create-supergraph-configmap.sh
kubectl apply -f configmap-supergraph-schema.yaml

# Install Apollo Router with Helm
helm install apollo-router apollo/router \
  --values values.yaml \
  --namespace default

# Apply ServiceMonitor for Prometheus
kubectl apply -f servicemonitor.yaml
```

## Schema Composition

### Initial Setup

The supergraph schema is pre-generated and stored in `configmap-supergraph-schema.yaml`.

### Updating Schema

When subgraph schemas change, run the schema composition Job:

```bash
# Run schema composition Job
kubectl apply -f job-schema-composition.yaml

# Check Job status
kubectl get jobs
kubectl logs job/apollo-schema-composition

# The Job will:
# 1. Fetch schemas from all subgraph services
# 2. Compose supergraph using Rover CLI
# 3. Update apollo-supergraph-schema ConfigMap
# 4. Restart Apollo Router pods (automatically via ConfigMap change)
```

### Manual Schema Composition

To manually compose the schema locally:

```bash
# Fetch schemas from subgraphs (requires services running)
cd graphql-gateway
./scripts/compose-supergraph.sh

# Generate ConfigMap from updated schema
cd ../deployments/k8s/infrastructure/apollo-router
./create-supergraph-configmap.sh

# Apply updated ConfigMap
kubectl apply -f configmap-supergraph-schema.yaml

# Restart Apollo Router pods
kubectl rollout restart deployment apollo-router
```

## Configuration

### Router Configuration

Edit `configmap-router.yaml` to update:

- CORS settings
- Header propagation rules
- OpenTelemetry configuration
- Rate limiting
- Timeouts

After changes:

```bash
kubectl apply -f configmap-router.yaml
kubectl rollout restart deployment apollo-router
```

### Helm Values

Edit `values.yaml` to update:

- Replica count
- Resource limits
- Health check settings
- Security context
- Environment variables

After changes:

```bash
helm upgrade apollo-router apollo/router \
  --values values.yaml \
  --namespace default
```

### Subgraph URLs

Subgraph routing URLs are configured in `supergraph.yaml` using Kubernetes Service DNS:

```yaml
subgraphs:
  auth-service:
    routing_url: http://auth-service:8081/graph
```

Update `supergraph.yaml` and re-run schema composition if URLs change.

## Observability

### Metrics

Apollo Router exposes Prometheus metrics on port 9091:

```bash
# Port-forward metrics endpoint
kubectl port-forward svc/apollo-router-metrics 9091:9091

# Access metrics
curl http://localhost:9091/metrics
```

**Key Metrics:**

- `apollo_router_http_requests_total` - Total HTTP requests
- `apollo_router_http_request_duration_seconds` - Request latency
- `apollo_router_graphql_error` - GraphQL errors
- `apollo_router_cache_hit_count` - Cache hit rate

### Health Checks

Health endpoint on port 4088:

```bash
# Port-forward health endpoint
kubectl port-forward deployment/apollo-router 4088:4088

# Check health
curl http://localhost:4088/health
```

### Tracing

Apollo Router sends OpenTelemetry traces to `otel-collector:4317`:

- View traces in Grafana Tempo
- Access Grafana at http://localhost:3000

### Logs

```bash
# View Apollo Router logs
kubectl logs -f deployment/apollo-router

# View schema composition Job logs
kubectl logs job/apollo-schema-composition
```

## Troubleshooting

### Router Not Starting

1. Check ConfigMaps exist:

```bash
kubectl get configmaps | grep apollo
```

2. Check pod logs:

```bash
kubectl logs deployment/apollo-router
```

3. Verify subgraph services are healthy:

```bash
kubectl get pods | grep -E 'auth|chat|notification|order|cart|payment'
```

### Schema Composition Fails

1. Check subgraph service endpoints:

```bash
kubectl get svc
```

2. Check Job logs:

```bash
kubectl logs job/apollo-schema-composition
```

3. Verify RBAC permissions:

```bash
kubectl get role apollo-schema-composer
kubectl get rolebinding apollo-schema-composer
```

### Subgraph Unreachable

1. Check service DNS resolution:

```bash
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://auth-service:8081/graph
```

2. Verify subgraph health:

```bash
kubectl exec -it deployment/apollo-router -- \
  wget -O- http://auth-service:8081/health
```

### High Latency

1. Check resource usage:

```bash
kubectl top pods -l app.kubernetes.io/name=router
```

2. Scale replicas:

```bash
kubectl scale deployment apollo-router --replicas=5
```

3. Check HPA status:

```bash
kubectl get hpa
```

## Scaling

### Manual Scaling

```bash
# Scale to 5 replicas
kubectl scale deployment apollo-router --replicas=5
```

### Auto-Scaling (HPA)

HPA is configured in `values.yaml`:

- **Min replicas**: 2
- **Max replicas**: 10
- **CPU threshold**: 70%
- **Memory threshold**: 80%

Adjust in `values.yaml` and upgrade Helm chart.

## Security

### Non-Root User

Apollo Router runs as non-root user (UID 1000) with:

- Read-only root filesystem
- Dropped capabilities
- Security context constraints

### RBAC

Schema composition Job has minimal RBAC:

- Read/write ConfigMaps
- Only in default namespace

## Ports

- **4000**: GraphQL endpoint (HTTP)
- **4088**: Health check endpoint
- **9091**: Prometheus metrics

## Dependencies

Apollo Router requires:

- All 6 subgraph services running
- `otel-collector` for tracing
- `redis-cluster` (optional, for caching)
- `kafka-cluster` (used by subgraphs)

## Further Reading

- [Apollo Router Documentation](https://www.apollographql.com/docs/router/)
- [Apollo Federation](https://www.apollographql.com/docs/federation/)
- [Rover CLI](https://www.apollographql.com/docs/rover/)
- [Apollo Router Helm Chart](https://github.com/apollographql/helm-charts)
