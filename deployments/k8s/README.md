# Kubernetes Deployment

Production-grade Kubernetes deployment for go-micro-commerce microservices platform.

## 🚀 Quick Start

### Local Development (Kind)

```bash
# 1. Create Kind cluster
kind create cluster --name go-micro-commerce

# 2. Build images with local tag
TAG=local bash ./scripts/build.sh

# 3. Load images into Kind cluster
for svc in api-gateway auth-service product-service order-service payment-service cart-service fulfillment-service notification-service search-service chat-service graphql-gateway; do
  kind load docker-image localhost:5000/${svc}:local --name go-micro-commerce
done

# 4. Deploy services
kubectl apply -k overlays/local

# 5. Verify pods (will show errors until infrastructure is configured)
kubectl get pods -l environment=local

# 6. Check logs
kubectl logs -l environment=local --tail=20
```

**Note**: Services require infrastructure (PostgreSQL, Redis, Kafka) to run fully. See [Local Development with Host Infrastructure](#-local-development-with-host-infrastructure) section below.

### Production

See [docs/kubernetes-deployment.md](../../docs/kubernetes-deployment.md) for complete production deployment guide.

---

## 🏠 Local Development with Host Infrastructure

### Prerequisites

Ensure Docker Compose infrastructure is running on host:

```bash
cd deployments/docker-compose
docker-compose -f infra.yaml up -d
```

This starts:
- PostgreSQL (9 databases on ports 15431-15439)
- Redis Cluster (6 nodes on ports 6379-6384)
- Kafka (3 brokers on ports 9092-9094)
- OpenTelemetry Collector (port 4318)
- Temporal, Elasticsearch, etc.

### Configure Services to Use Host Infrastructure

**Option 1: Update ConfigMaps (Recommended for testing)**

Update each service's ConfigMap in `overlays/local/` to point to `host.docker.internal`:

```yaml
POSTGRES_HOST: "host.docker.internal"
POSTGRES_PORT: "15431"  # Service-specific port
REDIS_ADDRS: "host.docker.internal:6379,host.docker.internal:6380,host.docker.internal:6381,host.docker.internal:6382,host.docker.internal:6383,host.docker.internal:6384"
KAFKA_BROKERS: "host.docker.internal:9092,host.docker.internal:9093,host.docker.internal:9094"
```

**Option 2: Deploy Infrastructure to Kubernetes**

See [Infrastructure Setup](#-infrastructure-setup) section for deploying PostgreSQL, Redis, and Kafka using Helm charts.

### Create Required Secrets

```bash
# JWT keys for API Gateway
kubectl create secret generic api-gateway-jwt-keys \
  --from-file=public.pem=../../api-gateway/keys/public.pem

# JWT keys for Auth Service
kubectl create secret generic auth-service-jwt-keys \
  --from-file=private.pem=../../auth-service/keys/private.pem \
  --from-file=public.pem=../../auth-service/keys/public.pem

# Database passwords
kubectl create secret generic postgres-credentials \
  --from-literal=password=postgres
```

---

## 📁 Directory Structure

```
k8s/
├── base/                       # Base Kubernetes manifests
│   ├── api-gateway/            # API Gateway deployment, service, configmap, etc.
│   ├── auth-service/
│   ├── product-service/
│   ├── order-service/
│   ├── payment-service/
│   ├── cart-service/
│   ├── fulfillment-service/
│   ├── notification-service/
│   ├── search-service/
│   ├── chat-service/
│   └── graphql-gateway/
├── overlays/                   # Environment-specific configurations
│   ├── local/                  # Kind/Minikube (1 replica, minimal resources)
│   ├── dev/                    # Dev cluster (1 replica, moderate resources)
│   ├── staging/                # Staging (2 replicas, prod-like)
│   └── prod/                   # Production (3+ replicas, HA)
│       ├── us-east/
│       ├── us-west/
│       └── eu-west/
├── infrastructure/             # Infrastructure components
│   ├── traefik/                # Traefik Ingress Controller
│   ├── linkerd/                # Linkerd Service Mesh (TODO)
│   ├── argocd/                 # Argo CD GitOps (TODO)
│   ├── redis/                  # Redis Operator (TODO)
│   └── monitoring/             # LGTM stack (TODO)
├── generate-manifests.sh       # Script to generate service manifests
└── README.md                   # This file
```

---

## 🛠️ Service Manifests

Each service in `base/<service-name>/` contains:

- **deployment.yaml**: Kubernetes Deployment with resource limits, health checks
- **service.yaml**: Kubernetes Service (ClusterIP)
- **configmap.yaml**: Configuration via ConfigMap
- **secret.yaml**: Sensitive data (template - populate per environment)
- **serviceaccount.yaml**: Service Account for RBAC
- **hpa.yaml**: HorizontalPodAutoscaler for auto-scaling
- **pdb.yaml**: PodDisruptionBudget for HA
- **kustomization.yaml**: Kustomize configuration

---

## 🌍 Environments

### Local (Kind/Minikube)

- **Purpose**: Local development
- **Replicas**: 1
- **Resources**: Minimal (50m CPU, 64Mi RAM)
- **HPA**: Disabled
- **Namespace**: `default`

### Dev

- **Purpose**: Development cluster
- **Replicas**: 1
- **Resources**: Small (100m CPU, 128Mi RAM)
- **HPA**: 1-3 replicas
- **Namespace**: `dev`

### Staging

- **Purpose**: Pre-production testing
- **Replicas**: 2
- **Resources**: Medium (200m CPU, 256Mi RAM)
- **HPA**: 2-5 replicas
- **Namespace**: `staging`

### Production

- **Purpose**: Production workloads
- **Replicas**: 3+
- **Resources**: Large (500m CPU, 512Mi RAM)
- **HPA**: 3-20 replicas
- **Namespace**: `production`
- **Regions**: us-east, us-west, eu-west

---

## 🔧 Customization

### Generate Manifests

Regenerate all service manifests:

```bash
./generate-manifests.sh
```

### Customize Service

Edit base manifests:

```bash
vi base/api-gateway/deployment.yaml
```

### Environment-Specific Config

Add overrides in overlays:

```bash
vi overlays/prod/us-east/kustomization.yaml
```

---

## 🏗️ Infrastructure Setup

### 1. Traefik Ingress

```bash
helm install traefik traefik/traefik \
  -f infrastructure/traefik/values.yaml \
  -n traefik \
  --create-namespace
```

See [infrastructure/traefik/README.md](infrastructure/traefik/README.md)

### 2. Linkerd Service Mesh

```bash
linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -
linkerd check
```

### 3. Argo CD (GitOps)

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

---

## 🔍 Verification

### Check Pods

```bash
kubectl get pods -n production
```

### Check Services

```bash
kubectl get svc -n production
```

### Check Linkerd

```bash
linkerd -n production stat deployment
```

### View Logs

```bash
kubectl logs -n production deployment/api-gateway -f
```

---

## 🚦 Deployment Workflow

### Manual Deployment

```bash
kubectl apply -k overlays/prod/us-east
```

### GitOps with Argo CD

1. Commit changes to Git
2. Argo CD auto-syncs
3. Monitor in Argo UI

### Rolling Update

```bash
kubectl set image deployment/api-gateway \
  api-gateway=registry.io/api-gateway:v1.1.0 \
  -n production
```

### Rollback

```bash
kubectl rollout undo deployment/api-gateway -n production
```

---

## 📊 Monitoring

- **Metrics**: Prometheus scrapes `/metrics` endpoint
- **Logs**: Loki aggregates pod logs
- **Traces**: Tempo collects OpenTelemetry traces
- **Dashboards**: Grafana visualizes all data

Access Grafana:

```bash
kubectl port-forward -n monitoring svc/grafana 3000:3000
```

---

## 🔒 Security

- **mTLS**: Linkerd provides automatic service-to-service encryption
- **Secrets**: Use Sealed Secrets or External Secrets Operator
- **RBAC**: ServiceAccounts configured per service
- **Pod Security**: SecurityContext with non-root user
- **Network Policies**: Restrict traffic between namespaces

---

## 🐛 Troubleshooting

### ImagePullBackOff Errors

**Symptom**: Pods stuck in `ImagePullBackOff` or `ErrImagePull`

**Cause**: Docker images not loaded into Kind cluster

**Solution**:

```bash
# Rebuild images
TAG=local bash ./scripts/build.sh

# Load into Kind
for svc in api-gateway auth-service product-service order-service payment-service cart-service fulfillment-service notification-service search-service chat-service graphql-gateway; do
  kind load docker-image localhost:5000/${svc}:local --name go-micro-commerce
done

# Force pod restart
kubectl delete pods -l environment=local
```

### CrashLoopBackOff - Missing .env File

**Symptom**: Pods crash with `panic: open .env: no such file or directory`

**Cause**: Services are configured to load from environment variables (fixed in latest version)

**Solution**: This should not occur in the latest version. If it does, ensure you have the latest code where `internal/config/config.go` has:

```go
//nolint:errcheck // .env file not required when using environment variables
_ = viper.ReadInConfig()
```

### CrashLoopBackOff - Missing Infrastructure

**Symptom**: Pods crash with connection refused errors to PostgreSQL/Redis/Kafka

**Cause**: Infrastructure services not accessible

**Solutions**:

1. **Use host infrastructure**: Update ConfigMaps to point to `host.docker.internal` (see [Local Development](#-local-development-with-host-infrastructure))
2. **Deploy to K8s**: Install PostgreSQL, Redis, Kafka using Helm charts
3. **Check connectivity**: `kubectl exec -it <pod> -- nc -zv host.docker.internal 15431`

### CrashLoopBackOff - Missing JWT Keys

**Symptom**: API Gateway crashes with `failed to load public key: open keys/public.pem: no such file or directory`

**Cause**: JWT keys not mounted as Kubernetes secrets

**Solution**:

```bash
# Create secrets from existing keys
kubectl create secret generic api-gateway-jwt-keys \
  --from-file=public.pem=../../api-gateway/keys/public.pem

# Update deployment to mount secret (see base/api-gateway/deployment.yaml)
```

### Container Runtime Errors - exec /main not found

**Symptom**: `Error: failed to create containerd task: exec: "/main": stat /main: no such file or directory`

**Cause**: Dockerfile ENTRYPOINT issue (fixed in latest version)

**Solution**: Ensure Dockerfiles use `ENTRYPOINT ["./main"]` not `ENTRYPOINT ["/main"]`

### Pod Not Starting

```bash
kubectl describe pod <pod-name> -n production
kubectl logs <pod-name> -n production
kubectl logs <pod-name> -n production --previous  # For crashed pods
```

### Service Discovery Issues

```bash
kubectl run -it --rm debug --image=busybox --restart=Never -- sh
nslookup local-api-gateway.default.svc.cluster.local
```

### Linkerd Issues

```bash
linkerd check
linkerd -n production tap deployment/api-gateway
```

---

## 📚 Resources

- [Full Deployment Guide](../../docs/kubernetes-deployment.md)
- [Traefik Setup](infrastructure/traefik/README.md)
- [Main README](../../README.md)

---

## ✅ Local Development Checklist

- [ ] Kind cluster created
- [ ] Docker Compose infrastructure running (PostgreSQL, Redis, Kafka)
- [ ] Images built with `TAG=local bash ./scripts/build.sh`
- [ ] Images loaded into Kind cluster
- [ ] JWT key secrets created
- [ ] ConfigMaps updated to point to host infrastructure
- [ ] Services deployed with `kubectl apply -k overlays/local`
- [ ] Pods running without CrashLoopBackOff
- [ ] Services can connect to databases
- [ ] Health checks passing

## ✅ Production Deployment Checklist

- [ ] Kubernetes cluster ready
- [ ] Images built and pushed to registry
- [ ] Infrastructure deployed (PostgreSQL, Redis, Kafka, etc.)
- [ ] Secrets configured (JWT keys, DB passwords, etc.)
- [ ] ConfigMaps updated for production endpoints
- [ ] Traefik installed
- [ ] Linkerd installed and injected
- [ ] DNS configured
- [ ] TLS certificates configured
- [ ] Monitoring stack deployed
- [ ] Services deployed
- [ ] Health checks passing
- [ ] Load testing completed
- [ ] Argo CD configured (optional)
