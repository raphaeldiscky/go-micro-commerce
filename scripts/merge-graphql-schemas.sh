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

echo ""
echo -e "${GREEN}Schema merging complete for ${#SERVICES[@]} services.${NC}"
