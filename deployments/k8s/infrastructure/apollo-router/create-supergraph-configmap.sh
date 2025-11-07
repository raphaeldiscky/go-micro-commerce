#!/bin/bash
# Script to compose K8s supergraph schema and create ConfigMap
# This script generates a K8s-specific supergraph schema using rover compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
SUPERGRAPH_CONFIG="${SCRIPT_DIR}/supergraph.yaml"
SCHEMA_FILE="${SCRIPT_DIR}/supergraph-schema.graphql"
SCHEMAS_DIR="${SCRIPT_DIR}/schemas"

echo -e "${BLUE}Apollo Federation Supergraph Composition for Kubernetes${NC}"
echo "========================================================="
echo ""

# Check if rover is installed
if ! command -v rover &> /dev/null; then
    echo -e "${YELLOW}Rover CLI not found. Installing...${NC}"
    curl -sSL https://rover.apollo.dev/nix/latest | sh
    export PATH="$HOME/.rover/bin:$PATH"

    if ! command -v rover &> /dev/null; then
        echo -e "${RED}Failed to install Rover CLI${NC}"
        echo "Please install manually: https://www.apollographql.com/docs/rover/getting-started"
        exit 1
    fi
    echo -e "${GREEN}Rover CLI installed successfully${NC}"
fi

echo -e "${BLUE}Rover version:${NC} $(rover --version)"
echo ""

# Check if supergraph.yaml exists
if [ ! -f "${SUPERGRAPH_CONFIG}" ]; then
    echo -e "${RED}Error: ${SUPERGRAPH_CONFIG} not found${NC}"
    exit 1
fi

# Create schemas directory and copy schema files from services
echo -e "${BLUE}Copying schema files from services...${NC}"
mkdir -p "${SCHEMAS_DIR}"

services=("auth-service" "cart-service" "chat-service" "notification-service" "order-service" "payment-service")
for service in "${services[@]}"; do
    service_schema="${ROOT_DIR}/${service}/graph/schema.graphqls"
    if [ -f "${service_schema}" ]; then
        cp "${service_schema}" "${SCHEMAS_DIR}/${service}.graphqls"
        echo -e "${GREEN}✓${NC} Copied ${service}"
    else
        echo -e "${RED}✗${NC} Schema not found: ${service_schema}"
        exit 1
    fi
done
echo ""

# Compose supergraph
echo -e "${BLUE}Composing supergraph schema for Kubernetes...${NC}"
export APOLLO_ELV2_LICENSE=accept

if rover supergraph compose --config "${SUPERGRAPH_CONFIG}" > "${SCHEMA_FILE}"; then
    echo -e "${GREEN}✓ Supergraph schema composed successfully${NC}"
    echo ""
    echo -e "${BLUE}Output:${NC} ${SCHEMA_FILE}"

    # Show schema stats
    lines=$(wc -l < "${SCHEMA_FILE}")
    size=$(du -h "${SCHEMA_FILE}" | cut -f1)
    echo -e "${BLUE}Schema size:${NC} $lines lines ($size)"

    # Verify K8s service URLs
    echo ""
    echo -e "${BLUE}Verifying K8s service URLs:${NC}"
    grep "@join__graph" "${SCHEMA_FILE}" | head -3
    echo ""
else
    echo -e "${RED}✗ Supergraph composition failed${NC}"
    exit 1
fi

# Create ConfigMap
echo -e "${BLUE}Creating ConfigMap...${NC}"
kubectl create configmap apollo-supergraph-schema \
  --from-file=supergraph-schema.graphql="${SCHEMA_FILE}" \
  --dry-run=client \
  -o yaml \
  > "${SCRIPT_DIR}/configmap-supergraph-schema.yaml"

# Add labels
cat >> "${SCRIPT_DIR}/configmap-supergraph-schema.yaml" <<EOF
  labels:
    app: graphql-gateway
    component: apollo-router
    managed-by: script
EOF

echo -e "${GREEN}✓ ConfigMap created at: ${SCRIPT_DIR}/configmap-supergraph-schema.yaml${NC}"
echo ""
echo "To apply the ConfigMap, run:"
echo -e "  ${BLUE}kubectl apply -f ${SCRIPT_DIR}/configmap-supergraph-schema.yaml${NC}"
echo ""
echo "Then restart Apollo Router:"
echo -e "  ${BLUE}kubectl rollout restart deployment/local-apollo-router -n default${NC}"
