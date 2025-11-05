#!/bin/bash
# Stop Tilt and optionally clean up resources

set -euo pipefail

echo "Stopping Tilt for go-micro-commerce"
echo ""

# Check if Tilt is installed
if ! command -v tilt &> /dev/null; then
    echo "Tilt is not installed"
    exit 1
fi

# Stop Tilt
echo "Stopping Tilt..."
tilt down

echo ""
echo "Tilt stopped successfully"
echo ""

# Ask if user wants to clean up resources
read -p "Do you want to delete all Kubernetes resources? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Cleaning up Kubernetes resources..."

    # Delete application resources
    kubectl delete all --all -n default --timeout=60s || true

    # Delete PVCs (persistent volumes)
    echo "Deleting persistent volumes..."
    kubectl delete pvc --all -n default --timeout=60s || true

    # Delete ConfigMaps and Secrets (excluding kube-system)
    kubectl delete configmap --all -n default --timeout=60s || true
    kubectl delete secret --all -n default --timeout=60s || true

    echo "Resources cleaned up"
else
    echo "Resources kept (PVCs, ConfigMaps, Secrets remain)"
fi

echo ""
echo "To delete the Kind cluster completely, run:"
echo "   kind delete cluster --name go-micro-commerce"
echo "To delete the MicroK8s cluster completely, run:"
echo "   sudo microk8s reset"
echo ""
