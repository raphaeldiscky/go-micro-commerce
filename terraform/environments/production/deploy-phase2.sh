#!/bin/bash
set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source common functions
source "${SCRIPT_DIR}/common.sh"

print_info "=== PHASE 2: Kubernetes Resources Deployment ==="
echo ""

# Check if Phase 1 was completed
if [ ! -f ".phase1-complete" ]; then
    print_error "Phase 1 not completed yet!"
    print_info "Please run ./deploy-phase1.sh first to:"
    print_info "  - Deploy Talos cluster"
    print_info "  - Generate kubeconfig and talosconfig"
    print_info "  - Install CNI (Cilium)"
    exit 1
fi

print_info "Checking prerequisites..."

# Check if required commands exist
if ! command_exists kubectl; then
    print_error "kubectl is not installed. Please install kubectl"
    exit 1
fi

if ! command_exists terraform; then
    print_error "Terraform is not installed. Please install Terraform"
    exit 1
fi

# Check if config files exist
if [ ! -f "./kubeconfig-production" ]; then
    print_error "kubeconfig-production not found!"
    print_info "Phase 1 may not have completed successfully"
    exit 1
fi

if [ ! -f "./talosconfig-production" ]; then
    print_error "talosconfig-production not found!"
    print_info "Phase 1 may not have completed successfully"
    exit 1
fi

print_success "Prerequisites met"
echo ""

# Verify Kubernetes API is ready
print_info "Verifying Kubernetes API is ready..."
export KUBECONFIG=./kubeconfig-production

if ! kubectl cluster-info > /dev/null 2>&1; then
    print_error "Kubernetes API is not responding!"
    print_info "Please check:"
    print_info "  1. Cluster status: kubectl get nodes"
    print_info "  2. Cilium pods: kubectl get pods -n kube-system -l k8s-app=cilium"
    print_info "  3. Talos health: export TALOSCONFIG=./talosconfig-production && talosctl health"
    exit 1
fi

print_success "Kubernetes API is ready"
echo ""

# Check node status
print_info "Current cluster status:"
kubectl get nodes -o wide
echo ""

# Check CNI status
print_info "Cilium CNI status:"
kubectl get pods -n kube-system -l k8s-app=cilium
echo ""

# Phase 2: Deploy Kubernetes resources
print_info "=== Deploying Kubernetes Resources ==="
print_info "This will deploy:"
print_info "  - Longhorn storage"
print_info "  - CloudNativePG operator"
print_info "  - External Secrets Operator (if enabled)"
echo ""

read -p "Proceed with Phase 2 deployment? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Phase 2 skipped by user"
    print_info "You can run Phase 2 later with:"
    print_info "  ./deploy-phase2.sh"
    print_info "Or manually with:"
    print_info "  terraform apply"
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
print_success "=== DEPLOYMENT COMPLETE ==="
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
echo "   talosctl health --nodes <node-ips>"
echo ""
print_info "4. Monitor CNI deployment (may take a few minutes):"
echo "   kubectl get pods -n kube-system -w"
echo ""
print_info "5. Check Longhorn storage:"
echo "   kubectl get pods -n longhorn-system"
echo "   # Access Longhorn UI (if using port-forward):"
echo "   kubectl port-forward -n longhorn-system svc/longhorn-frontend 8080:80"
echo ""
print_info "Configuration files:"
echo "   - ./kubeconfig-production (Kubernetes access)"
echo "   - ./talosconfig-production (Talos management)"
echo ""
print_warning "IMPORTANT: Do not commit these files to version control!"
