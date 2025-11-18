# Terraform Infrastructure

Enterprise-grade infrastructure as code for the Go Micro-Commerce platform on Google Cloud Platform (GCP).

## Overview

This Terraform configuration provisions a cost-optimized, production-ready GKE cluster with:

- **5-Tier Node Pool Architecture**: Dedicated pools for stateful (databases), stateless (microservices), monitoring (observability), control-plane (operators), and gateway (ingress) workloads
- **Cost Optimization**: Spot VMs for microservices (~60% savings), cost-effective e2-small for control plane
- **Complete Operator Stack**: CloudNative PostgreSQL, Strimzi Kafka, Redis Operator
- **Full Observability**: Prometheus, Grafana, Loki, Tempo monitoring stack with dedicated pool
- **GitOps Ready**: ArgoCD for application deployments on dedicated control plane pool
- **Production Ingress**: Traefik ingress controller with dedicated gateway pool

### Cost Breakdown

**Note**: Optimized for learning/testing with 250GB total disk limit in asia-southeast2-a zone.

| Component                | Configuration                                             |
| ------------------------ | --------------------------------------------------------- |
| **Stateful Pool**        | 3 × e2-standard-2 (regular VMs, 50GB balanced)            |
| **Stateless Pool**       | 2-10 × e2-medium (Spot VMs, 20GB balanced, autoscaling)   |
| **Monitoring Pool**      | 1-3 × e2-medium (regular VMs, 30GB balanced, autoscaling) |
| **Control Plane Pool**   | 1-2 × e2-small (regular VMs, 15GB balanced, autoscaling)  |
| **Gateway Pool**         | 1-3 × e2-medium (regular VMs, 15GB balanced, autoscaling) |
| **Frontend Hosting**     | Cloudflare Pages (React + Vite)                           |
| **Total Infrastructure** | -                                                         |

**Total Disk Allocation**: 250GB (150GB stateful + 40GB stateless + 30GB monitoring + 15GB control plane + 15GB gateway)

**Savings**: ~60% compared to using all regular (non-Spot) VMs

## Frontend Deployment

**The frontend is NOT managed by Terraform**. Instead, it's deployed via **Cloudflare Pages** for:

- ✅ Zero cost (free tier)
- ✅ Automatic deployments from GitHub
- ✅ Global CDN with 300+ edge locations
- ✅ Built-in SPA routing (no configuration needed)
- ✅ Preview deployments for every PR

**See**: [`CLOUDFLARE_PAGES_SETUP.md`](./CLOUDFLARE_PAGES_SETUP.md) for complete frontend setup instructions.

**Architecture**:

```
Frontend: GitHub → Cloudflare Pages → https://go.micro.commerce.discky.com
Backend:  GKE → Traefik LoadBalancer → https://api.discky.com
```

## Prerequisites

### Required Tools

1. **Terraform** >= 1.13.5

   ```bash
   # Install via official website
   # https://developer.hashicorp.com/terraform/install
   ```

2. **gcloud CLI** (Google Cloud SDK)

   ```bash
   # Install and authenticate
   # https://cloud.google.com/sdk/docs/install
   gcloud auth login
   gcloud config set project YOUR_PROJECT_ID
   ```

3. **kubectl** (Kubernetes CLI)
   ```bash
   # Install via gcloud
   gcloud components install kubectl
   ```

### GCP Setup

1. Create a GCP project or use an existing one
2. Enable required APIs:

   ```bash
   gcloud services enable compute.googleapis.com
   gcloud services enable container.googleapis.com
   gcloud services enable servicenetworking.googleapis.com
   ```

3. Set up authentication:
   ```bash
   gcloud auth application-default login
   ```

## Directory Structure

```
terraform/
├── modules/                          # Reusable Terraform modules
│   ├── gcp-network/                  # VPC network with secondary IP ranges
│   ├── gke-cluster/                  # GKE cluster with 5-tier node pool architecture
│   ├── cloudnative-pg-operator/      # PostgreSQL operator
│   ├── strimzi-kafka-operator/       # Kafka operator (KRaft mode)
│   ├── redis-operator/               # Redis operator
│   ├── monitoring/                   # Prometheus, Grafana, Loki, Tempo
│   ├── argocd/                       # GitOps controller
│   └── traefik/                      # Ingress controller
├── environments/
│   └── prod/                         # Production environment
│       ├── main.tf                   # Module composition
│       ├── variables.tf              # Variable definitions
│       ├── outputs.tf                # Output definitions
│       ├── providers.tf              # Provider configuration
│       ├── backend.hcl               # Backend configuration
│       ├── terraform.tfvars.example  # Example configuration
│       └── terraform.tfvars          # Your actual config (gitignored)
├── shared/                           # Shared configurations
│   ├── backend.tf                    # Backend template
│   └── versions.tf                   # Provider versions
└── scripts/                          # Deployment scripts
    ├── init-backend.sh               # Initialize GCS backend
    ├── plan-prod.sh                  # Plan infrastructure changes
    ├── apply-prod.sh                 # Apply infrastructure
    └── destroy-prod.sh               # Destroy infrastructure
```

## Infrastructure Components

### 1. Network Module (`gcp-network`)

Creates VPC network with:

- Primary subnet for nodes
- Secondary IP ranges for pods and services (VPC-native cluster)
- Cloud NAT for private node internet access
- VPC flow logs for monitoring

### 2. GKE Cluster Module (`gke-cluster`)

Provisions GKE cluster with 5-tier node pool architecture:

**Stateful Pool** (Databases: PostgreSQL, Kafka, Redis)

- 3 × e2-standard-2 nodes (2 vCPU, 8GB RAM)
- 50GB balanced persistent disk per node (150GB total)
- Regular VMs for reliability
- Fixed node count (no autoscaling)
- Taint: `workload-type=stateful:NoSchedule`

**Stateless Pool** (Microservices)

- 2-10 × e2-medium nodes (1 vCPU, 4GB RAM)
- 20GB balanced persistent disk per node (40-200GB total)
- **Spot VMs** for 60-91% cost savings
- Autoscaling based on workload
- Taint: `cloud.google.com/gke-spot=true:NoSchedule`

**Monitoring Pool** (Observability: Prometheus, Grafana, Loki, Tempo, Alloy)

- 1-3 × e2-medium nodes (1 vCPU, 4GB RAM)
- 30GB balanced persistent disk per node (30-90GB total)
- Regular VMs for monitoring reliability
- Autoscaling based on metrics load
- Taint: `workload-type=monitoring:NoSchedule`

**Control Plane Pool** (Operators, ArgoCD, ESO)

- 1-2 × e2-small nodes (0.5 vCPU, 2GB RAM)
- 15GB balanced persistent disk per node (15-30GB total)
- Regular VMs for control plane reliability
- Autoscaling for operator workloads
- Taint: `workload-type=control-plane:NoSchedule`

**Gateway Pool** (Traefik, Apollo Router, API Gateway)

- 1-3 × e2-medium nodes (1 vCPU, 4GB RAM)
- 15GB balanced persistent disk per node (15-45GB total)
- Regular VMs for gateway reliability
- Autoscaling based on traffic
- Taint: `workload-type=gateway:NoSchedule`

**Security Features**

- **Private nodes enabled** (nodes have no external IPs)
- **Cloud NAT** for outbound internet access
- **Public control plane endpoint** (for kubectl access)
- Workload Identity enabled
- Shielded nodes enabled
- Private Google access
- GKE managed Prometheus

### 3. Operator Modules

**CloudNative PostgreSQL Operator** (`cloudnative-pg-operator`)

- Helm chart version: 0.26.1
- Namespace: `cnpg-system`
- Manages 9 PostgreSQL instances for microservices

**Strimzi Kafka Operator** (`strimzi-kafka-operator`)

- Helm chart version: 0.48.0
- Namespace: `kafka-system`
- KRaft mode enabled (no Zookeeper)

**Redis Operator** (`redis-operator`)

- Helm chart version: 0.22.2
- Namespace: `redis-system`
- Supports cluster mode (3 master + 3 replica)

**External Secrets Operator** (`external-secrets-operator`)

- Helm chart version: 1.0.0
- Namespace: `external-secrets`
- Syncs secrets from Google Secret Manager to Kubernetes Secrets
- Uses Workload Identity for secure authentication
- Automatic hourly sync of secrets
- ClusterSecretStore: `gcp-secret-manager`
- **See**: [`EXTERNAL_SECRETS_SETUP.md`](./EXTERNAL_SECRETS_SETUP.md) for complete setup guide

### 4. Platform Modules

**Monitoring Stack** (`monitoring`)

- **kube-prometheus-stack**: Prometheus + Grafana (chart v79.5.0)
- **Loki**: Log aggregation (chart v6.46.0)
- **Tempo**: Distributed tracing (chart v1.24.0)
- Persistent storage for all components

**ArgoCD** (`argocd`)

- Helm chart version: 9.1.2
- Namespace: `argocd`
- Optional bootstrap ApplicationSet for git repo sync

**Traefik** (`traefik`)

- Helm chart version: 37.3.0
- Namespace: `traefik`
- LoadBalancer service type
- Dashboard enabled with Prometheus metrics

## Deployment Workflow

### Step 1: Configure Environment

1. Navigate to the production environment:

   ```bash
   cd terraform/environments/prod
   ```

2. Copy the example configuration:

   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

3. Edit `terraform.tfvars` and update:

   ```hcl
   # REQUIRED: Update with your GCP project ID
   project_id = "your-gcp-project-id"

   # Optional: Customize region/zone (default: asia-southeast2)
   region = "asia-southeast2"
   zone   = "asia-southeast2-a"

   # Optional: Customize cluster configuration
   cluster_name = "go-micro-commerce-prod"

   # Optional: Set Grafana admin password
   grafana_admin_password = "your-strong-password"
   ```

### Step 2: Initialize Backend

Initialize Terraform with GCS backend for state storage:

```bash
./terraform/scripts/init-backend.sh prod
```

This script:

- Creates GCS bucket for Terraform state (if not exists)
- Enables versioning for state file recovery
- Initializes Terraform with backend configuration

### Step 3: Plan Infrastructure

Preview the infrastructure changes:

```bash
./terraform/scripts/plan-prod.sh
```

Review the plan output carefully before applying.

### Step 4: Apply Infrastructure

Deploy the infrastructure:

```bash
./terraform/scripts/apply-prod.sh
```

This will:

1. Create VPC network and subnets
2. Provision GKE cluster with 5-tier node pool architecture
3. Install External Secrets Operator with Google Secret Manager integration
4. Install all operators (PostgreSQL, Kafka, Redis)
5. Deploy monitoring stack (Prometheus, Grafana, Loki, Tempo)
6. Install ArgoCD and Traefik
7. Configure kubectl with cluster credentials
8. Set up Cloudflare DNS for backend API

**Deployment time**: ~15-20 minutes

**If failed**:

```bash
gcloud container clusters get-credentials go-micro-commerce-prod --zone=asia-southeast2-a --project=go-micro-commerce
./terraform/scripts/apply-prod.sh
```

### Step 5: Verify Deployment

After successful deployment:

1. Check cluster status:

   ```bash
   kubectl cluster-info
   kubectl get nodes -o wide
   ```

2. Verify namespaces:

   ```bash
   kubectl get namespaces
   ```

3. Check node pools:
   ```bash
   gcloud container node-pools list --cluster=go-micro-commerce-prod --zone=asia-southeast2-a
   ```

### Step 6: Access Platform Services

**Grafana Dashboard**

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
# Visit: http://localhost:3000
# Default credentials: admin / <password from terraform.tfvars>
```

**ArgoCD UI**

```bash
kubectl port-forward -n argocd svc/argocd-server 8080:443
# Visit: https://localhost:8080
# Get password: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

**Traefik Dashboard**

```bash
kubectl port-forward -n traefik svc/traefik 9000:9000
# Visit: http://localhost:9000/dashboard/
```

**Prometheus UI**

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090
# Visit: http://localhost:9090
```

### Step 7: Populate Secrets with External Secrets Operator

Before deploying applications, populate all required secrets in Google Secret Manager:

1. **Run the secret population script**:

   ```bash
   ./terraform/scripts/populate-secrets.sh
   ```

   This interactive script will guide you through creating:
   - JWT private/public keys for authentication
   - Stripe API keys for payment processing
   - SMTP credentials for notifications
   - Database passwords
   - Monitoring credentials

2. **Deploy ExternalSecret resources**:

   ```bash
   kubectl apply -k deployments/k8s/base/external-secrets/
   ```

3. **Verify secrets are synced**:

   ```bash
   kubectl get externalsecrets -A
   kubectl get secrets | grep external-secrets
   ```

**Complete setup guide**: See [`EXTERNAL_SECRETS_SETUP.md`](./EXTERNAL_SECRETS_SETUP.md) for detailed instructions on secret management, rotation, and troubleshooting.

### Step 8: Deploy Frontend (Cloudflare Pages)

After backend infrastructure is ready, deploy the frontend:

1. **Get Traefik LoadBalancer IP**:

   ```bash
   terraform output traefik_load_balancer_ip
   ```

2. **Follow Cloudflare Pages setup**:
   - See: [`CLOUDFLARE_PAGES_SETUP.md`](./CLOUDFLARE_PAGES_SETUP.md)
   - Configure environment variables with API URL
   - Connect GitHub repository
   - Automatic deployments will handle the rest

3. **Verify deployment**:
   - Frontend: https://go.micro.commerce.discky.com
   - Backend API: https://api.discky.com/health

**Note**: Frontend deployment is separate from Terraform and happens automatically via Cloudflare Pages when you push to GitHub.

## Workload Scheduling

### Stateful Workloads (Databases: PostgreSQL, Kafka, Redis)

To schedule on the stateful pool, add tolerations and node selectors:

```yaml
spec:
  tolerations:
    - key: "workload-type"
      operator: "Equal"
      value: "stateful"
      effect: "NoSchedule"
  nodeSelector:
    workload-type: "stateful"
```

### Stateless Workloads (Microservices)

To schedule on the stateless pool (Spot VMs):

```yaml
spec:
  tolerations:
    - key: "cloud.google.com/gke-spot"
      operator: "Equal"
      value: "true"
      effect: "NoSchedule"
  nodeSelector:
    workload-type: "stateless"
```

### Monitoring Workloads (Prometheus, Grafana, Loki, Tempo, Alloy)

To schedule on the monitoring pool:

```yaml
spec:
  tolerations:
    - key: "workload-type"
      operator: "Equal"
      value: "monitoring"
      effect: "NoSchedule"
  nodeSelector:
    workload-type: "monitoring"
```

### Control Plane Workloads (Operators, ArgoCD, ESO)

To schedule on the control plane pool:

```yaml
spec:
  tolerations:
    - key: "workload-type"
      operator: "Equal"
      value: "control-plane"
      effect: "NoSchedule"
  nodeSelector:
    workload-type: "control-plane"
```

### Gateway Workloads (Traefik, Apollo Router, API Gateway)

To schedule on the gateway pool:

```yaml
spec:
  tolerations:
    - key: "workload-type"
      operator: "Equal"
      value: "gateway"
      effect: "NoSchedule"
  nodeSelector:
    workload-type: "gateway"
```

## ArgoCD Application Deployment

Once infrastructure is deployed, use ArgoCD to manage application deployments:

1. Update ArgoCD configuration in `terraform.tfvars`:

   ```hcl
   argocd_git_repo_url      = "https://github.com/your-org/go-micro-commerce.git"
   argocd_git_repo_path     = "deployments/k8s"
   argocd_enable_bootstrap  = true
   ```

2. Re-apply Terraform:

   ```bash
   ./terraform/scripts/apply-prod.sh
   ```

3. ArgoCD will automatically sync applications from `deployments/k8s/` directory

## Managing Infrastructure

### View Outputs

Show all infrastructure outputs:

```bash
cd terraform/environments/prod
terraform output
```

Show specific output:

```bash
terraform output cluster_name
terraform output cost_summary
terraform output kubeconfig_command
```

### Update Infrastructure

1. Modify `terraform.tfvars` or module configurations
2. Run plan to preview changes:
   ```bash
   ./terraform/scripts/plan-prod.sh
   ```
3. Apply changes:
   ```bash
   ./terraform/scripts/apply-prod.sh
   ```

### Destroy Infrastructure

**WARNING**: This will delete all infrastructure!

```bash
./terraform/scripts/destroy-prod.sh
```

The script includes safety checks:

- Requires typing "yes" to confirm
- Requires typing the cluster name to confirm
- 3-second countdown before destruction

## Troubleshooting

### Issue: Backend initialization fails

**Solution**: Ensure you're authenticated with gcloud:

```bash
gcloud auth application-default login
gcloud config set project YOUR_PROJECT_ID
```

### Issue: Provider plugins fail to install

**Solution**: Clear Terraform cache and re-initialize:

```bash
cd terraform/environments/prod
rm -rf .terraform .terraform.lock.hcl
terraform init -backend-config=backend.hcl -reconfigure
```

### Issue: Helm releases fail to install

**Solution**: Check kubectl configuration:

```bash
kubectl cluster-info
kubectl get nodes
```

If not configured, get credentials:

```bash
gcloud container clusters get-credentials go-micro-commerce-prod --zone=asia-southeast2-a
```

### Issue: Spot VMs are preempted frequently

**Solution**: Adjust autoscaling settings in `terraform.tfvars`:

```hcl
stateless_pool_min_nodes = 3  # Increase minimum nodes
```

### Issue: Node pool runs out of capacity

**Solution**:

1. Check pod distribution:

   ```bash
   kubectl top nodes
   kubectl describe nodes
   ```

2. Increase max nodes for stateless pool:
   ```hcl
   stateless_pool_max_nodes = 15
   ```

## Security Best Practices

1. **State File Security**
   - State bucket is private by default
   - Enable versioning for recovery
   - Consider encrypting state with customer-managed keys

2. **Secrets Management**
   - Never commit `terraform.tfvars` to git
   - Use GCP Secret Manager for sensitive values
   - Rotate passwords regularly

3. **Network Security**
   - Private GKE nodes (no public IPs)
   - Cloud NAT for outbound traffic
   - VPC flow logs enabled for monitoring

4. **Access Control**
   - Use Workload Identity for pod-to-GCP authentication
   - Enable binary authorization for production
   - Implement pod security policies

## Cost Optimization Tips

1. **Right-size node pools**: Monitor actual resource usage and adjust machine types
2. **Use Spot VMs**: Already configured for stateless pool (60% savings)
3. **Enable autoscaling**: Already configured for stateless pool
4. **Set resource requests/limits**: Ensures efficient pod packing
5. **Use committed use discounts**: For predictable workloads (stateful pool)

## Monitoring and Alerts

### Key Metrics to Monitor

- **Node utilization**: CPU, memory, disk usage
- **Pod status**: Restarts, failures, pending
- **Spot VM preemptions**: Track preemption rate
- **Cost tracking**: Use GCP cost management tools

### Access Monitoring

All metrics are available in Grafana dashboards:

- Kubernetes cluster overview
- Node resource usage
- Pod resource usage
- Kafka metrics (if enabled)
- PostgreSQL metrics (if enabled)
- Redis metrics (if enabled)

## Provider Versions

- **Terraform**: >= 1.13.5
- **Google Provider**: ~> 7.11
- **Kubernetes Provider**: ~> 2.38
- **Helm Provider**: ~> 3.1.0
- **Kubectl Provider**: ~> 1.19

Version constraints are managed in `shared/versions.tf`.

## Support and Contributions

For issues or questions:

1. Check troubleshooting section above
2. Review Terraform plan output
3. Check GKE cluster logs
4. Consult GCP documentation

## License

This Terraform configuration is part of the Go Micro-Commerce project.
