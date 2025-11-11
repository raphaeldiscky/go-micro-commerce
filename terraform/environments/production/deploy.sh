#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
print_info "Checking prerequisites..."

if ! command_exists terraform; then
    print_error "Terraform is not installed. Please install Terraform >= 1.13.5"
    exit 1
fi

if ! command_exists talosctl; then
    print_error "talosctl is not installed. Please install talosctl"
    print_info "Install with: curl -sL https://talos.dev/install | sh"
    exit 1
fi

# Check Terraform version
TERRAFORM_VERSION=$(terraform version -json | grep -o '"terraform_version":"[^"]*"' | cut -d'"' -f4)
print_success "Terraform version: $TERRAFORM_VERSION"

# Check talosctl version
TALOSCTL_VERSION=$(talosctl version --client --short)
print_success "talosctl version: $TALOSCTL_VERSION"

# Check if terraform.tfvars exists
if [ ! -f "terraform.tfvars" ]; then
    print_error "terraform.tfvars not found!"
    print_info "Please create terraform.tfvars from terraform.tfvars.example"
    print_info "cp terraform.tfvars.example terraform.tfvars"
    print_info "Then edit terraform.tfvars with your actual values"
    exit 1
fi

print_success "All prerequisites met"
echo ""

# Initialize Terraform
print_info "Initializing Terraform..."
terraform init
print_success "Terraform initialized"
echo ""

# Validate configuration
print_info "Validating Terraform configuration..."
terraform validate
print_success "Configuration is valid"
echo ""

# Show plan for Phase 1
print_info "=== PHASE 1: Deploy Talos Cluster ==="
print_info "This will create the Talos cluster and generate kubeconfig"
echo ""

read -p "Do you want to see the plan first? (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    terraform plan -target=module.talos_cluster
    echo ""
    read -p "Proceed with deployment? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Deployment cancelled by user"
        exit 0
    fi
fi

# Phase 1: Deploy Talos cluster
print_info "Deploying Talos cluster..."
terraform apply -target=module.talos_cluster -auto-approve

if [ $? -ne 0 ]; then
    print_error "Phase 1 failed. Please check the errors above."
    exit 1
fi

print_success "Phase 1 completed successfully"
echo ""

# Check if kubeconfig was created
if [ ! -f "./kubeconfig-production" ]; then
    print_error "Kubeconfig file not found at ./kubeconfig-production"
    print_error "Cluster may not be ready yet"
    exit 1
fi

print_success "Kubeconfig generated at ./kubeconfig-production"
echo ""

# Wait for Kubernetes API to be ready
print_info "Waiting for Kubernetes API to be ready..."
KUBECONFIG=./kubeconfig-production kubectl cluster-info > /dev/null 2>&1 || {
    print_warning "Kubernetes API not ready yet. Waiting 30 seconds..."
    sleep 30
    KUBECONFIG=./kubeconfig-production kubectl cluster-info > /dev/null 2>&1 || {
        print_error "Kubernetes API is not responding. Please check cluster status with talosctl"
        exit 1
    }
}

print_success "Kubernetes API is ready"
echo ""

# Phase 2: Deploy Kubernetes resources
print_info "=== PHASE 2: Deploy Kubernetes Resources ==="
print_info "This will deploy:"
print_info "  - Cilium CNI"
print_info "  - Longhorn storage"
print_info "  - CloudNativePG operator"
print_info "  - External Secrets Operator (if enabled)"
echo ""

read -p "Proceed with Phase 2? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Phase 2 skipped by user"
    print_info "You can run Phase 2 later with: terraform apply"
    exit 0
fi

print_info "Deploying Kubernetes resources..."
terraform apply -auto-approve

if [ $? -ne 0 ]; then
    print_error "Phase 2 failed. Some resources may be partially deployed."
    print_warning "You can retry with: terraform apply"
    exit 1
fi

print_success "Phase 2 completed successfully"
echo ""

# Final verification
print_info "=== Deployment Verification ==="
export KUBECONFIG=./kubeconfig-production

print_info "Cluster nodes:"
kubectl get nodes -o wide

echo ""
print_info "Deployed namespaces:"
kubectl get namespaces

echo ""
print_info "Longhorn status:"
kubectl get pods -n longhorn-system 2>/dev/null || print_warning "Longhorn not yet ready"

echo ""
print_info "CNI status (Cilium):"
kubectl get pods -n kube-system -l k8s-app=cilium 2>/dev/null || print_warning "Cilium not yet ready"

echo ""
print_success "=== Deployment Complete ==="
echo ""
print_info "Next steps:"
echo ""
print_info "1. Save the Talos configuration (only needs to be done once):"
echo "   export TALOSCONFIG=\$(pwd)/talosconfig-production"
echo "   talosctl config merge \$TALOSCONFIG"
echo ""
print_info "2. Export kubeconfig to your environment:"
echo "   export KUBECONFIG=\$(pwd)/kubeconfig-production"
echo "   # Or merge into default kubeconfig:"
echo "   KUBECONFIG=~/.kube/config:\$(pwd)/kubeconfig-production kubectl config view --flatten > ~/.kube/config.new"
echo "   mv ~/.kube/config.new ~/.kube/config"
echo ""
print_info "3. Verify Talos cluster health:"
echo "   talosctl health --nodes 10.184.0.8,10.184.0.9,10.184.0.10"
echo ""
print_info "4. Monitor CNI deployment (may take a few minutes):"
echo "   kubectl get pods -n kube-system -w"
echo ""
print_info "5. Check Longhorn storage:"
echo "   kubectl get pods -n longhorn-system"
echo "   # Access Longhorn UI (if using port-forward):"
echo "   kubectl port-forward -n longhorn-system svc/longhorn-frontend 8080:80"
echo ""
print_info "Configuration files generated:"
echo "   - ./kubeconfig-production (Kubernetes access)"
echo "   - ./talosconfig-production (Talos management)"
echo ""
print_warning "IMPORTANT: Do not commit these files to version control!"
