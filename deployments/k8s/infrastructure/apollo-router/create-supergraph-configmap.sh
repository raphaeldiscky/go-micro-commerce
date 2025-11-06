#!/bin/bash
# Script to create/update Apollo supergraph schema ConfigMap
# This script creates a ConfigMap from the pre-composed supergraph schema

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
SCHEMA_FILE="${ROOT_DIR}/graphql-gateway/supergraph-schema.graphql"

if [ ! -f "${SCHEMA_FILE}" ]; then
  echo "Error: Supergraph schema file not found at ${SCHEMA_FILE}"
  exit 1
fi

echo "Creating ConfigMap from supergraph schema..."
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

echo "ConfigMap created at: ${SCRIPT_DIR}/configmap-supergraph-schema.yaml"
echo ""
echo "To apply the ConfigMap, run:"
echo "  kubectl apply -f ${SCRIPT_DIR}/configmap-supergraph-schema.yaml"
