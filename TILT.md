# Tilt Development Guide

Complete guide to using Tilt for local Kubernetes development with go-micro-commerce.

## What is Tilt?

Tilt is a developer tool that makes Kubernetes development fast and efficient. It provides:

- **Auto-rebuild**: Detects code changes and rebuilds only what changed
- **Hot-reload**: Syncs code to running containers without full rebuilds
- **Unified UI**: Monitor all services, logs, and builds in one place
- **Smart dependencies**: Starts services in the correct order

## Quick Start

```bash
# One-time setup
task k8s:create_cluster    # Create Kind cluster
task k8s:create_secrets    # Generate JWT keys

# Daily workflow
task tilt:up               # Start everything
# Make code changes → auto-reload
task tilt:down             # Stop Tilt
```

## What Gets Deployed?

When you run `tilt up`, the following resources are deployed automatically:

### Infrastructure (Deployed First)

**Databases** (9 separate PostgreSQL instances):

- postgresql-auth (auth_db)
- postgresql-product (product_db)
- postgresql-order (order_db)
- postgresql-payment (payment_db)
- postgresql-cart (cart_db)
- postgresql-fulfillment (fulfillment_db)
- postgresql-notification (notification_db)
- postgresql-search (search_db)
- postgresql-chat (chat_db)

**Caching & Messaging**:

- Redis Cluster (6 nodes: 3 masters + 3 replicas)
- Kafka Cluster (3 brokers with KRaft mode)

**Monitoring Stack**:

- Prometheus (metrics)
- Grafana (dashboards)
- Loki (log aggregation)
- Tempo (distributed tracing)
- OpenTelemetry Collector (telemetry pipeline)

**Dev Tools**:

- Kafka UI
- Redis Insight
- MailHog (SMTP testing)

### Application Services (11 microservices)

Deployed after infrastructure is ready:

1. api-gateway
2. graphql-gateway
3. auth-service
4. product-service
5. order-service
6. payment-service
7. cart-service
8. fulfillment-service
9. notification-service
10. search-service
11. chat-service

## Tilt UI Overview

Access at: **http://localhost:10350**

### Resource Groups

Resources are organized into logical groups:

- **infra-db**: All PostgreSQL databases
- **infra-cache**: Redis Cluster + Redis Insight
- **infra-messaging**: Kafka + Kafka UI
- **monitoring**: Prometheus, Grafana, Loki, Tempo, OTEL
- **dev-tools**: MailHog, admin UIs
- **apps**: All 11 microservices

### Resource States

- 🟢 **Green**: Healthy and running
- 🟡 **Yellow**: Building or updating
- 🔴 **Red**: Error or crash loop
- ⚪ **Gray**: Pending or disabled

### Key Features

**1. Live Logs**: Click any resource to see real-time logs

**2. Build Status**: See which services are building and why

**3. Trigger Rebuilds**: Press spacebar or click "Force Update"

**4. Resource Dependencies**: Visual graph showing startup order

## Development Workflow

### Making Code Changes

1. Edit code in any service (e.g., `product-service/internal/handler/product.go`)
2. Save the file
3. Tilt automatically:
   - Detects the change
   - Syncs code to the pod (if live_update is configured)
   - Rebuilds the binary inside the container
   - Restarts the service
   - Shows logs in UI

**Typical reload time**: 2-5 seconds

### Viewing Logs

**Option 1: Tilt UI**

- Click on any resource in the Tilt UI
- Logs appear in the right panel
- Auto-scroll, search, and filter available

**Option 2: kubectl**

```bash
task tilt:logs SERVICE=product-service
# or
kubectl logs -f -l app=product-service -n default
```

### Debugging

**Check Pod Status**:

```bash
task k8s:status
# or
kubectl get pods -n default
```

**Describe Pod**:

```bash
kubectl describe pod -l app=product-service -n default
```

**Exec into Pod**:

```bash
kubectl exec -it <pod-name> -n default -- /bin/sh
```

**Check Database Connection**:

```bash
kubectl exec -it postgresql-product-0 -n default -- psql -U postgres -d product_db
```

## Port Forwards

### Automatically Accessible UIs

All UIs are automatically port-forwarded by Tilt:

| Service       | URL                    | Credentials |
| ------------- | ---------------------- | ----------- |
| Tilt UI       | http://localhost:10350 | -           |
| Grafana       | http://localhost:3000  | admin/admin |
| Prometheus    | http://localhost:9090  | -           |
| Kafka UI      | http://localhost:8090  | -           |
| Redis Insight | http://localhost:5540  | -           |
| MailHog       | http://localhost:8025  | -           |

### Application Services

To access application services directly, uncomment port forwards in `Tiltfile`:

```python
# Tiltfile
k8s_resource('product-service', port_forwards=['8082:8080'])
```

## Resource Requirements

Recommended system resources for full stack:

| Component          | Memory        | CPU          | Notes                            |
| ------------------ | ------------- | ------------ | -------------------------------- |
| PostgreSQL (9 DBs) | ~4.5 Gi       | ~1 CPU       | 512Mi each                       |
| Redis Cluster      | ~1.5 Gi       | ~0.5 CPU     | 6 nodes                          |
| Kafka              | ~3 Gi         | ~1 CPU       | 3 brokers                        |
| Monitoring         | ~2 Gi         | ~1 CPU       | Prometheus, Grafana, Loki, Tempo |
| Apps (11 services) | ~2.8 Gi       | ~2 CPU       | 256Mi each                       |
| **Total**          | **~14-16 Gi** | **~5-6 CPU** |                                  |

**Minimum**: 16GB RAM, 4 CPU cores
**Recommended**: 32GB RAM, 8 CPU cores

## Troubleshooting

### Issue: Tilt says "cluster not ready"

**Solution**:

```bash
kubectl cluster-info
# If cluster is down:
task k8s:create_cluster
```

### Issue: Pods stuck in "ImagePullBackOff"

**Solution**: Images need to be built or loaded

```bash
# Check if local registry is accessible
docker ps | grep kind-registry
# Rebuild images
tilt up  # Tilt builds automatically
```

### Issue: Service fails to connect to database

**Solution**: Check database is ready

```bash
kubectl get pods | grep postgresql
# All should show 1/1 READY
# If not, check logs:
kubectl logs postgresql-product-0
```

### Issue: Out of memory

**Solution**: Reduce replicas or disable monitoring stack

```bash
# Edit Tiltfile and comment out monitoring section
# Then restart:
tilt down
tilt up
```

### Issue: Slow builds

**Solution**: Enable live_update in Tiltfile (already configured for Go services)

## Advanced Configuration

### Customizing Resources

Edit `Tiltfile` to:

- Add/remove services
- Change resource limits
- Configure port forwards
- Enable/disable resource groups

### Selective Deployment

Deploy only specific resources:

```python
# Tiltfile
config.set_enabled_resources(['product-service', 'postgresql-product'])
```

Or use Tilt's built-in filtering:

```bash
tilt up postgresql-product product-service
```

### Environment Variables

Override configs via ConfigMaps in `overlays/local/kustomization.yaml`

## CI/CD Integration

Tilt is for **local development only**. For CI/CD:

```bash
# CI pipelines should use:
task build           # Build production images
task lint            # Code quality
task test            # Run tests
kubectl apply -k overlays/dev  # Deploy to dev cluster
```

## Tips & Best Practices

1. **Keep Tilt Running**: Leave `tilt up` running during your dev session
2. **Use Resource Groups**: Filter by label in Tilt UI to focus on specific services
3. **Check Dependencies**: If a service won't start, check its dependencies in Tilt UI
4. **Clean Slate**: If things get messy, `tilt down` and `tilt up` again
5. **PVC Cleanup**: Databases persist between sessions. To reset: `kubectl delete pvc --all`
6. **Watch Build Times**: Tilt shows build duration - optimize slow builds

## Getting Help

- **Tilt Docs**: https://docs.tilt.dev
- **Project Issues**: Check K8s README and CLAUDE.md
- **Logs**: Always check Tilt UI logs first
- **Community**: Tilt has an active Slack community

## Next Steps

- Explore Grafana dashboards for service metrics
- Set up alerts in Prometheus
- Customize Tiltfile for your workflow
- Add custom Tilt extensions

Happy developing! 🚀
