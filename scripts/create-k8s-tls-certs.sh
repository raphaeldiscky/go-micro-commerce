#!/bin/bash
# Generate or use existing TLS certificates for local Kubernetes development
# Prefers mkcert certificates if available, falls back to self-signed
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
MKCERT_CERT="frontend/${DOMAIN}.pem"
MKCERT_KEY="frontend/${DOMAIN}-key.pem"

echo -e "${CYAN}Setting up TLS certificates for ${DOMAIN}${NC}"
echo ""

# Create cert directory if it doesn't exist
mkdir -p "${CERT_DIR}"

# Check for existing mkcert certificates (preferred)
if [ -f "${MKCERT_CERT}" ] && [ -f "${MKCERT_KEY}" ]; then
    echo -e "${GREEN}Found existing mkcert certificates!${NC}"
    echo -e "  Certificate: ${MKCERT_CERT}"
    echo -e "  Private Key: ${MKCERT_KEY}"
    echo ""

    # Verify it's a mkcert certificate
    if openssl x509 -in "${MKCERT_CERT}" -noout -text | grep -q "mkcert development"; then
        echo -e "${GREEN}✓ Verified: mkcert certificate (trusted by your browser)${NC}"
        echo ""

        # Copy to cert directory for consistency
        cp "${MKCERT_CERT}" "${CERT_DIR}/${DOMAIN}.crt"
        cp "${MKCERT_KEY}" "${CERT_DIR}/${DOMAIN}.key"

        # Create or update Kubernetes secret
        echo -e "${BLUE}Creating Kubernetes secret '${SECRET_NAME}' in namespace '${NAMESPACE}'...${NC}"

        # Delete secret if it exists
        if kubectl get secret "${SECRET_NAME}" -n "${NAMESPACE}" &> /dev/null; then
            echo -e "${YELLOW}Secret '${SECRET_NAME}' already exists, updating...${NC}"
            kubectl delete secret "${SECRET_NAME}" -n "${NAMESPACE}"
        fi

        # Create new secret from mkcert certificates
        kubectl create secret tls "${SECRET_NAME}" \
            --cert="${MKCERT_CERT}" \
            --key="${MKCERT_KEY}" \
            -n "${NAMESPACE}"

        echo ""
        echo -e "${GREEN}TLS secret '${SECRET_NAME}' created successfully from mkcert certificates!${NC}"
        echo ""
        echo -e "${CYAN}Certificate details:${NC}"
        openssl x509 -in "${MKCERT_CERT}" -noout -text | grep -A2 "Subject:"
        openssl x509 -in "${MKCERT_CERT}" -noout -text | grep -A1 "Validity"
        openssl x509 -in "${MKCERT_CERT}" -noout -text | grep "DNS:"
        echo ""
        echo -e "${GREEN}✓ Your browser will trust this certificate!${NC}"
        echo -e "${GREEN}✓ No SSL warnings for https://${DOMAIN}${NC}"
        echo ""
        exit 0
    else
        echo -e "${YELLOW}Warning: Certificate exists but is not from mkcert${NC}"
        echo -e "${YELLOW}Falling back to self-signed certificate generation...${NC}"
        echo ""
    fi
fi

# Check if self-signed certificates already exist
if [ -f "${CERT_DIR}/${DOMAIN}.crt" ] && [ -f "${CERT_DIR}/${DOMAIN}.key" ]; then
    echo -e "${YELLOW}Self-signed certificates already exist at ${CERT_DIR}${NC}"
    read -p "Do you want to regenerate them? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${GREEN}Using existing self-signed certificates${NC}"

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

# Generate self-signed certificates
echo -e "${YELLOW}No mkcert certificates found. Generating self-signed certificates...${NC}"
echo -e "${CYAN}Tip: Install mkcert for browser-trusted certificates:${NC}"
echo -e "  https://github.com/FiloSottile/mkcert#installation"
echo ""

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
echo -e "${GREEN}Self-signed certificates generated successfully:${NC}"
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
