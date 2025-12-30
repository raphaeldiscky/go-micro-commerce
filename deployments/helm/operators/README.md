# Kubernetes Operators

Operators extend Kubernetes functionality by managing custom resources.

## Operators vs Custom Resources

### Operator (Infra Pool) - The Manager

The operator is **controller software** that watches and manages resources.

- **Installed via**: Helm (cluster-wide)
- **Runs in**: System namespaces (e.g., `cnpg-system`)
- **Purpose**: Watches for Custom Resources and manages them
- **Location**: `base/operators/*/operator-values.yaml`
- **Lifecycle**: Installed once, runs continuously

### Custom Resource (Data Plane) - What's Being Managed

Custom Resources are **workload declarations** that operators manage.

- **Deployed via**: Kustomize (per namespace)
- **Runs in**: Application namespaces (e.g., `application`)
- **Purpose**: Declares what you want created
- **Location**: `base/{postgres,kafka,redis}/*.yaml`
- **Lifecycle**: Created/deleted as needed

## Directory Structure

```
operators/
├── cloudnative-pg/          # Manages PostgreSQL Cluster CRDs
│   └── operator-values.yaml
├── strimzi-kafka/           # Manages Kafka CRDs
│   └── operator-values.yaml
└── redis-operator/          # Manages RedisCluster CRDs
    └── operator-values.yaml
```

## How It Works

```
┌─────────────────────────────────────┐
│  Operator                               │
│  Installed once via Helm                │
│  Location: operators/*/values.yaml      │
└─────────────────────────────────────┘
           ↓ Watches & Manages ↓
┌─────────────────────────────────────┐
│  Custom Resources (Data Plane)          │
│  Deployed via Kustomize                 │
│  Location: {postgres,kafka,redis}/      │
└─────────────────────────────────────┘
           ↓ Creates & Manages ↓
┌─────────────────────────────────────┐
│  Actual Workloads                       │
│  PostgreSQL pods, Kafka brokers, etc    │
└─────────────────────────────────────┘
```

## Example: PostgreSQL

1. **Install CloudNativePG operator** (via Tilt/Helm):

   ```bash
   helm install cloudnative-pg cnpg/cloudnative-pg \
     --values operators/cloudnative-pg/operator-values.yaml
   ```

2. **Create a Cluster CRD** (`postgres/auth-pg.yaml`):

   ```yaml
   apiVersion: postgresql.cnpg.io/v1
   kind: Cluster
   metadata:
     name: auth-pg
   spec:
     instances: 3
   ```

3. **Operator sees the CRD** and creates:
   - 3 PostgreSQL pods
   - Services for read/write endpoints
   - PersistentVolumeClaims for storage
   - Monitoring resources

## Key Difference

| Aspect     | Operator                   | Custom Resource                 |
| ---------- | -------------------------- | ------------------------------- |
| What it is | Controller software        | Workload declaration            |
| Analogy    | Restaurant chef            | Menu order                      |
| Location   | `operators/`               | `postgres/`, `kafka/`, `redis/` |
| Purpose    | Knows how to create things | Says what to create             |
