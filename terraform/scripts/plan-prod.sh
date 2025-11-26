#!/bin/bash
# Plan Terraform changes for production environment
# This script runs terraform plan to preview infrastructure changes

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
    log_info "Copy terraform.tfvars.example to terraform.tfvars and customize it"
    log_info "  cp $ENV_DIR/terraform.tfvars.example $ENV_DIR/terraform.tfvars"
    exit 1
fi

cd "$ENV_DIR"

log_info "Planning Terraform changes for: $ENV"

# Check if already initialized
if [[ ! -d ".terraform" ]]; then
    log_warn "Terraform not initialized. Run init-backend.sh first"
    log_info "Running: terraform init"
    terraform init -backend-config=backend.hcl
fi

# Get GKE credentials (if cluster exists)
CLUSTER_NAME=$(grep -E '^cluster_name\s*=' terraform.tfvars | cut -d'"' -f2 || echo "")
REGION=$(grep -E '^region\s*=' "$ENV_DIR/terraform.tfvars" | cut -d'"' -f2 || echo "")
PROJECT_ID=$(grep -E '^project_id\s*=' terraform.tfvars | cut -d'"' -f2 || echo "")

if [[ -n "$CLUSTER_NAME" ]] && [[ -n "$REGION" ]] && [[ -n "$PROJECT_ID" ]]; then
    if gcloud container clusters describe "$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" &> /dev/null; then
        log_info "Getting GKE credentials for existing cluster"
        gcloud container clusters get-credentials "$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID"
    else
        log_warn "GKE cluster not found (this is expected for first deployment)"
    fi
fi

# Run terraform plan
log_info "Running: terraform plan"
terraform plan -out=terraform.tfplan

log_info "Plan saved to: terraform.tfplan"
log_info "To apply this plan, run: ./terraform/scripts/apply-prod.sh"
