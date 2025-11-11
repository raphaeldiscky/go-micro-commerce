#!/bin/bash
set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source common functions
source "${SCRIPT_DIR}/common.sh"

print_info "=== Talos Kubernetes Cluster Deployment ==="
print_info "This script will deploy the entire cluster in two phases:"
echo ""
print_info "Phase 1: Talos Cluster + CNI (Cilium)"
print_info "  - Deploy Talos cluster on VMs"
print_info "  - Generate kubeconfig and talosconfig"
print_info "  - Install Cilium CNI for pod networking"
echo ""
print_info "Phase 2: Kubernetes Resources"
print_info "  - Deploy Longhorn storage"
print_info "  - Deploy CloudNativePG operator"
print_info "  - Deploy External Secrets Operator (if enabled)"
echo ""
print_info "You can also run each phase separately:"
print_info "  ./deploy-phase1.sh  - Deploy Talos cluster + CNI"
print_info "  ./deploy-phase2.sh  - Deploy Kubernetes resources"
echo ""

read -p "Proceed with full deployment? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Deployment cancelled by user"
    exit 0
fi

echo ""
echo "========================================================================"
echo ""

# Run Phase 1
print_info "Starting Phase 1..."
echo ""

if ! bash "${SCRIPT_DIR}/deploy-phase1.sh"; then
    print_error "Phase 1 failed!"
    print_info "Fix the errors and try again, or run Phase 1 manually:"
    print_info "  ./deploy-phase1.sh"
    exit 1
fi

echo ""
echo "========================================================================"
echo ""

# Run Phase 2
print_info "Starting Phase 2..."
echo ""

if ! bash "${SCRIPT_DIR}/deploy-phase2.sh"; then
    print_error "Phase 2 failed!"
    print_info "Phase 1 completed successfully, but Phase 2 failed."
    print_info "You can retry Phase 2 with:"
    print_info "  ./deploy-phase2.sh"
    print_info "Or manually with:"
    print_info "  terraform apply"
    exit 1
fi

echo ""
echo "========================================================================"
echo ""
print_success "=== FULL DEPLOYMENT COMPLETE ==="
echo ""
print_info "Both phases completed successfully!"
print_info "Your Talos Kubernetes cluster is ready to use."
