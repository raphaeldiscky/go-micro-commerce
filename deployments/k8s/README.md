# Kubernetes Deployments - Industry-Standard GitOps Structure

This directory contains Kubernetes manifests for deploying the go-micro-commerce platform using **ArgoCD GitOps** with an industry-standard hybrid pattern.

## Architecture Overview

```
deployments/k8s/
├── apps/                          # ArgoCD Application definitions
│   └── applicationsets/           # ApplicationSet generators
│       ├── infrastructure.yaml    # Auto-discovers infrastructure/*
│       └── workloads.yaml         # Auto-discovers workloads/overlays/prod/*
├── infrastructure/                # Platform services (flat structure)
│   ├── ingress/
│   │   └── api-backend/          # API ingress with TLS & middlewares
│   ├── monitoring/               # Prometheus, Grafana, Loki
│   ├── kafka/                    # Kafka infrastructure
│   ├── redis/                    # Redis caching
│   ├── postgres/                 # PostgreSQL databases
│   └── traefik/                  # Ingress controller
└── workloads/                    # Microservices (base + overlays)
    ├── base/                     # Base configurations (DRY)
    │   ├── api-gateway/
    │   ├── auth-service/
    │   ├── cart-service/
    │   ├── chat-service/
    │   ├── fulfillment-service/
    │   ├── notification-service/
    │   ├── order-service/
    │   ├── payment-service/
    │   ├── product-service/
    │   ├── search-service/
    │   └── external-secrets/
    └── overlays/                 # Environment-specific overrides
        ├── dev/                  # Development environment
        └── prod/                 # Production environment
            ├── api-gateway/
            │   ├── kustomization.yaml
            │   ├── patch-replicas.yaml    # 3 replicas
            │   └── patch-resources.yaml   # Higher limits
            ├── product-service/
            └── ... (all 10 services)
```

## GitOps Workflow

### How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                    1. Developer Push                         │
│   git push → deployments/k8s/workloads/overlays/prod/       │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────────┐
│           2. ArgoCD Detects Changes (Auto-Sync)              │
│   Bootstrap ApplicationSet → Deploys infrastructure.yaml +   │
│                               workloads.yaml                 │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────────┐
│           3. ApplicationSets Generate Applications           │
│   - infrastructure.yaml: Discovers infrastructure/*/*        │
│   - workloads.yaml: Discovers workloads/overlays/prod/*     │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────────┐
│              4. ArgoCD Syncs to Cluster                      │
│   Each discovered directory becomes an Application          │
│   Kustomize builds manifests → kubectl apply                │
└─────────────────────────────────────────────────────────────┘
```

### Bootstrap Flow (Terraform → ArgoCD)

1. **Terraform provisions** GKE cluster and installs ArgoCD
2. **Terraform deploys bootstrap ApplicationSet** pointing to `apps/applicationsets/*.yaml`
3. **Bootstrap ApplicationSet discovers**:
   - `infrastructure.yaml` → Deploys all infrastructure components
   - `workloads.yaml` → Deploys all production microservices
4. **Applications auto-sync** on every git push

## Directory Structure Explained

### `/apps/` - ArgoCD Definitions

Contains **ApplicationSet** manifests that define auto-discovery patterns:

- **Purpose**: Meta-layer that generates ArgoCD Applications
- **Pattern**: Industry-standard GitOps control plane
- **Managed by**: Terraform bootstrap ApplicationSet

### `/infrastructure/` - Platform Services

Contains infrastructure and platform components using **flat structure**:

- **Purpose**: Services that don't vary significantly between environments
- **Examples**: Ingress controllers, monitoring, databases, message queues
- **Pattern**: Direct manifests or Kustomize (no overlays needed)
- **Deployed by**: `infrastructure.yaml` ApplicationSet

**Why flat structure?**

- Infrastructure configs are mostly environment-agnostic
- Changes are infrequent and controlled
- Simpler to manage without overlay complexity

### `/workloads/` - Microservices

Contains application workloads using **base + overlays pattern**:

#### `/workloads/base/` - Base Configurations

- **Purpose**: Shared configuration for all environments (DRY principle)
- **Contains**: deployment.yaml, service.yaml, configmap.yaml, hpa.yaml, etc.
- **Principle**: Define once, override selectively

#### `/workloads/overlays/` - Environment Overrides

- **Purpose**: Environment-specific patches and configurations
- **Pattern**: Kustomize overlays that reference base/
- **Environments**:
  - `dev/`: Development (local Kubernetes, lower resources)
  - `prod/`: Production (GKE, higher replicas, more resources)

**Why overlays?**

- Microservices vary significantly between environments (replicas, resources, secrets)
- Changes are frequent (daily deployments)
- Maintains DRY principle while allowing environment-specific configs

## Deployment Guide

### Prerequisites

1. **Terraform applied** with ArgoCD module configured:

   ```hcl
   # terraform/environments/prod/terraform.tfvars
   argocd_git_repo_url     = "https://github.com/YOUR_ORG/go-micro-commerce.git"
   argocd_enable_bootstrap = true
   ```

2. **Git repository** accessible by ArgoCD (public or with credentials)

3. **ArgoCD installed** at `https://argocd.api.discky.com`

### Initial Setup

```bash
# 1. Ensure ApplicationSet manifests are ready
ls deployments/k8s/apps/applicationsets/
# Output: infrastructure.yaml  workloads.yaml

# 2. Update git repo URL in ApplicationSets
# Edit these files and replace YOUR_ORG with your GitHub organization
vim deployments/k8s/apps/applicationsets/infrastructure.yaml
vim deployments/k8s/apps/applicationsets/workloads.yaml

# 3. Commit and push
git add deployments/k8s/
git commit -m "feat(k8s): implement industry-standard GitOps structure"
git push origin main

# 4. Apply Terraform to deploy bootstrap
cd terraform/environments/prod
terraform apply

# 5. Verify ArgoCD deployed ApplicationSets
kubectl get applicationsets -n argocd
# Output:
# NAME                        AGE
# bootstrap-applicationsets   1m
# infrastructure              30s
# workloads                   30s

# 6. Watch applications being created
kubectl get applications -n argocd -w
```

### Adding New Services

#### For Infrastructure Components

```bash
# 1. Create new directory under infrastructure/
mkdir -p deployments/k8s/infrastructure/cert-manager/

# 2. Add Kubernetes manifests
cat > deployments/k8s/infrastructure/cert-manager/deployment.yaml << EOF
apiVersion: apps/v1
kind: Deployment
...
EOF

# 3. Commit and push
git add deployments/k8s/infrastructure/cert-manager/
git commit -m "feat(infra): add cert-manager"
git push

# ArgoCD auto-discovers and deploys within 3 minutes
```

#### For Microservices

```bash
# 1. Create base configuration
mkdir -p deployments/k8s/workloads/base/new-service/
# Add deployment.yaml, service.yaml, etc.

# 2. Create production overlay
mkdir -p deployments/k8s/workloads/overlays/prod/new-service/
cat > deployments/k8s/workloads/overlays/prod/new-service/kustomization.yaml << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: default

resources:
  - ../../../base/new-service

commonLabels:
  environment: production

patchesStrategicMerge:
  - patch-replicas.yaml
EOF

# 3. Add patches
# patch-replicas.yaml, patch-resources.yaml

# 4. Commit and push
git add deployments/k8s/workloads/
git commit -m "feat(services): add new-service"
git push

# ArgoCD auto-discovers and deploys
```

### Monitoring Deployments

```bash
# Access ArgoCD UI
open https://argocd.api.discky.com

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath='{.data.password}' | base64 -d

# CLI: List all applications
argocd app list

# CLI: Sync specific application
argocd app sync api-gateway

# CLI: Get application details
argocd app get product-service
```

## Environment Management

### Development (dev/)

```yaml
# Lower resources, fewer replicas
spec:
  replicas: 1
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
```

### Production (prod/)

```yaml
# Higher resources, HA setup
spec:
  replicas: 3 # High availability
  resources:
    requests:
      cpu: "500m"
      memory: "512Mi"
    limits:
      cpu: "1000m"
      memory: "1Gi"
```

## Best Practices

### 1. Base Configurations

- ✅ Keep base/ generic and reusable
- ✅ No environment-specific values in base/
- ✅ Use ConfigMap/Secret generators in overlays
- ❌ Don't hardcode image tags in base/

### 2. Overlays

- ✅ Use strategic merge patches for small changes
- ✅ Use JSON patches for precise modifications
- ✅ Override image tags per environment
- ✅ Add environment labels
- ❌ Don't duplicate entire manifests

### 3. Infrastructure

- ✅ Keep infrastructure configs simple (flat structure OK)
- ✅ Use Helm for complex charts (via ArgoCD)
- ✅ Separate infrastructure from workloads
- ❌ Don't mix application and platform concerns

### 4. GitOps Workflow

- ✅ Always commit before deploying
- ✅ Use pull requests for production changes
- ✅ Enable auto-sync for non-critical apps
- ✅ Use manual sync for critical apps (databases)
- ❌ Never run `kubectl apply` manually

## Troubleshooting

### ApplicationSet not creating Applications

```bash
# Check ApplicationSet status
kubectl describe applicationset infrastructure -n argocd
kubectl describe applicationset workloads -n argocd

# Check if directories exist
ls deployments/k8s/infrastructure/
ls deployments/k8s/workloads/overlays/prod/

# Force refresh
argocd appset get infrastructure
```

### Application Out of Sync

```bash
# Check diff
argocd app diff api-gateway

# Sync manually
argocd app sync api-gateway

# Hard refresh (ignore cache)
argocd app sync api-gateway --force
```

### Kustomize Build Errors

```bash
# Test kustomize build locally
cd deployments/k8s/workloads/overlays/prod/api-gateway/
kustomize build .

# Common issues:
# - Wrong path in resources[]
# - Missing patch files
# - Invalid YAML syntax
```

## Migration from Old Structure

If migrating from `base/` + `overlays/local/`:

1. ✅ Already done: Moved to `workloads/base/` + `workloads/overlays/dev/`
2. ✅ Production overlays created: `workloads/overlays/prod/`
3. ✅ ApplicationSets created: `apps/applicationsets/`
4. ⏩ Next: Update git repo URL and enable bootstrap in Terraform

## References

- [ArgoCD ApplicationSets](https://argo-cd.readthedocs.io/en/stable/user-guide/application-set/)
- [Kustomize Documentation](https://kustomize.io/)
- [CNCF GitOps Principles](https://opengitops.dev/)
- [Kubernetes SIG Multi-Tenancy Best Practices](https://github.com/kubernetes-sigs/multi-tenancy)

## Support

- **ArgoCD UI**: https://argocd.api.discky.com
- **Grafana Dashboards**: https://grafana.api.discky.com
- **Kubernetes Docs**: https://kubernetes.io/docs/

---

**Pattern**: Industry-standard hybrid GitOps (infrastructure flat + workloads base+overlays)
**Inspiration**: Google GKE, AWS EKS GitOps Bridge, Red Hat OpenShift GitOps
**Compliance**: CNCF GitOps v1.0, Kubernetes SIG recommendations
