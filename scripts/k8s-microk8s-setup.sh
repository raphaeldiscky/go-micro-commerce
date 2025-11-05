#!/bin/bash
# Setup MicroK8s for local development with Tilt
# Enables required addons and configures kubectl context

set -euo pipefail

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}Setting up MicroK8s for go-micro-commerce${NC}"
echo ""

# Check if MicroK8s is installed
if ! command -v microk8s &> /dev/null; then
    echo -e "${RED}MicroK8s is not installed. Installing now...${NC}"
    echo ""
    echo -e "${BLUE}Running: sudo snap install microk8s --classic${NC}"
    sudo snap install microk8s --classic

    echo -e "${BLUE}Adding current user to microk8s group...${NC}"
    sudo usermod -a -G microk8s $USER
    sudo chown -f -R $USER ~/.kube

    echo -e "${YELLOW}NOTE: You may need to log out and back in for group changes to take effect${NC}"
    echo ""
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${YELLOW}kubectl is not installed. Installing via snap...${NC}"
    sudo snap install kubectl --classic
fi

# Wait for MicroK8s to be ready
echo -e "${BLUE}Waiting for MicroK8s to be ready...${NC}"
sudo microk8s status --wait-ready

# Enable required addons
echo -e "${BLUE}Enabling required MicroK8s addons...${NC}"
echo ""

echo -e "${CYAN}  → Enabling DNS addon...${NC}"
sudo microk8s enable dns

echo -e "${CYAN}  → Enabling registry addon (localhost:32000)...${NC}"
sudo microk8s enable registry

echo -e "${CYAN}  → Enabling storage addon...${NC}"
sudo microk8s enable storage

echo -e "${CYAN}  → Enabling hostpath-storage addon...${NC}"
sudo microk8s enable hostpath-storage

echo ""
echo -e "${BLUE}Configuring kubectl context...${NC}"

# Create or update kubeconfig
mkdir -p ~/.kube
sudo microk8s kubectl config view --flatten > ~/.kube/microk8s-config

# Merge with existing config if it exists
if [ -f ~/.kube/config ]; then
    KUBECONFIG=~/.kube/microk8s-config:~/.kube/config kubectl config view --flatten > ~/.kube/temp-config
    mv ~/.kube/temp-config ~/.kube/config
else
    cp ~/.kube/microk8s-config ~/.kube/config
fi

# Switch to microk8s context
kubectl config use-context microk8s

echo ""
echo -e "${GREEN}MicroK8s setup complete!${NC}"
echo ""
echo -e "${CYAN}Cluster info:${NC}"
kubectl cluster-info
echo ""
echo -e "${CYAN}Enabled addons:${NC}"
sudo microk8s status
echo ""
echo -e "${BLUE}Registry available at: ${GREEN}localhost:32000${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Run 'task k8s_create_secrets' to generate secrets"
echo "  2. Run 'task tilt_up' to start all services"
echo ""
echo -e "${YELLOW}Useful commands:${NC}"
echo "  - Check status: sudo microk8s status"
echo "  - View addons: sudo microk8s status --addon"
echo "  - Stop MicroK8s: sudo microk8s stop"
echo "  - Start MicroK8s: sudo microk8s start"
echo "  - Reset cluster: sudo microk8s reset (WARNING: destroys all data)"
echo ""
