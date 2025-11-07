#!/bin/bash
# Generate self-signed TLS certificates for local Kubernetes development
# Creates K8s secret for Traefik Ingress

set -euo pipefail

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

CERT_DIR="deployments/k8s/overlays/local/secrets/tls"
DOMAIN="${DOMAIN:-go.micro.commerce}"
SECRET_NAME="${SECRET_NAME:-go-micro-commerce-tls}"
NAMESPACE="${NAMESPACE:-default}"

echo -e "${CYAN}Generating TLS certificates for ${DOMAIN}${NC}"
echo ""

# Create cert directory if it doesn't exist
mkdir -p "${CERT_DIR}"

# Check if certificates already exist
if [ -f "${CERT_DIR}/${DOMAIN}.crt" ] && [ -f "${CERT_DIR}/${DOMAIN}.key" ]; then
    echo -e "${YELLOW}Certificates already exist at ${CERT_DIR}${NC}"
    read -p "Do you want to regenerate them? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${GREEN}Using existing certificates${NC}"

        # Check if secret exists
        if kubectl get secret "${SECRET_NAME}" -n "${NAMESPACE}" &> /dev/null; then
            echo -e "${GREEN}Secret '${SECRET_NAME}' already exists in namespace '${NAMESPACE}'${NC}"
            exit 0
        else
            echo -e "${BLUE}Creating K8s secret from existing certificates...${NC}"
            kubectl create secret tls "${SECRET_NAME}" \
                --cert="${CERT_DIR}/${DOMAIN}.crt" \
                --key="${CERT_DIR}/${DOMAIN}.key" \
                -n "${NAMESPACE}"
            echo -e "${GREEN}Secret '${SECRET_NAME}' created successfully${NC}"
            exit 0
        fi
    fi
fi

# Generate private key
echo -e "${BLUE}Generating private key...${NC}"
openssl genrsa -out "${CERT_DIR}/${DOMAIN}.key" 2048

# Generate certificate signing request
echo -e "${BLUE}Generating certificate signing request...${NC}"
openssl req -new -key "${CERT_DIR}/${DOMAIN}.key" \
    -out "${CERT_DIR}/${DOMAIN}.csr" \
    -subj "/C=US/ST=Local/L=Local/O=Development/OU=Engineering/CN=${DOMAIN}"

# Generate self-signed certificate (valid for 1 year)
echo -e "${BLUE}Generating self-signed certificate (valid for 365 days)...${NC}"
openssl x509 -req -days 365 \
    -in "${CERT_DIR}/${DOMAIN}.csr" \
    -signkey "${CERT_DIR}/${DOMAIN}.key" \
    -out "${CERT_DIR}/${DOMAIN}.crt" \
    -extfile <(printf "subjectAltName=DNS:${DOMAIN},DNS:*.${DOMAIN}")

# Clean up CSR
rm "${CERT_DIR}/${DOMAIN}.csr"

echo ""
echo -e "${GREEN}Certificates generated successfully:${NC}"
echo -e "  Certificate: ${CERT_DIR}/${DOMAIN}.crt"
echo -e "  Private Key: ${CERT_DIR}/${DOMAIN}.key"
echo ""

# Create or update Kubernetes secret
echo -e "${BLUE}Creating Kubernetes secret '${SECRET_NAME}' in namespace '${NAMESPACE}'...${NC}"

# Delete secret if it exists
if kubectl get secret "${SECRET_NAME}" -n "${NAMESPACE}" &> /dev/null; then
    echo -e "${YELLOW}Secret '${SECRET_NAME}' already exists, deleting...${NC}"
    kubectl delete secret "${SECRET_NAME}" -n "${NAMESPACE}"
fi

# Create new secret
kubectl create secret tls "${SECRET_NAME}" \
    --cert="${CERT_DIR}/${DOMAIN}.crt" \
    --key="${CERT_DIR}/${DOMAIN}.key" \
    -n "${NAMESPACE}"

echo ""
echo -e "${GREEN}TLS secret '${SECRET_NAME}' created successfully!${NC}"
echo ""
echo -e "${CYAN}Certificate details:${NC}"
openssl x509 -in "${CERT_DIR}/${DOMAIN}.crt" -noout -text | grep -A2 "Subject:"
openssl x509 -in "${CERT_DIR}/${DOMAIN}.crt" -noout -text | grep -A1 "Validity"
openssl x509 -in "${CERT_DIR}/${DOMAIN}.crt" -noout -text | grep "DNS:"
echo ""
echo -e "${YELLOW}Note: This is a self-signed certificate for local development only.${NC}"
echo -e "${YELLOW}Your browser will show a security warning when accessing https://${DOMAIN}${NC}"
echo ""
