# Production-Grade Kubernetes Deployment Plan

Architecture Overview

Multi-cluster strategy:

- Dev cluster (local Kind/Minikube)
- Staging cluster (cloud, single region)
- Production clusters (multi-region for HA: us-east, us-west, eu-west)

Service mesh: Linkerd (replacing Consul for service discovery + mesh features)

Ingress: Traefik Ingress Controller (external traffic) + API Gateway (application-level)

GitOps: Argo CD for declarative deployments

Implementation Plan

Phase 1: Remove Consul & Prepare for Kubernetes

1.  Remove Consul dependencies from services

- Update service clients to use Kubernetes DNS (e.g., product-service.production.svc.cluster.local:8082)
- Remove Consul registration code from all services (pkg/consul/, service main.go)
- Update API Gateway to use Kubernetes Service discovery instead of Consul
- Create environment-based configuration (dev/staging/prod service endpoints)

2.  Update configuration management

- Add Kubernetes-specific configs (service names, namespaces)
- Environment-specific ConfigMaps and Secrets structure
- Health check endpoints remain unchanged (already K8s-compatible)

Phase 2: Kubernetes Manifests & Helm Charts

1.  Base Kubernetes resources (deployments/k8s/base/)

- Deployments for each microservice (10 services total)
- Services (ClusterIP for internal, LoadBalancer for gateways)
- ConfigMaps and Secrets
- ServiceAccounts and RBAC
- HorizontalPodAutoscalers (HPA) for auto-scaling
- PodDisruptionBudgets (PDB) for HA

2.  Environment-specific overlays (deployments/k8s/overlays/{dev,staging,prod}/)

- Kustomize overlays for each environment
- Resource limits/requests per environment
- Replica counts (dev: 1, staging: 2, prod: 3+)
- Environment variables and secrets

3.  Stateful services setup

- In-cluster with operators: Redis (Redis Operator)
- Managed services: PostgreSQL (RDS/Cloud SQL), Kafka (MSK/Confluent Cloud), Elasticsearch (Elastic Cloud/AWS OpenSearch)
- NOT DEPLOYED: Temporal (kept external or not deployed to K8s)
- StatefulSets and PersistentVolumeClaims where needed

Phase 3: Traefik Ingress Controller

1.  Install Traefik via Helm

- Configure as Kubernetes Ingress Controller
- TLS/SSL termination with cert-manager
- Rate limiting and middleware

2.  Ingress resources

- External ingress: api.yourdomain.com → API Gateway
- External ingress: graphql.yourdomain.com → GraphQL Gateway
- Internal services use ClusterIP (no external exposure)

3.  Keep custom API Gateway

- Runs as Kubernetes Deployment
- Uses K8s DNS for backend service discovery (http://product-service:8082)
- Handles application-level concerns (auth, rate limiting, circuit breaker)

Phase 4: Linkerd Service Mesh

1.  Install Linkerd

- Automatic mTLS for service-to-service communication
- Traffic management and retries
- Observability (metrics, traces)
- Multi-cluster traffic routing (for HA setup)

2.  Mesh configuration

- Inject Linkerd proxy into all service pods
- Configure retry policies and timeouts
- Traffic splitting for canary deployments
- Cross-cluster service mirroring for HA

Phase 5: Argo CD GitOps

1.  Install Argo CD per cluster

- Central Argo CD in prod-primary cluster manages all clusters
- Cluster credentials stored securely

2.  Git repository structure
    deployments/
    ├── k8s/
    │ ├── base/ # Base manifests
    │ ├── overlays/
    │ │ ├── dev/
    │ │ ├── staging/
    │ │ └── prod/
    │ │ ├── us-east/
    │ │ ├── us-west/
    │ │ └── eu-west/
    │ └── argocd/
    │ ├── applications/ # Argo Application CRDs
    │ └── projects/ # Argo Project definitions
3.  Argo CD Applications

- One Application per service per environment
- Automated sync with health checks
- Rollback capabilities

Phase 6: Observability & Monitoring

1.  OpenTelemetry (already implemented, ensure K8s compatibility)

- Deploy OTEL Collector as DaemonSet
- Configure service endpoints

2.  LGTM Stack in-cluster

- Prometheus (metrics)
- Loki (logs)
- Tempo (traces)
- Grafana (dashboards)

Phase 7: Multi-Cluster HA Setup

1.  Cross-cluster service discovery via Linkerd

- Service mirroring between prod clusters
- Automatic failover

2.  Global load balancing

- Cloud provider GLB or external DNS
- Health-based routing

File Structure

deployments/
├── k8s/
│ ├── base/
│ │ ├── api-gateway/
│ │ ├── auth-service/
│ │ ├── product-service/
│ │ ├── order-service/
│ │ ├── payment-service/
│ │ ├── cart-service/
│ │ ├── fulfillment-service/
│ │ ├── notification-service/
│ │ ├── search-service/
│ │ ├── chat-service/
│ │ ├── graphql-gateway/
│ │ └── infrastructure/
│ │ ├── redis/
│ │ └── monitoring/
│ ├── overlays/
│ │ ├── local/ # Kind/Minikube
│ │ ├── dev/
│ │ ├── staging/
│ │ └── prod/
│ │ ├── us-east/
│ │ ├── us-west/
│ │ └── eu-west/
│ ├── argocd/
│ ├── linkerd/
│ └── traefik/
└── helm/ # Optional: Helm charts for services

Migration Strategy

1.  Start with dev/local: Deploy to Kind cluster first
2.  Test staging: Single staging cluster deployment
3.  Production rollout: Blue-green deployment per region
4.  Multi-cluster setup: Enable Linkerd multi-cluster
5.  GitOps: Move to Argo CD management

Services to Deploy (11 total)

1.  api-gateway
2.  auth-service
3.  product-service
4.  order-service (using custom saga, not Temporal)
5.  payment-service
6.  cart-service
7.  fulfillment-service
8.  notification-service
9.  search-service
10. chat-service
11. graphql-gateway

Key Benefits

✅ No Consul: Simplified stack, K8s-native service discovery
✅ Linkerd: mTLS, observability, multi-cluster without complexity
✅ Dual gateway layers: Traefik (ingress) + API Gateway (app-level)
✅ GitOps: Declarative, auditable, rollback-friendly
✅ Hybrid stateful: Balance between control and operational simplicity
✅ Multi-cluster HA: Production resilience across regions
✅ No Temporal deployment: Simpler stack, order-service uses custom saga

This approach is production-grade, follows cloud-native best practices, and aligns with your current architecture patterns.
