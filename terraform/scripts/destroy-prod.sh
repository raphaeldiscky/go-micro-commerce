#!/bin/bash
# Destroy Terraform-managed production infrastructure
# WARNING: This will delete all infrastructure managed by Terraform!

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="${SCRIPT_DIR}/.."
ENV="prod"
ENV_DIR="${TERRAFORM_DIR}/environments/${ENV}"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Check if terraform is installed
if ! command -v terraform &> /dev/null; then
    log_error "Terraform is not installed. Please install it first."
    exit 1
fi

# Check if environment directory exists
if [[ ! -d "$ENV_DIR" ]]; then
    log_error "Environment directory not found: $ENV_DIR"
    exit 1
fi

cd "$ENV_DIR"

# Safety check
echo ""
log_error "=========================================="
log_error "  DANGER: INFRASTRUCTURE DESTRUCTION"
log_error "=========================================="
echo ""
log_warn "This will DESTROY all production infrastructure including:"
log_warn "  - GKE Cluster and all workloads"
log_warn "  - VPC Network and subnets"
log_warn "  - All deployed operators (PostgreSQL, Kafka, Redis)"
log_warn "  - Monitoring stack (Prometheus, Grafana, Loki, Tempo)"
log_warn "  - ArgoCD and Traefik"
echo ""
log_warn "Data in persistent volumes may be deleted!"
echo ""

# Get cluster name from terraform.tfvars
CLUSTER_NAME=$(grep -E '^cluster_name\s*=' terraform.tfvars | cut -d'"' -f2 || echo "unknown")

echo ""
log_error "You are about to destroy: $CLUSTER_NAME ($ENV)"
echo ""
read -p "Type 'yes' to confirm destruction: " -r
echo

if [[ ! $REPLY =~ ^yes$ ]]; then
    log_info "Destruction cancelled"
    exit 0
fi

# Second confirmation
echo ""
log_error "FINAL WARNING: This action cannot be undone!"
read -p "Type the cluster name '$CLUSTER_NAME' to confirm: " -r
echo

if [[ "$REPLY" != "$CLUSTER_NAME" ]]; then
    log_info "Destruction cancelled (name mismatch)"
    exit 0
fi

log_info "Proceeding with destruction..."
sleep 3

# Check if already initialized
if [[ ! -d ".terraform" ]]; then
    log_warn "Terraform not initialized. Initializing..."
    terraform init -backend-config=backend.hcl
fi

# Run terraform destroy
log_warn "Running: terraform destroy"
terraform destroy

log_info "Infrastructure destroyed successfully"
log_warn "Note: The GCS state bucket was NOT deleted. To delete it manually:"
log_warn "  gsutil rm -r gs://go-micro-commerce-terraform-state-${ENV}"
