#!/usr/bin/env bash

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Merging GraphQL schemas...${NC}"
echo ""

SERVICES=()

shopt -s nullglob

for dir in */ ; do
  dir="${dir%/}"
  schema_dir="$dir/graph/schema"
  if compgen -G "$schema_dir/*.graphql" > /dev/null; then
    SERVICES+=("$dir")
  fi
done

# Function to merge schema files
merge_schemas() {
    local service=$1
    local schema_dir="$service/graph/schema"
    local output_file="$service/graph/schema.graphqls"

    if [ ! -d "$schema_dir" ]; then
        echo "Warning: Schema directory not found: $schema_dir"
        return 1
    fi

    # Remove old merged file
    rm -f "$output_file"

    # Concatenate all .graphql files
    cat "$schema_dir"/*.graphql > "$output_file"

    echo -e "${GREEN}✓${NC} $service → merged into $output_file"
}

# Merge schemas for each service
for service in "${SERVICES[@]}"; do
  echo -e "${BLUE}Merging schemas for service:${NC} $service"
  merge_schemas "$service"
done

# Copy schemas to graphql-gateway directory (for local development)
echo ""
echo -e "${BLUE}Copying schemas to graphql-gateway...${NC}"
mkdir -p graphql-gateway/schemas
for service in "${SERVICES[@]}"; do
  if [ -f "$service/graph/schema.graphqls" ]; then
    cp "$service/graph/schema.graphqls" "graphql-gateway/schemas/$service.graphqls"
    echo -e "${GREEN}✓${NC} $service → graphql-gateway/schemas/$service.graphqls"
  fi
done

# Copy schemas to apollo-router k8s directory (for Kubernetes deployment)
echo ""
echo -e "${BLUE}Copying schemas to apollo-router k8s...${NC}"
mkdir -p deployments/k8s/infrastructure/apollo-router/schemas
for service in "${SERVICES[@]}"; do
  if [ -f "$service/graph/schema.graphqls" ]; then
    cp "$service/graph/schema.graphqls" "deployments/k8s/infrastructure/apollo-router/schemas/$service.graphqls"
    echo -e "${GREEN}✓${NC} $service → deployments/k8s/infrastructure/apollo-router/schemas/$service.graphqls"
  fi
done

echo ""
echo -e "${GREEN}Schema merging and distribution complete for ${#SERVICES[@]} services.${NC}"
