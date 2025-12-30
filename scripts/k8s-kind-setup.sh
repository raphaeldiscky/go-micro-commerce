#!/bin/bash
# Setup Kind cluster for local development with Tilt
# Creates a Kind cluster with local registry

set -euo pipefail

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

CLUSTER_NAME="${CLUSTER_NAME:-go-micro-commerce}"
REGISTRY_NAME="${REGISTRY_NAME:-kind-registry}"
REGISTRY_PORT="${REGISTRY_PORT:-5000}"

echo -e "${CYAN}Setting up Kind cluster for go-micro-commerce${NC}"
echo ""

# Check if Kind is installed
if ! command -v kind &> /dev/null; then
    echo -e "${RED}Kind is not installed. Please install it first:${NC}"
    echo "   https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl is not installed. Please install it first:${NC}"
    echo "     https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi

# Check if cluster already exists
if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo -e "${YELLOW}Kind cluster '${CLUSTER_NAME}' already exists${NC}"
    read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Deleting existing cluster...${NC}"
        kind delete cluster --name "${CLUSTER_NAME}"
    else
        echo -e "${GREEN}Using existing cluster${NC}"
        exit 0
    fi
fi

# Create local registry if it doesn't exist
if [ "$(docker ps -q -f name=${REGISTRY_NAME})" ]; then
    echo -e "${GREEN}Registry '${REGISTRY_NAME}' already exists${NC}"
else
    echo -e "${BLUE}Creating local registry...${NC}"
    docker run -d \
        --restart=always \
        -p "${REGISTRY_PORT}:5000" \
        --name "${REGISTRY_NAME}" \
        registry:2
fi

# Create Kind cluster with registry
echo -e "${BLUE}Creating Kind cluster '${CLUSTER_NAME}'...${NC}"
cat <<EOF | kind create cluster --name "${CLUSTER_NAME}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: infra
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${REGISTRY_PORT}"]
      endpoint = ["http://${REGISTRY_NAME}:5000"]
EOF

# Connect registry to cluster network
echo -e "${BLUE}Connecting registry to cluster network...${NC}"
docker network connect "kind" "${REGISTRY_NAME}" 2>/dev/null || true

# Document the local registry
echo -e "${BLUE}Documenting local registry in cluster...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${REGISTRY_PORT}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

echo ""
echo -e "${GREEN}Kind cluster '${CLUSTER_NAME}' created successfully!${NC}"
echo ""
echo -e "${CYAN}Cluster info:${NC}"
kubectl cluster-info --context "kind-${CLUSTER_NAME}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Run 'task k8s_create_secrets' to generate secrets"
echo "  2. Run 'task tilt_up' to start all services"
echo ""