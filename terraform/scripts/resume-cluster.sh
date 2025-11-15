#!/bin/bash
# Resume GKE cluster by scaling node pools back to operational sizes
# Usage: ./resume-cluster.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="${SCRIPT_DIR}/.."
ENV="prod"
ENV_DIR="${TERRAFORM_DIR}/environments/${ENV}"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${BLUE}[SUCCESS]${NC} $1"
}

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    log_error "gcloud CLI is not installed. Please install it first."
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl is not installed. Please install it first."
    exit 1
fi

# Check if environment directory exists
if [[ ! -d "$ENV_DIR" ]]; then
    log_error "Environment directory not found: $ENV_DIR"
    exit 1
fi

# Check if terraform.tfvars exists
if [[ ! -f "$ENV_DIR/terraform.tfvars" ]]; then
    log_error "terraform.tfvars not found in: $ENV_DIR"
    exit 1
fi

# Read cluster configuration from terraform.tfvars
CLUSTER_NAME=$(grep -E '^cluster_name\s*=' "$ENV_DIR/terraform.tfvars" | cut -d'"' -f2 || echo "")
ZONE=$(grep -E '^zone\s*=' "$ENV_DIR/terraform.tfvars" | cut -d'"' -f2 || echo "")
PROJECT_ID=$(grep -E '^project_id\s*=' "$ENV_DIR/terraform.tfvars" | cut -d'"' -f2 || echo "")

if [[ -z "$CLUSTER_NAME" ]] || [[ -z "$ZONE" ]] || [[ -z "$PROJECT_ID" ]]; then
    log_error "Could not read cluster configuration from terraform.tfvars"
    log_error "Required: cluster_name, zone, project_id"
    exit 1
fi

log_info "Cluster: $CLUSTER_NAME"
log_info "Zone: $ZONE"
log_info "Project: $PROJECT_ID"
echo ""

# Check if cluster exists
if ! gcloud container clusters describe "$CLUSTER_NAME" --zone="$ZONE" --project="$PROJECT_ID" &> /dev/null; then
    log_error "Cluster '$CLUSTER_NAME' not found in zone '$ZONE'"
    exit 1
fi

# Get current node pool sizes
log_info "Current node pool status:"
echo ""
gcloud container node-pools list --cluster="$CLUSTER_NAME" --zone="$ZONE" --project="$PROJECT_ID" \
    --format="table(name,status,config.machineType,autoscaling.enabled,initialNodeCount)"
echo ""

log_info "Resuming cluster by scaling node pools..."
echo ""

# Scale stateful-pool to 3 nodes
log_info "Scaling stateful-pool to 3 nodes..."
if gcloud container clusters resize "$CLUSTER_NAME" \
    --node-pool stateful-pool \
    --num-nodes 3 \
    --zone="$ZONE" \
    --project="$PROJECT_ID" \
    --quiet; then
    log_success "stateful-pool scaled to 3 nodes"
else
    log_error "Failed to scale stateful-pool"
    exit 1
fi

echo ""

# Scale stateless-pool to 2 nodes (minimum for autoscaling)
log_info "Scaling stateless-pool to 2 nodes..."
if gcloud container clusters resize "$CLUSTER_NAME" \
    --node-pool stateless-pool \
    --num-nodes 2 \
    --zone="$ZONE" \
    --project="$PROJECT_ID" \
    --quiet; then
    log_success "stateless-pool scaled to 2 nodes"
else
    log_error "Failed to scale stateless-pool"
    exit 1
fi

echo ""

# Scale monitoring-pool to 1 node (minimum for autoscaling)
log_info "Scaling monitoring-pool to 1 node..."
if gcloud container clusters resize "$CLUSTER_NAME" \
    --node-pool monitoring-pool \
    --num-nodes 1 \
    --zone="$ZONE" \
    --project="$PROJECT_ID" \
    --quiet; then
    log_success "monitoring-pool scaled to 1 node"
else
    log_error "Failed to scale monitoring-pool"
    exit 1
fi

echo ""

# Scale control-plane-pool to 1 node (minimum for autoscaling)
log_info "Scaling control-plane-pool to 1 node..."
if gcloud container clusters resize "$CLUSTER_NAME" \
    --node-pool control-plane-pool \
    --num-nodes 1 \
    --zone="$ZONE" \
    --project="$PROJECT_ID" \
    --quiet; then
    log_success "control-plane-pool scaled to 1 node"
else
    log_error "Failed to scale control-plane-pool"
    exit 1
fi

echo ""

# Scale gateway-pool to 1 node (minimum for autoscaling)
log_info "Scaling gateway-pool to 1 node..."
if gcloud container clusters resize "$CLUSTER_NAME" \
    --node-pool gateway-pool \
    --num-nodes 1 \
    --zone="$ZONE" \
    --project="$PROJECT_ID" \
    --quiet; then
    log_success "gateway-pool scaled to 1 node"
else
    log_error "Failed to scale gateway-pool"
    exit 1
fi

echo ""
log_info "Waiting for nodes to become ready..."
echo ""

# Get cluster credentials
gcloud container clusters get-credentials "$CLUSTER_NAME" --zone="$ZONE" --project="$PROJECT_ID"

# Wait for nodes to be ready
log_info "Checking node status..."
RETRY_COUNT=0
MAX_RETRIES=30

while [[ $RETRY_COUNT -lt $MAX_RETRIES ]]; do
    READY_NODES=$(kubectl get nodes --no-headers 2>/dev/null | grep -c " Ready " || echo "0")
    TOTAL_NODES=$(kubectl get nodes --no-headers 2>/dev/null | wc -l || echo "0")

    if [[ $READY_NODES -eq 8 ]]; then
        log_success "All nodes are ready! ($READY_NODES/8)"
        break
    fi

    echo -ne "\r  Ready: $READY_NODES/8 nodes (waiting...)"
    sleep 10
    ((RETRY_COUNT++))
done

echo ""
echo ""

if [[ $READY_NODES -eq 8 ]]; then
    log_success "Cluster resumed successfully!"
else
    log_warn "Some nodes may still be initializing ($READY_NODES/8 ready)"
    log_info "Run 'kubectl get nodes' to check node status"
fi

echo ""

# Show final status
log_info "Final node pool status:"
echo ""
gcloud container node-pools list --cluster="$CLUSTER_NAME" --zone="$ZONE" --project="$PROJECT_ID" \
    --format="table(name,status,config.machineType,autoscaling.enabled,initialNodeCount)"
echo ""

log_info "Node status:"
kubectl get nodes -o wide
echo ""

log_info "Cluster information:"
kubectl cluster-info
echo ""

log_info "To pause the cluster again, run: ./pause-cluster.sh"
