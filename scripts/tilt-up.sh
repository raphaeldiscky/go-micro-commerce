#!/bin/bash
# Start Tilt with pre-flight checks
# Ensures cluster, secrets, and dependencies are ready

set -euo pipefail

echo "Starting Tilt for go-micro-commerce"
echo ""

# Check if Tilt is installed
if ! command -v tilt &> /dev/null; then
    echo "Tilt is not installed. Please install it first:"
    echo "   https://docs.tilt.dev/install.html"
    exit 1
fi

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "kubectl is not installed"
    exit 1
fi

# Check if cluster is running
if ! kubectl cluster-info &> /dev/null; then
    echo "Kubernetes cluster is not running"
    echo "   Run './scripts/kind-setup.sh' to create a Kind cluster"
    exit 1
fi

echo "Kubernetes cluster is running"

# Get cluster context
CONTEXT=$(kubectl config current-context)
echo "Current context: ${CONTEXT}"

# Check if secrets exist
SECRETS_DIR="deployments/k8s/overlays/local/secrets"
if [ ! -f "${SECRETS_DIR}/auth-service-private.pem" ]; then
    echo "JWT secrets not found"
    echo "   Run './scripts/create-secrets.sh' to generate secrets"
    exit 1
fi

echo "Secrets are configured"
echo ""

# Check if Helm is installed (optional but recommended)
if ! command -v helm &> /dev/null; then
    echo "Helm is not installed (recommended for infrastructure deployment)"
    echo "   Install from: https://helm.sh/docs/intro/install/"
    echo ""
fi

# Start Tilt
echo "Starting Tilt..."
echo "   Press Ctrl+C to stop"
echo "   Tilt UI will open at: http://localhost:10350"
echo ""

tilt up

