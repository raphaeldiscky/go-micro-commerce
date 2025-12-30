#!/bin/bash
# Pause GKE cluster by scaling all node pools to zero
# This preserves persistent volumes and data while stopping compute costs
# Usage: ./pause-cluster.sh

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
REGION=$(grep -E '^region\s*=' "$ENV_DIR/terraform.tfvars" | cut -d'"' -f2 || echo "")
PROJECT_ID=$(grep -E '^project_id\s*=' "$ENV_DIR/terraform.tfvars" | cut -d'"' -f2 || echo "")

if [[ -z "$CLUSTER_NAME" ]] || [[ -z "$REGION" ]] || [[ -z "$PROJECT_ID" ]]; then
    log_error "Could not read cluster configuration from terraform.tfvars"
    log_error "Required: cluster_name, region, project_id"
    exit 1
fi

log_info "Cluster: $CLUSTER_NAME"
log_info "Region: $REGION"
log_info "Project: $PROJECT_ID"
echo ""

# Check if cluster exists
if ! gcloud container clusters describe "$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" &> /dev/null; then
    log_error "Cluster '$CLUSTER_NAME' not found in region '$REGION'"
    exit 1
fi

# Get current node pool sizes
log_info "Current node pool status:"
echo ""
gcloud container node-pools list --cluster="$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" \
    --format="table(name,status,config.machineType,autoscaling.enabled,initialNodeCount)"
echo ""

# Confirmation prompt
log_warn "This will scale all node pools to 0 nodes"
log_warn "Persistent volumes and data will be preserved"
echo ""
read -p "Continue? (yes/no): " -r
if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
    log_info "Cancelled by user"
    exit 0
fi

echo ""
log_info "Pausing cluster by scaling node pools to 0..."
echo ""

# Scale stateful-pool to 0
log_info "Scaling stateful-pool to 0 nodes..."
if gcloud container clusters resize "$CLUSTER_NAME" \
    --node-pool stateful-pool \
    --num-nodes 0 \
    --region="$REGION" \
    --project="$PROJECT_ID" \
    --quiet; then
    log_success "stateful-pool scaled to 0"
else
    log_error "Failed to scale stateful-pool"
    exit 1
fi

echo ""

# Scale stateless-pool to 0
log_info "Scaling stateless-pool to 0 nodes..."
if gcloud container clusters resize "$CLUSTER_NAME" \
    --node-pool stateless-pool \
    --num-nodes 0 \
    --region="$REGION" \
    --project="$PROJECT_ID" \
    --quiet; then
    log_success "stateless-pool scaled to 0"
else
    log_error "Failed to scale stateless-pool"
    exit 1
fi

echo ""

# Scale monitoring-pool to 0
log_info "Scaling monitoring-pool to 0 nodes..."
if gcloud container clusters resize "$CLUSTER_NAME" \
    --node-pool monitoring-pool \
    --num-nodes 0 \
    --region="$REGION" \
    --project="$PROJECT_ID" \
    --quiet; then
    log_success "monitoring-pool scaled to 0"
else
    log_error "Failed to scale monitoring-pool"
    exit 1
fi

echo ""

# Scale infra-pool to 0
log_info "Scaling infra-pool to 0 nodes..."
if gcloud container clusters resize "$CLUSTER_NAME" \
    --node-pool infra-pool \
    --num-nodes 0 \
    --region="$REGION" \
    --project="$PROJECT_ID" \
    --quiet; then
    log_success "infra-pool scaled to 0"
else
    log_error "Failed to scale infra-pool"
    exit 1
fi

echo ""

# Scale gateway-pool to 0
log_info "Scaling gateway-pool to 0 nodes..."
if gcloud container clusters resize "$CLUSTER_NAME" \
    --node-pool gateway-pool \
    --num-nodes 0 \
    --region="$REGION" \
    --project="$PROJECT_ID" \
    --quiet; then
    log_success "gateway-pool scaled to 0"
else
    log_error "Failed to scale gateway-pool"
    exit 1
fi

echo ""
log_success "Cluster paused successfully!"
echo ""

# Show final status
log_info "Final node pool status:"
echo ""
gcloud container node-pools list --cluster="$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" \
    --format="table(name,status,config.machineType,autoscaling.enabled,initialNodeCount)"
echo ""

log_info "To resume the cluster, run: ./resume-cluster.sh"
