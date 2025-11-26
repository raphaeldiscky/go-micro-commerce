#!/bin/bash
# Check GKE cluster and node pool status
# Usage: ./cluster-status.sh

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
    log_warn "Cluster may not be created yet"
    exit 1
fi

# Get cluster status
CLUSTER_STATUS=$(gcloud container clusters describe "$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" --format="value(status)")

log_info "Cluster Status: $CLUSTER_STATUS"
echo ""

# Get node pool information
log_info "Node Pools:"
echo ""
gcloud container node-pools list --cluster="$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" \
    --format="table(name,status,config.machineType,autoscaling.enabled,initialNodeCount)"
echo ""

# Get node counts
STATEFUL_COUNT=$(gcloud container node-pools describe stateful-pool \
    --cluster="$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" \
    --format="value(initialNodeCount)" 2>/dev/null || echo "0")

STATELESS_COUNT=$(gcloud container node-pools describe stateless-pool \
    --cluster="$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" \
    --format="value(initialNodeCount)" 2>/dev/null || echo "0")

TOTAL_NODES=$((STATEFUL_COUNT + STATELESS_COUNT))

log_info "Node Counts:"
echo "  stateful-pool:  $STATEFUL_COUNT nodes"
echo "  stateless-pool: $STATELESS_COUNT nodes"
echo "  Total:          $TOTAL_NODES nodes"
echo ""

# Determine cluster state
if [[ $TOTAL_NODES -eq 0 ]]; then
    log_warn "Cluster is PAUSED (all node pools scaled to 0)"
    echo ""
    log_info "Persistent volumes and data are preserved"
    log_info "To resume: ./resume-cluster.sh"
elif [[ $TOTAL_NODES -ge 5 ]]; then
    log_success "Cluster is RUNNING (normal operation)"
    echo ""
    log_info "To pause: ./pause-cluster.sh"
else
    log_warn "Cluster is PARTIALLY RUNNING ($TOTAL_NODES/5 nodes)"
    echo ""
    log_info "Expected: 5 nodes (3 stateful + 2 stateless)"
fi

# If kubectl is available and cluster has nodes, show node status
if command -v kubectl &> /dev/null && [[ $TOTAL_NODES -gt 0 ]]; then
    echo ""
    log_info "Fetching cluster credentials..."

    if gcloud container clusters get-credentials "$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" &> /dev/null; then
        echo ""
        log_info "Node Status:"
        kubectl get nodes -o wide 2>/dev/null || log_warn "Could not fetch node status"
        echo ""

        log_info "Pods by Namespace:"
        kubectl get pods --all-namespaces --field-selector=status.phase=Running 2>/dev/null | \
            awk 'NR>1 {count[$1]++} END {for (ns in count) printf "  %-30s %d pods\n", ns, count[ns]}' || \
            log_warn "Could not fetch pod status"
    fi
fi
