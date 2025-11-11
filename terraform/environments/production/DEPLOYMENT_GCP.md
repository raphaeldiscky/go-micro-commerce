# GCP Deployment Guide for go-micro-commerce

## Overview

This guide covers deploying the go-micro-commerce microservices platform on GCP using:

- **Talos Linux** v1.11.5 (immutable Kubernetes OS)
- **3 VMs** (e2-standard-2: 2 vCPUs, 8GB RAM each)
- **GCP Secret Manager** for secrets via External Secrets Operator
- **Longhorn** for distributed block storage
- **CloudNativePG** for PostgreSQL clusters
- **Terraform** for infrastructure as code
- **Domain**: discky.com

## Architecture

```
GCP asia-southeast2-a (Jakarta)
├── Control Plane: 1x e2-standard-2 (2 vCPU, 8GB RAM)
├── Worker 1: 1x e2-standard-2 (2 vCPU, 8GB RAM)
└── Worker 2: 1x e2-standard-2 (2 vCPU, 8GB RAM)

Storage:
- Longhorn (2-replica across workers)
- GCS bucket: discky-storage

Secrets:
- GCP Secret Manager (19 secrets)
- External Secrets Operator (syncs to K8s)

Services:
- 9 microservices (auth, product, order, payment, cart, fulfillment, notification, search, chat)
- 9 PostgreSQL clusters (CloudNativePG)
- Kafka, Redis, etc
```

## Prerequisites

✅ 3 GCP VMs (e2-standard-2) running Debian 12
✅ GCP Project with billing enabled
✅ gcloud CLI installed and configured
✅ Terraform >= 1.12.2
✅ kubectl
✅ talosctl

## Deployment Steps

### Phase 1: GCP Setup (15 minutes)

#### 1.1 Authenticate and Configure

```bash
# Authenticate with GCP
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
gcloud config set compute/zone asia-southeast2-a
```

#### 1.2 Run GCP Setup Script

```bash
./scripts/gcp-setup.sh
```

This script:

- Enables required GCP APIs
- Creates service account for External Secrets Operator
- Downloads service account key to `~/external-secrets-sa-key.json`
- Displays your VM information

#### 1.3 Note Your VM IPs

```bash
gcloud compute instances list --zones=asia-southeast2-a
```

Save the internal IPs - you'll need them for terraform.tfvars.

#### 1.4 Create GCP Secrets

```bash
./scripts/create-gcp-secrets.sh
```

Creates 19 secrets in GCP Secret Manager:

- PostgreSQL credentials for 9 services
- JWT secret for auth service

Verify:

```bash
gcloud secrets list
```

### Phase 2: Install Talos on VMs (30 minutes)

#### 2.1 Generate Custom Talos Image

1. Visit https://factory.talos.dev/
2. Configure:
   - Talos version: `v1.11.5`
   - Platform: `metal`
   - Architecture: `amd64`
   - Extensions:
     - `siderolabs/iscsi-tools`
     - `siderolabs/util-linux-tools`
3. Copy schematic ID
4. Image URL: `factory.talos.dev/installer/<schematic-id>:v1.11.5`

#### 2.2 Install Talos via kexec on Each VM

**For Control Plane:**

```bash
# Copy script to control-plane-1
gcloud compute scp scripts/talos-kexec-install.sh control-plane-1:/tmp/ \
    --project=go-micro-commerce \
    --zone=asia-southeast2-a

# Then SSH back in to run it
gcloud compute ssh control-plane-1 \
    --project=go-micro-commerce \
    --zone=asia-southeast2-a

# Once inside, run the script as root
sudo bash /tmp/talos-kexec-install.sh
# VM will reboot into Talos
```

**Repeat for Worker 1 and Worker 2**

After all VMs reboot, Talos runs in RAM. Terraform will persist to disk.

**Verify Talos is Running**

gcloud compute instances get-serial-port-output control-plane-1 \
 --project=go-micro-commerce \
 --zone=asia-southeast2-a \
 --port=1

**Important**: After kexec, VMs will boot into Talos Linux maintenance mode and be accessible only via Talos API (port 50000).

### Phase 3: Deploy Cluster with Terraform

#### 3.1 Configure Terraform

```bash
cd terraform/environments/production

# Copy example config
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
nano terraform.tfvars
```

Update these values:

```hcl
gcp_project_id = "your-project-id"
gcp_sa_key_path = "/home/youruser/external-secrets-sa-key.json"

cluster_name       = "go-micro-commerce-prod"
kubernetes_version = "1.34.1"
talos_version      = "v1.11.5"

control_plane_nodes = [
  { name = "control-plane-1", ip = "10.x.x.x" }  # Your actual IP
]

worker_nodes = [
  { name = "worker-1", ip = "10.x.x.x" },  # Your actual IP
  { name = "worker-2", ip = "10.x.x.x" }   # Your actual IP
]

control_plane_external_ips = ["34.101.240.213"]
worker_external_ips        = ["34.101.106.58", "34.101.86.117"]

cluster_endpoint = "https://10.x.x.x:6443"  # Control plane IP
install_image = "factory.talos.dev/installer/<your-schematic-id>:v1.11.5"

enable_longhorn    = true
longhorn_disk_path = "/var/lib/longhorn"

enable_external_secrets = true
```

### 3.2 Run Terraform on Bastion VM inside VPC

```bash
# Copy terraform directory from correct location from local host
gcloud compute scp --recurse \
    $HOME/repos/go-micro-commerce/terraform \
    bastion:~/ \
    --zone=asia-southeast2-a \
    --project=go-micro-commerce

# Copy service account key
gcloud compute scp \
    $HOME/external-secrets-sa-key.json \
    bastion:~/ \
    --zone=asia-southeast2-a \
    --project=go-micro-commerce

# SSH into bastion
 gcloud compute ssh bastion \
   --zone=asia-southeast2-a \
   --project=go-micro-commerce

# Install Terraform with matching version with config
# https://developer.hashicorp.com/terraform/install

# Test PING
ping -c 2 10.x.x.x  # control-plane-1
ping -c 2 10.x.x.x  # worker-1
ping -c 2 10.x.x.x  # worker-2

# Test Talos API Port
timeout 5 bash -c "cat < /dev/null > /dev/tcp/10.x.x.x/50000" && echo "Port 50000 open on control-plane" || echo "Failed"
timeout 5 bash -c "cat < /dev/null > /dev/tcp/10.x.x.x/50000" && echo "Port 50000 open on worker-1" || echo "Failed"
timeout 5 bash -c "cat < /dev/null > /dev/tcp/10.x.x.x/50000" && echo "Port 50000 open on worker-2" || echo "Failed"
# Expected: All show "Port 50000 open"

cd terraform/environments/production
chmod +x deploy-phase1.sh
chmod +x deploy-phase2.sh
chmod +x deploy.sh

# Run automated deployment
./deploy.sh
```

The script will:

1. Check prerequisites (terraform, talosctl, kubectl, helm)
2. Initialize Terraform
3. **Phase 1**: Deploy Talos cluster and generate kubeconfig
4. Wait for Kubernetes API to be ready
5. Install Cilium CNI (networking)
6. **Phase 2**: Deploy Kubernetes resources (CNI, Longhorn, etc.)
7. Verify deployment
8. Show next steps

### Option 2: Manual Deployment

If you prefer manual control or troubleshooting:

#### Phase 1: Deploy Talos Cluster

```bash
# Initialize Terraform
terraform init

# Review plan
terraform plan -target=module.talos_cluster

# Apply (creates cluster and kubeconfig)
terraform apply -target=module.talos_cluster
```

**What happens:**

- Generates Talos machine configurations
- Applies configurations to all nodes
- Bootstraps the cluster
- Generates `./kubeconfig-production` and `./talosconfig-production`

**Wait for cluster**: The bootstrap process takes ~30-60 seconds.

Verify cluster is ready:

```bash
export KUBECONFIG=./kubeconfig-production
kubectl cluster-info
kubectl get nodes
```

#### Phase 2: Deploy Kubernetes Resources

```bash
# Review plan
terraform plan

# Apply (deploys CNI, storage, operators)
terraform apply
```

**What happens:**

- Deploys Longhorn (storage)
- Deploys CloudNativePG operator
- Deploys External Secrets Operator (if enabled)

## Post-Deployment

### 1. Configure Talos CLI

```bash
# Export Talos configuration
export TALOSCONFIG=$(pwd)/talosconfig-production

# Or merge into default talosconfig
talosctl config merge talosconfig-production

# Set default endpoint
talosctl config endpoint 10.184.0.8 10.184.0.9 10.184.0.10

# Verify connection
talosctl health --nodes 10.184.0.8,10.184.0.9,10.184.0.10
```

### 2. Configure kubectl

```bash
# Option A: Use directly
export KUBECONFIG=$(pwd)/kubeconfig-production

# Option B: Merge into default kubeconfig
KUBECONFIG=~/.kube/config:$(pwd)/kubeconfig-production \
  kubectl config view --flatten > ~/.kube/config.new
mv ~/.kube/config.new ~/.kube/config
```

### 3. Verify Deployment

```bash
# Check nodes
kubectl get nodes -o wide

# Check CNI (Cilium)
kubectl get pods -n kube-system -l k8s-app=cilium
kubectl exec -n kube-system ds/cilium -- cilium status

# Check Longhorn
kubectl get pods -n longhorn-system
kubectl get storageclass

# Check External Secrets (if enabled)
kubectl get pods -n external-secrets-system
```

### 4. Access Longhorn UI (Optional)

```bash
# Port-forward to local machine
kubectl port-forward -n longhorn-system svc/longhorn-frontend 8080:80

# Access at http://localhost:8080
```

## Architecture Details

### Why kexec Instead of Native GCP Images?

**Reason**: Talos doesn't provide pre-built GCP images, only AWS/Azure.

**Approach**: Install Talos on existing Debian VMs via kexec (kernel execute). This boots Talos without reprovisioning VMs.

**Platform**: Use `platform=metal` (not `platform=gcp`) because kexec installation is equivalent to bare metal.

### Certificate Configuration

Talos uses TLS certificates for API access. Certificates must include all access IPs:

- **Internal IPs** (10.184.0.x): Used for Terraform connections from bastion
- **External IPs** (34.101.x.x): Used for external access (optional)

Both are included in `certSANs` for flexibility.

### Two-Phase Terraform Apply

**Why needed?** Kubernetes and Helm providers need the kubeconfig file, which is generated by the Talos module in the same run.

**Solution**: Use targeted apply:

1. Phase 1 creates cluster and kubeconfig
2. Phase 2 uses kubeconfig to deploy Kubernetes resources

## Troubleshooting

### Terraform Connection Timeout

**Error**: `timeout while waiting for state to become 'ready'`

**Cause**: Terraform cannot reach internal IPs from outside VPC.

**Solution**: Run Terraform from bastion VM inside GCP VPC.

### Certificate Validation Error

**Error**: `x509: certificate is valid for X, not Y`

**Causes**:

1. Using wrong IP (external vs internal)
2. VMs retain old certificates from previous attempts

**Solutions**:

1. Verify `terraform.tfvars` uses **internal IPs** for node connections
2. Reset VMs to clean state:
   ```bash
   # Re-run kexec installation on each node
   # This clears old certificates
   ```

### Kubernetes Provider Initialization Error

**Error**: `'config_path' refers to an invalid path: "./kubeconfig-production": stat ./kubeconfig-production: no such file or directory`

**Cause**: The Kubernetes provider in main.tf tries to validate the kubeconfig path during provider initialization, before Phase 1 has created the file.

**Solutions**:

1. **Automated (via deploy.sh)**: The deploy script automatically creates a placeholder kubeconfig before terraform init.

2. **Manual workaround**: Create a dummy kubeconfig before running terraform:

   ```bash
   cat > ./kubeconfig-production << 'EOF'
   apiVersion: v1
   kind: Config
   clusters:
   - cluster:
       server: https://127.0.0.1:6443
     name: placeholder
   contexts:
   - context:
       cluster: placeholder
       user: placeholder
     name: placeholder
   current-context: placeholder
   preferences: {}
   users:
   - name: placeholder
     user:
       token: placeholder
   EOF
   chmod 600 ./kubeconfig-production
   ```

   Then run terraform as normal. The real kubeconfig will overwrite this placeholder during Phase 1.

3. **Code fix**: The main.tf uses try() to handle missing kubeconfig gracefully during Phase 1 deployment.

### VMs Stuck at Boot

**Symptom**: Serial console shows "EFI stub: Loaded initrd..." and hangs.

**Cause**: Previous Terraform apply installed Talos to disk, creating corrupted boot entries.

**Solution**: Delete and recreate VMs, then re-run kexec installation.

### Kubernetes API Not Ready

**Error**: `connection refused` when accessing Kubernetes API

**Check**:

```bash
# Verify bootstrap completed
talosctl bootstrap -n 10.184.0.8

# Check cluster health
talosctl health -n 10.184.0.8

# Check etcd
talosctl etcd members -n 10.184.0.8

# Get kubeconfig again
talosctl kubeconfig -n 10.184.0.8
```

### Longhorn Pods CrashLoopBackOff

**Cause**: Missing iSCSI kernel modules.

**Solution**: Verify custom Talos image includes required extensions:

```bash
# Check loaded modules
talosctl -n 10.184.0.9 get extensions

# Should see: iscsi-tools, util-linux-tools
```

If missing, update `install_image` in `terraform.tfvars` with correct Image Factory URL.

## Maintenance

### Updating Cluster

```bash
# Update variables in terraform.tfvars
# Review changes
terraform plan

# Apply changes
terraform apply
```

### Upgrading Talos

See [Talos Upgrade Guide](https://www.talos.dev/latest/talos-guides/upgrading-talos/)

```bash
talosctl upgrade --nodes 10.184.0.9,10.184.0.10 \
  --image factory.talos.dev/installer/YOUR_IMAGE_ID:v1.12.0

# Upgrade control plane last
talosctl upgrade --nodes 10.184.0.8 \
  --image factory.talos.dev/installer/YOUR_IMAGE_ID:v1.12.0
```

### Accessing Node Logs

```bash
# System logs
talosctl logs -n 10.184.0.8 -f

# Kubernetes logs
talosctl logs -n 10.184.0.8 -k -f

# Specific service
talosctl logs -n 10.184.0.8 machined
```

### Destroying Cluster

```bash
# Destroy all resources
terraform destroy

# Manually delete VMs if needed
gcloud compute instances delete control-plane-1 worker-1 worker-2 --zone=asia-southeast2-a
```

## Security Considerations

1. **Sensitive Files**: Never commit these files:
   - `terraform.tfvars` (contains GCP project ID, SA key path)
   - `terraform.tfstate` (contains secrets, IPs)
   - `kubeconfig-production` (cluster admin access)
   - `talosconfig-production` (node admin access)
   - `*.json` (service account keys)

2. **Service Account Permissions**: Grant minimum required permissions to GCP SA.

3. **Network Security**: Consider restricting access:

   ```bash
   # Allow only bastion to reach Talos API (port 50000)
   gcloud compute firewall-rules create allow-talos-from-bastion \
     --allow=tcp:50000 \
     --source-ranges=10.184.0.7/32
   ```

4. **Preemptible VMs**: External IPs change on restart. Update `control_plane_external_ips` and `worker_external_ips` after VM restarts.

## Files Generated

After successful deployment:

- `./kubeconfig-production` - Kubernetes cluster access
- `./talosconfig-production` - Talos node management
- `./terraform.tfstate` - Terraform state (contains secrets)
- `./.terraform/` - Terraform plugins and modules

**Backup**: Store these securely. Losing `talosconfig` means losing node access.

## References

- [Talos Linux Documentation](https://www.talos.dev/latest/)
- [Talos Image Factory](https://factory.talos.dev/)
- [Terraform Talos Provider](https://registry.terraform.io/providers/siderolabs/talos/latest/docs)
- [Longhorn Documentation](https://longhorn.io/docs/)
- [Cilium Documentation](https://docs.cilium.io/)
- [External Secrets Operator](https://external-secrets.io/)
