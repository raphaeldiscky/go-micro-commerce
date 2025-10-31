# Kubernetes Deployment

Production-grade Kubernetes deployment for go-micro-commerce microservices platform.

## 🚀 Quick Start

### Local Development (Kind)

```bash
# 1. Create Kind cluster
kind create cluster --name go-micro-commerce

# 2. Build and load images
task build
for svc in api-gateway auth-service product-service order-service payment-service cart-service fulfillment-service notification-service search-service chat-service graphql-gateway; do
  kind load docker-image ${svc}:latest --name go-micro-commerce
done

# 3. Deploy
kubectl apply -k overlays/local

# 4. Verify
kubectl get pods
kubectl port-forward svc/local-api-gateway 8080:8080
```

### Production

See [docs/kubernetes-deployment.md](../../docs/kubernetes-deployment.md) for complete production deployment guide.

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

### Pod Not Starting
```bash
kubectl describe pod <pod-name> -n production
kubectl logs <pod-name> -n production
```

### Service Discovery Issues
```bash
kubectl run -it --rm debug --image=busybox --restart=Never -- sh
nslookup api-gateway.production.svc.cluster.local
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

## ✅ Deployment Checklist

- [ ] Kubernetes cluster ready
- [ ] Images built and pushed to registry
- [ ] Secrets configured
- [ ] Traefik installed
- [ ] Linkerd installed and injected
- [ ] DNS configured
- [ ] TLS certificates configured
- [ ] Monitoring stack deployed
- [ ] Services deployed
- [ ] Health checks passing
- [ ] Load testing completed
- [ ] Argo CD configured (optional)
