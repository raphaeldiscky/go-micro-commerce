# Kubernetes Deployment Guide

Complete guide for deploying go-micro-commerce on Kubernetes with production-grade practices.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Prerequisites](#prerequisites)
3. [Quick Start - Local Deployment](#quick-start---local-deployment)
4. [Production Deployment](#production-deployment)
5. [Infrastructure Components](#infrastructure-components)
6. [Monitoring & Observability](#monitoring--observability)
7. [Multi-Cluster Setup](#multi-cluster-setup)
8. [Troubleshooting](#troubleshooting)

---

## Architecture Overview

### Deployment Layers

```
┌─────────────────────────────────────────────────────┐
│         Internet / External Traffic                  │
└──────────────────┬──────────────────────────────────┘
                   │
         ┌─────────▼──────────┐
         │  Traefik Ingress   │  (TLS, Rate Limiting, External Routing)
         └─────────┬──────────┘
                   │
         ┌─────────▼──────────┐
         │   API Gateway      │  (App-level routing, Circuit Breaking, Auth)
         └─────────┬──────────┘
                   │
    ┌──────────────┴──────────────┐
    │     Linkerd Service Mesh    │  (mTLS, Observability, Traffic Management)
    └──────────────┬──────────────┘
                   │
    ┌──────────────┴──────────────────────┐
    │      Kubernetes Services            │
    │  (DNS-based Service Discovery)      │
    └──────────────┬──────────────────────┘
                   │
    ┌──────────────┴──────────────────────┐
    │         Microservices (11 total)    │
    └─────────────────────────────────────┘
```

### Service Discovery

- Services discover each other via K8s DNS: `<service-name>.<namespace>.svc.cluster.local`
- Linkerd provides mTLS, load balancing, and cross-cluster routing
- **No Consul needed** - Kubernetes native service discovery

---

## Prerequisites

- **Kubernetes 1.25+** (EKS, GKE, AKS, or local Kind/Minikube)
- **kubectl** CLI installed
- **Kustomize** (built into kubectl)
- **Helm 3.x** for infrastructure components
- **Docker** for building images

---

## Quick Start - Local Deployment

### 1. Create Kind Cluster

```bash
kind create cluster --name go-micro-commerce
```

### 2. Build and Load Images

```bash
# Build all services
task build

# Load images into Kind
for service in api-gateway auth-service product-service order-service payment-service cart-service fulfillment-service notification-service search-service chat-service graphql-gateway; do
  kind load docker-image ${service}:latest --name go-micro-commerce
done
```

### 3. Deploy Services

```bash
kubectl apply -k deployments/k8s/overlays/local
```

### 4. Verify Deployment

```bash
kubectl get pods
kubectl get svc
```

### 5. Access API Gateway

```bash
kubectl port-forward svc/local-api-gateway 8080:8080
curl http://localhost:8080/health
```

---

## Production Deployment

### Step 1: Configure Container Registry

Update image references in overlay files:

```yaml
# deployments/k8s/overlays/prod/us-east/kustomization.yaml
images:
  - name: api-gateway
    newName: your-registry.io/api-gateway
    newTag: v1.0.0
```

### Step 2: Create Namespace

```bash
kubectl create namespace production
```

### Step 3: Configure Secrets

**Option A: Plain Secrets** (dev/staging only):
```bash
kubectl create secret generic api-gateway-secrets \
  --from-literal=JWT_SECRET=your-secret \
  -n production
```

**Option B: Sealed Secrets** (recommended for production):
```bash
# Install sealed-secrets controller
helm install sealed-secrets sealed-secrets/sealed-secrets -n kube-system

# Create sealed secret
kubeseal -f secret.yaml -w sealed-secret.yaml
kubectl apply -f sealed-secret.yaml
```

**Option C: External Secrets Operator** (best for production):
```bash
# Install external-secrets
helm install external-secrets external-secrets/external-secrets -n external-secrets --create-namespace

# Configure secret store (AWS Secrets Manager, Vault, etc.)
```

### Step 4: Deploy Infrastructure

#### Install Traefik

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update

helm install traefik traefik/traefik \
  -f deployments/k8s/infrastructure/traefik/values.yaml \
  -n traefik \
  --create-namespace
```

#### Install cert-manager (for TLS)

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.crds.yaml

helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.13.0
```

#### Install Linkerd

```bash
# Install Linkerd CLI
curl --proto '=https' --tlsv1.2 -sSfL https://run.linkerd.io/install | sh
export PATH=$PATH:$HOME/.linkerd2/bin

# Install Linkerd
linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -

# Verify
linkerd check
```

#### Enable Linkerd on Production Namespace

```bash
kubectl annotate namespace production linkerd.io/inject=enabled
```

### Step 5: Deploy Services

```bash
# Deploy to production (us-east)
kubectl apply -k deployments/k8s/overlays/prod/us-east

# Verify deployment
kubectl get pods -n production
kubectl get svc -n production

# Check Linkerd injection
linkerd -n production stat deployment
```

### Step 6: Configure Ingress

```bash
# Apply ingress resources
kubectl apply -f deployments/k8s/infrastructure/traefik/ingress.yaml

# Get LoadBalancer IP
kubectl get svc -n traefik traefik

# Configure DNS
# api.yourdomain.com -> <EXTERNAL-IP>
# graphql.yourdomain.com -> <EXTERNAL-IP>
```

### Step 7: Configure TLS

```bash
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: traefik
EOF
```

---

## Infrastructure Components

### Traefik Ingress Controller

See [deployments/k8s/infrastructure/traefik/README.md](../deployments/k8s/infrastructure/traefik/README.md) for detailed setup.

**Key Features**:
- TLS termination with Let's Encrypt
- Rate limiting
- Security headers
- Prometheus metrics

### Linkerd Service Mesh

**Why Linkerd over Consul**:
- ✅ Lightweight and fast
- ✅ Automatic mTLS
- ✅ Built-in observability
- ✅ Multi-cluster support
- ✅ No external dependencies

**Verify Mesh**:
```bash
linkerd -n production stat deployment
linkerd viz dashboard
```

### GitOps with Argo CD

```bash
# Install Argo CD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Access UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

---

## Monitoring & Observability

### OpenTelemetry

Services are instrumented with OpenTelemetry. Deploy OTEL Collector:

```bash
kubectl apply -f deployments/k8s/infrastructure/monitoring/otel-collector.yaml
```

### Access Dashboards

```bash
# Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000

# Prometheus
kubectl port-forward -n monitoring svc/prometheus 9090:9090

# Linkerd Viz
linkerd viz dashboard
```

---

## Multi-Cluster Setup

### Architecture

```
Production Clusters (HA across regions):
- us-east (primary)
- us-west (secondary)
- eu-west (europe)
```

### Linkerd Multi-Cluster

```bash
# Install on all clusters
linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -

# Enable multi-cluster
linkerd multicluster install | kubectl apply -f -

# Link clusters
linkerd --context=us-east multicluster link --cluster-name us-west | \
  kubectl --context=us-west apply -f -

# Export services
kubectl --context=us-east label svc/api-gateway \
  mirror.linkerd.io/exported=true \
  -n production
```

---

## Scaling

### Manual Scaling

```bash
kubectl scale deployment api-gateway --replicas=5 -n production
```

### Auto-Scaling (HPA)

HPA is pre-configured in base manifests:

```bash
kubectl get hpa -n production
```

### Vertical Scaling (VPA)

```bash
# Install VPA
kubectl apply -f https://github.com/kubernetes/autoscaler/releases/download/vertical-pod-autoscaler-0.13.0/vpa-v0.13.0.yaml
```

---

## Rolling Updates & Rollbacks

### Update Image

```bash
kubectl set image deployment/api-gateway \
  api-gateway=your-registry.io/api-gateway:v1.1.0 \
  -n production

# Monitor rollout
kubectl rollout status deployment/api-gateway -n production
```

### Rollback

```bash
kubectl rollout undo deployment/api-gateway -n production
```

### Canary Deployment with Linkerd

```yaml
apiVersion: split.smi-spec.io/v1alpha1
kind: TrafficSplit
metadata:
  name: api-gateway-split
  namespace: production
spec:
  service: api-gateway
  backends:
  - service: api-gateway-v1
    weight: 90
  - service: api-gateway-v2
    weight: 10
```

---

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n production
kubectl describe pod <pod-name> -n production
kubectl logs -n production <pod-name> -f
```

### Test Service Discovery

```bash
kubectl run -it --rm debug --image=busybox --restart=Never -- sh
nslookup api-gateway.production.svc.cluster.local
```

### Check Linkerd

```bash
linkerd check
linkerd -n production stat deployment
linkerd -n production tap deployment/api-gateway
```

### Check Metrics

```bash
kubectl port-forward -n production svc/api-gateway 8080:8080
curl http://localhost:8080/metrics
```

### Common Issues

**Pod CrashLoopBackOff**:
```bash
kubectl logs <pod-name> --previous
kubectl describe pod <pod-name>
```

**ImagePullBackOff**:
- Check image name and tag
- Verify registry credentials

**Service Unreachable**:
- Check service and endpoint: `kubectl get svc,ep -n production`
- Verify Linkerd injection: `linkerd -n production stat pod`

---

## Environment Configuration

| Environment | Replicas | CPU Request | Memory Request | HPA Min/Max |
|-------------|----------|-------------|----------------|-------------|
| Local       | 1        | 50m         | 64Mi           | 1/1         |
| Dev         | 1        | 100m        | 128Mi          | 1/3         |
| Staging     | 2        | 200m        | 256Mi          | 2/5         |
| Production  | 3        | 500m        | 512Mi          | 3/20        |

---

## Security Best Practices

1. ✅ Use **Sealed Secrets** or **External Secrets Operator**
2. ✅ Enable **Pod Security Standards**
3. ✅ Implement **Network Policies**
4. ✅ Configure **RBAC** properly
5. ✅ Scan images in CI/CD
6. ✅ Use Linkerd **mTLS** everywhere
7. ✅ Regular security audits

---

## Next Steps

- [ ] Set up CI/CD pipeline
- [ ] Configure managed databases
- [ ] Set up monitoring alerts
- [ ] Implement backup strategy
- [ ] Performance testing
- [ ] Security hardening

---

## References

- [Kubernetes Docs](https://kubernetes.io/docs/)
- [Linkerd Docs](https://linkerd.io/docs/)
- [Traefik Kubernetes](https://doc.traefik.io/traefik/providers/kubernetes-ingress/)
- [Argo CD Docs](https://argo-cd.readthedocs.io/)
