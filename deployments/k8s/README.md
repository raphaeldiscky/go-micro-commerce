# Kubernetes Deployments - GitOps

This directory contains Kubernetes manifests for deploying the go-micro-commerce platform using **ArgoCD GitOps**.

## Architecture Overview

```
deployments/k8s/
├── apps/                          # ArgoCD Application definitions
│   ├── applicationsets/           # ApplicationSet generators
│   │   ├── infrastructure.yaml    # Auto-discovers infrastructure/overlays/prod/*
│   │   └── workloads.yaml         # Auto-discovers workloads/overlays/prod/*
│   └── projects/                  # ArgoCD Project definitions (RBAC)
│       ├── applications-project.yaml
│       └── infrastructure-project.yaml
├── infrastructure/                # Platform services (base + overlays)
│   ├── base/
│   │   ├── api-gateway/          # API Gateway service
│   │   ├── apollo-router/        # GraphQL federation gateway
│   │   ├── gateway/              # API Gateway HTTPRoute and middlewares
│   │   ├── graphql/              # GraphQL HTTPRoute with ReferenceGrant
│   │   ├── ingress-controller/   # Traefik with Kubernetes Gateway API
│   │   │   ├── gatewayclass.yaml # GatewayClass (Traefik controller)
│   │   │   └── gateway.yaml      # Shared Gateway for HTTPS traffic
│   │   ├── kafka/                # Kafka CRDs
│   │   ├── monitoring/           # Prometheus, Grafana, Tempo, Loki, Alloy
│   │   ├── namespaces/           # Namespace definitions
│   │   ├── postgres/             # PostgreSQL Cluster CRDs (9 databases)
│   │   └── redis/                # Redis CRDs
│   └── overlays/
│       ├── local/                # Local dev (Tilt/Minikube)
│       │   ├── api-gateway/      # Local API Gateway config
│       │   ├── graphql/          # Local GraphQL config
│       │   ├── ingress-routes/   # Local ingress (*.localhost)
│       │   ├── kafka/            # Single broker
│       │   ├── mailer/           # MailHog for local testing
│       │   ├── monitoring/       # Dev monitoring stack
│       │   ├── postgres/         # 1 replica, low resources
│       │   └── redis/            # Minimal config
│       └── prod/                 # Production (ArgoCD/GKE)
│           ├── api-gateway/      # Production API Gateway
│           ├── common/           # Shared Kustomize component
│           ├── graphql/          # Production GraphQL config
│           ├── ingress-controller/
│           ├── kafka/            # Multi-broker cluster
│           ├── monitoring/       # Production monitoring
│           ├── namespaces/       # Production namespaces
│           ├── postgres/         # 3 replicas, HA config
│           └── redis/            # Production Redis
└── workloads/                    # Microservices (base + overlays)
    ├── base/                     # Base configurations (environment-agnostic)
    │   ├── auth-service/
    │   ├── cart-service/
    │   ├── chat-service/
    │   ├── external-secrets/     # External Secrets Operator configs
    │   ├── fulfillment-service/
    │   ├── notification-service/
    │   ├── order-service/
    │   ├── payment-service/
    │   └── product-service/
    └── overlays/                   # Environment-specific overrides
        ├── local/                  # Local development (Tilt)
        │   ├── kustomization.yaml  # Single file, monolithic
        │   └── secrets/            # TLS and JWT keys
        └── prod/                   # Production (ArgoCD/GKE)
            ├── auth-service/
            │   ├── kustomization.yaml
            │   ├── patch-replicas.yaml        # 3 replicas
            │   ├── patch-resources.yaml       # Higher limits
            │   ├── patch-hpa.yaml             # Autoscaling
            │   └── patch-image-pull-secrets.yaml
            ├── common/           # Shared Kustomize component
            └── ... (all 9 services, modular)
```

## GitOps Workflow

### How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                    1. Developer Push                              │
│   git push -> deployments/k8s/workloads/overlays/prod/            │
└────────────────────┬────────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│           2. ArgoCD Detects Changes (Auto-Sync)                   │
│   Bootstrap ApplicationSet -> Deploys infrastructure.yaml +       │
│                               workloads.yaml                      │
└────────────────────┬────────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│           3. ApplicationSets Generate Applications                │
│   - infrastructure.yaml: Discovers infrastructure/overlays/prod/* │
│   - workloads.yaml: Discovers workloads/overlays/prod/*           │
└────────────────────┬────────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│              4. ArgoCD Syncs to Cluster                           │
│   Each discovered directory becomes an Application                │
│   Kustomize builds manifests -> kubectl apply                     │
└─────────────────────────────────────────────────────────────┘
```

### Bootstrap Flow (Terraform -> ArgoCD)

1. **Terraform provisions** GKE cluster and installs ArgoCD
2. **Terraform deploys bootstrap ApplicationSet** pointing to `apps/applicationsets/*.yaml`
3. **Bootstrap ApplicationSet discovers**:
   - `infrastructure.yaml` -> Deploys all infrastructure components
   - `workloads.yaml` -> Deploys all production microservices
4. **Applications auto-sync** on every git push

## Directory Structure Explained

### `/apps/` - ArgoCD Definitions

Contains **ApplicationSet** manifests that define auto-discovery patterns:

- **Purpose**: Meta-layer that generates ArgoCD Applications
- **Pattern**: Industry-standard GitOps structure
- **Managed by**: Terraform bootstrap ApplicationSet

### `/infrastructure/` - Platform Services

Contains infrastructure and platform components using **base + overlays pattern**:

- **Purpose**: Databases, message queues, ingress, monitoring, operators
- **Examples**: PostgreSQL, Kafka, Redis, Traefik, Prometheus, Grafana
- **Pattern**: Base configurations + environment-specific overlays
- **Deployed by**:
  - Local: Tilt with `kustomize('infrastructure/overlays/local')`
  - Production: ArgoCD `infrastructure.yaml` ApplicationSet

**Why base + overlays?** Infrastructure varies between local (1 replica, low resources) and production (3+ replicas, HA). Operators are deployed via Helm, CRDs via Kustomize.

#### `/apps/projects/` - ArgoCD Project Definitions

Contains ArgoCD Project manifests that define RBAC boundaries for applications:

- **applications-project.yaml**: RBAC for workload applications (microservices)
- **infrastructure-project.yaml**: RBAC for infrastructure applications (databases, monitoring)

**Purpose**: Projects provide namespace isolation and permission boundaries for ArgoCD-managed resources.

#### Kubernetes Operators

Operators are **controller software** that manage custom resources. They are installed via:

- **Terraform**: Production deployment with Helm charts
- **Tilt**: Local development with Helm charts

The operators manage CRDs defined in `infrastructure/base/`:

- **CloudNativePG**: Manages PostgreSQL Cluster CRDs in `postgres/`
- **Strimzi Kafka**: Manages Kafka/KafkaNodePool CRDs in `kafka/`
- **Redis Operator**: Manages RedisCluster CRDs in `redis/`

#### `/infrastructure/base/ingress-controller/` - Kubernetes Gateway API

Traffic routing uses the **Kubernetes Gateway API** (CNCF standard) with Traefik as the gateway controller:

- **GatewayClass**: Defines Traefik as the controller (`traefik.io/gateway-controller`)
- **Gateway**: Shared entry point with HTTPS listeners for `api.discky.com`
- **HTTPRoute**: Path-based routing to services (API Gateway at `/`, GraphQL at `/graph`)
- **ReferenceGrant**: Explicit cross-namespace routing security

**Traffic Flow**: `Client → Gateway → HTTPRoute → Service`

### `/workloads/` - Microservices

Contains application workloads using **base + overlays pattern**:

#### `/workloads/base/` - Base Configurations

- **Purpose**: Shared, environment-agnostic configuration (DRY principle)
- **Contains**: deployment.yaml, service.yaml, configmap.yaml, hpa.yaml, secret.yaml, etc.
- **Important**: Base contains NO environment-specific values:
  - ❌ No hardcoded namespaces (set in overlays)
  - ❌ No `APP_ENVIRONMENT` values (set in overlays)
  - ❌ No production/development-specific configs
  - ✅ Generic defaults and sensible fallbacks only
- **Principle**: Define once, override selectively in overlays

#### `/workloads/overlays/` - Environment Overrides

- **Purpose**: Environment-specific patches, namespaces, and configurations
- **Pattern**: Kustomize overlays that reference `../../../base/`
- **Environments**:
  - `local/`: Local development (Tilt/Minikube, lower resources)
    - Single monolithic `kustomization.yaml` file
    - Namespaces: `application` and `gateway`
  - `prod/`: Production (ArgoCD/GKE, higher replicas, more resources)
    - Modular structure (one directory per service)
    - Namespaces: `application` and `gateway`
    - Patches: replicas, resources, HPA, image pull secrets

**Why overlays?** Microservices vary between environments (replicas, resources, namespaces). This pattern maintains DRY while allowing environment-specific configs.

## Local Development with Tilt

Tilt orchestrates local Kubernetes development using Kustomize overlays:

- Installs operators via Helm (CloudNativePG, Strimzi Kafka, Redis)
- Deploys infrastructure: `kustomize('infrastructure/overlays/local')`
- Deploys workloads: `kustomize('workloads/overlays/local')`

### Local Characteristics

- **Low resources**: Minimal CPU/memory requests for laptop development
- **MailHog**: Local SMTP server at `http://localhost:8025`
- **Local ingress**: Services accessible at `*.localhost`
- **Namespaces**: `application` and `gateway` (matching production)
- **Hot reload**: Air for Go services, live updates via Tilt

### Running Locally

```bash
# Start infrastructure and services
tilt up

# View Tilt UI
open http://localhost:10350

# Check resources
kubectl get pods -n application
kubectl get pods -n gateway

# Access services
# API Gateway: http://localhost:8080
# MailHog UI: http://localhost:8025
# Grafana: http://localhost:3000
```

### Local vs Production

| Aspect         | Local               | Production         |
| -------------- | ------------------- | ------------------ |
| **Deployment** | Tilt + Kustomize    | ArgoCD + Kustomize |
| **Replicas**   | 1 (single instance) | 3+ (HA)            |
| **Resources**  | 256Mi RAM, 100m CPU | 1Gi RAM, 500m CPU  |
| **Databases**  | 1 pod               | 3 pods (HA)        |

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
# 1. Create base configuration
mkdir -p deployments/k8s/infrastructure/base/cert-manager/

# 2. Add Kubernetes manifests (environment-agnostic)
cat > deployments/k8s/infrastructure/base/cert-manager/deployment.yaml << EOF
apiVersion: apps/v1
kind: Deployment
...
EOF

# 3. Create local overlay
mkdir -p deployments/k8s/infrastructure/overlays/local/cert-manager/
cat > deployments/k8s/infrastructure/overlays/local/cert-manager/kustomization.yaml << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../../base/cert-manager
EOF

# 4. Create production overlay
mkdir -p deployments/k8s/infrastructure/overlays/prod/cert-manager/
# Add patches for prod-specific configs

# 5. Commit and push
git add deployments/k8s/infrastructure/
git commit -m "feat(infra): add cert-manager"
git push

# ArgoCD auto-discovers and deploys within 3 minutes
```

#### For Microservices

```bash
# 1. Create base configuration (environment-agnostic, no namespace)
mkdir -p deployments/k8s/workloads/base/new-service/
cat > deployments/k8s/workloads/base/new-service/deployment.yaml << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: new-service
spec:
  template:
    spec:
      containers:
      - name: new-service
        image: new-service:latest
EOF
# Add service.yaml, configmap.yaml, secret.yaml, hpa.yaml, etc.

# 2. Update local overlay (add to existing monolithic kustomization)
# Edit deployments/k8s/workloads/overlays/local/kustomization.yaml
# Add "- ../../base/new-service" to resources

# 3. Create production overlay (modular structure)
mkdir -p deployments/k8s/workloads/overlays/prod/new-service/
cat > deployments/k8s/workloads/overlays/prod/new-service/kustomization.yaml << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: application  # Set namespace in overlay, NOT base

resources:
  - ../../../base/new-service

labels:
  - pairs:
      environment: production
      app.kubernetes.io/managed-by: argocd

patches:
  - path: patch-replicas.yaml
  - path: patch-resources.yaml
  - path: patch-hpa.yaml
  - path: patch-image-pull-secrets.yaml
EOF

# 4. Add patches
cat > deployments/k8s/workloads/overlays/prod/new-service/patch-replicas.yaml << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: new-service
spec:
  replicas: 3
EOF
# Add patch-resources.yaml, patch-hpa.yaml, etc.

# 5. Commit and push
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

**Local** (`overlays/local/`): Tilt deployment, 1 replica, 256Mi RAM, MailHog, `*.localhost`

**Production** (`overlays/prod/`): ArgoCD deployment, 3+ replicas, 1Gi RAM, real SMTP, TLS, HPA enabled

## Best Practices

### Base Configurations

- NO hardcoded namespaces or environment-specific values (`APP_ENVIRONMENT`, etc.)
- Use placeholder secrets with `# NOTE: Override in overlays` comments
- Generic defaults only, image tags set in overlays

### Overlays

- Set namespaces in overlays, never in base
- Use patch files (`patch-*.yaml`) for environment-specific changes
- Consistent naming: `patch-image-pull-secrets.yaml` (kebab-case)
- Add environment labels

### Infrastructure & Operators

- Operators installed via Helm (Terraform for prod, Tilt for local)
- CRDs (Kustomize) in `infrastructure/base/{postgres,kafka,redis}/`
- Environment patches in `infrastructure/overlays/{local,prod}/`
- ArgoCD Projects in `apps/projects/` for RBAC boundaries

## Troubleshooting

### ApplicationSet not creating Applications

```bash
# Check ApplicationSet status
kubectl describe applicationset infrastructure -n argocd
kubectl describe applicationset workloads -n argocd

# Check if overlay directories exist (ArgoCD looks here)
ls deployments/k8s/infrastructure/overlays/prod/
ls deployments/k8s/workloads/overlays/prod/

# Verify base directories exist
ls deployments/k8s/infrastructure/base/
ls deployments/k8s/workloads/base/

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
cd deployments/k8s/workloads/overlays/prod/auth-service/
kustomize build .

# Common issues:
# - Wrong path in resources[]
# - Missing patch files
# - Invalid YAML syntax
```

## References

- [ArgoCD ApplicationSets](https://argo-cd.readthedocs.io/en/stable/user-guide/application-set/)
- [Kustomize Documentation](https://kustomize.io/)
- [CNCF GitOps Principles](https://opengitops.dev/)
- [Kubernetes SIG Multi-Tenancy Best Practices](https://github.com/kubernetes-sigs/multi-tenancy)

## Support

- **ArgoCD UI**: <https://argocd.api.discky.com>
- **Grafana Dashboards**: <https://grafana.api.discky.com>
- **Kubernetes Docs**: <https://kubernetes.io/docs/>

---

**Pattern**: GitOps with base + overlays for infrastructure and workloads
**Local**: Tilt + Kustomize (`overlays/local/`) | **Production**: ArgoCD + Kustomize (`overlays/prod/`)
**Compliance**: CNCF GitOps v1.0, ~9.5/10 best practices
