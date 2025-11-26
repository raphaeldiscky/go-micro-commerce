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
REGION=$(grep -E '^region\s*=' "$ENV_DIR/terraform.tfvars" | cut -d'"' -f2 || echo "")
PROJECT_ID=$(grep -E '^project_id\s*=' terraform.tfvars | cut -d'"' -f2 || echo "")

if [[ -n "$CLUSTER_NAME" ]] && [[ -n "$REGION" ]] && [[ -n "$PROJECT_ID" ]]; then
    if gcloud container clusters describe "$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" &> /dev/null; then
        log_info "Getting GKE credentials for existing cluster"
        gcloud container clusters get-credentials "$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID"
    fi
fi

# Apply terraform changes
if [[ -f "terraform.tfplan" ]]; then
    log_info "Applying saved plan: terraform.tfplan"
    terraform apply terraform.tfplan
    rm -f terraform.tfplan
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
if gcloud container clusters describe "$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID" &> /dev/null; then
    gcloud container clusters get-credentials "$CLUSTER_NAME" --region="$REGION" --project="$PROJECT_ID"

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

    log_success "=== Access Your Services ==="
    echo ""

    # Get outputs from terraform
    ARGOCD_URL=$(terraform output -raw argocd_public_url 2>/dev/null || echo "")
    GRAFANA_URL=$(terraform output -raw grafana_public_url 2>/dev/null || echo "")

    if [[ -n "$ARGOCD_URL" ]]; then
        log_success "ArgoCD (HTTPS - Production):"
        log_info "  URL: $ARGOCD_URL"
        log_info "  Username: admin"
        log_info "  Get Password: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d"
        echo ""
    fi

    if [[ -n "$GRAFANA_URL" ]]; then
        log_success "Grafana (HTTPS - Production):"
        log_info "  URL: $GRAFANA_URL"
        log_info "  Username: admin"
        log_info "  Password: <from your terraform.tfvars grafana_admin_password>"
        echo ""
    fi

    log_info "Alternative Access (Port-forward):"
    log_info "  Grafana:  kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80"
    log_info "            http://localhost:3000"
    log_info "  ArgoCD:   kubectl port-forward -n argocd svc/argocd-server 8080:80"
    log_info "            http://localhost:8080"
    log_info "  Traefik:  kubectl port-forward -n traefik svc/traefik 9000:9000"
    log_info "            http://localhost:9000/dashboard/"
    echo ""

    log_info "TLS Certificate Status:"
    log_info "  kubectl get certificates -A"
    echo ""
else
    log_error "Failed to get cluster credentials"
    exit 1
fi
