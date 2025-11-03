#!/bin/bash
# Create Kubernetes secrets for JWT keys and other sensitive data
# Generates RSA key pairs if they don't exist

set -euo pipefail

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}Creating Kubernetes secrets for go-micro-commerce${NC}"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl is not installed${NC}"
    exit 1
fi

# Check if openssl is available
if ! command -v openssl &> /dev/null; then
    echo -e "${RED}openssl is not installed${NC}"
    exit 1
fi

# Create secrets directory if it doesn't exist
SECRETS_DIR="deployments/k8s/overlays/local/secrets"
mkdir -p "${SECRETS_DIR}"

echo -e "${BLUE}Secrets directory: ${SECRETS_DIR}${NC}"
echo ""

# Generate JWT keys for auth-service if they don't exist
if [ ! -f "${SECRETS_DIR}/auth-service-private.pem" ] || [ ! -f "${SECRETS_DIR}/auth-service-public.pem" ]; then
    echo -e "${YELLOW}Generating RSA key pair for auth-service...${NC}"
    openssl genrsa -out "${SECRETS_DIR}/auth-service-private.pem" 2048
    openssl rsa -in "${SECRETS_DIR}/auth-service-private.pem" -pubout -out "${SECRETS_DIR}/auth-service-public.pem"
    echo -e "${GREEN}auth-service keys generated${NC}"
else
    echo -e "${GREEN}auth-service keys already exist${NC}"
fi

# Generate JWT public key for api-gateway (copy from auth-service)
if [ ! -f "${SECRETS_DIR}/api-gateway-public.pem" ]; then
    echo -e "${YELLOW}Copying public key for api-gateway...${NC}"
    cp "${SECRETS_DIR}/auth-service-public.pem" "${SECRETS_DIR}/api-gateway-public.pem"
    echo -e "${GREEN}api-gateway public key created${NC}"
else
    echo -e "${GREEN}api-gateway public key already exists${NC}"
fi

# Generate JWT public key for chat-service (copy from auth-service)
if [ ! -f "${SECRETS_DIR}/chat-service-public.pem" ]; then
    echo -e "${YELLOW}Copying public key for chat-service...${NC}"
    cp "${SECRETS_DIR}/auth-service-public.pem" "${SECRETS_DIR}/chat-service-public.pem"
    echo -e "${GREEN}chat-service public key created${NC}"
else
    echo -e "${GREEN}chat-service public key already exists${NC}"
fi

echo ""
echo -e "${CYAN}Created secrets:${NC}"
ls -lh "${SECRETS_DIR}/"

echo ""
echo -e "${GREEN}All secrets created successfully!${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Run 'task tilt_up' to start all services"
echo "  2. Secrets will be automatically mounted to pods via Kustomize"
echo ""