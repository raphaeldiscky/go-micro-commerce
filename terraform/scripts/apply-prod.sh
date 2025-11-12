#!/bin/bash
# Apply Terraform changes for production environment
# This script applies the Terraform plan to create/update infrastructure

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
    exit 1
fi

cd "$ENV_DIR"

log_info "Applying Terraform changes for: $ENV"

# Check if already initialized
if [[ ! -d ".terraform" ]]; then
    log_warn "Terraform not initialized. Run init-backend.sh first"
    exit 1
fi

# Get GKE credentials (if cluster exists)
CLUSTER_NAME=$(grep -E '^cluster_name\s*=' terraform.tfvars | cut -d'"' -f2 || echo "")
ZONE=$(grep -E '^zone\s*=' terraform.tfvars | cut -d'"' -f2 || echo "")
PROJECT_ID=$(grep -E '^project_id\s*=' terraform.tfvars | cut -d'"' -f2 || echo "")

if [[ -n "$CLUSTER_NAME" ]] && [[ -n "$ZONE" ]] && [[ -n "$PROJECT_ID" ]]; then
    if gcloud container clusters describe "$CLUSTER_NAME" --zone="$ZONE" --project="$PROJECT_ID" &> /dev/null; then
        log_info "Getting GKE credentials for existing cluster"
        gcloud container clusters get-credentials "$CLUSTER_NAME" --zone="$ZONE" --project="$PROJECT_ID"
    fi
fi

# Apply terraform changes
if [[ -f "tfplan" ]]; then
    log_info "Applying saved plan: tfplan"
    terraform apply tfplan
    rm -f tfplan
else
    log_warn "No saved plan found. Running apply with auto-approve"
    log_warn "Press Ctrl+C within 5 seconds to cancel..."
    sleep 5
    terraform apply -auto-approve
fi

log_success "Infrastructure deployed successfully!"
echo ""
log_info "Getting cluster credentials..."

# Get fresh credentials after cluster creation
if gcloud container clusters describe "$CLUSTER_NAME" --zone="$ZONE" --project="$PROJECT_ID" &> /dev/null; then
    gcloud container clusters get-credentials "$CLUSTER_NAME" --zone="$ZONE" --project="$PROJECT_ID"

    log_success "kubectl configured successfully"
    echo ""
    log_info "Cluster information:"
    kubectl cluster-info
    echo ""
    log_info "Node pools:"
    kubectl get nodes -o wide
    echo ""
    log_info "Namespaces:"
    kubectl get namespaces
    echo ""

    log_info "To view Grafana dashboard:"
    log_info "  kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80"
    log_info "  Then visit: http://localhost:3000"
    echo ""
    log_info "To view ArgoCD UI:"
    log_info "  kubectl port-forward -n argocd svc/argocd-server 8080:443"
    log_info "  Then visit: https://localhost:8080"
    echo ""
    log_info "To view Traefik dashboard:"
    log_info "  kubectl port-forward -n traefik svc/traefik 9000:9000"
    log_info "  Then visit: http://localhost:9000/dashboard/"
else
    log_error "Failed to get cluster credentials"
    exit 1
fi
