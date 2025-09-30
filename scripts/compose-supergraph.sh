#!/usr/bin/env bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Apollo Federation Supergraph Composition${NC}"
echo "========================================="
echo ""

# Check if Rover is installed
if ! command -v rover &> /dev/null; then
    echo -e "${YELLOW}Rover CLI not found. Installing...${NC}"

    # Install Rover
    curl -sSL https://rover.apollo.dev/nix/latest | sh

    # Add to PATH for this session
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

# Define paths
GATEWAY_DIR="graphql-gateway"
SUPERGRAPH_CONFIG="$GATEWAY_DIR/supergraph.yaml"
OUTPUT_FILE="$GATEWAY_DIR/supergraph-schema.graphql"

# Check if supergraph.yaml exists
if [ ! -f "$SUPERGRAPH_CONFIG" ]; then
    echo -e "${RED}Error: $SUPERGRAPH_CONFIG not found${NC}"
    exit 1
fi

echo -e "${BLUE}Composing supergraph from:${NC} $SUPERGRAPH_CONFIG"
echo ""

# Check if subgraph services are running
echo -e "${YELLOW}Checking subgraph availability...${NC}"

check_service() {
    local service_url=$1
    local service_name=$2

    if curl -s -f -X POST "$service_url" \
        -H "Content-Type: application/json" \
        -d '{"query": "{ __schema { queryType { name } } }"}' > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $service_name is running"
        return 0
    else
        echo -e "${RED}✗${NC} $service_name is not accessible at $service_url"
        return 1
    fi
}

# Check each subgraph
CHAT_SERVICE_URL="http://localhost:8088/graph"
AUTH_SERVICE_URL="http://localhost:8081/graph"

all_services_up=true

check_service "$CHAT_SERVICE_URL" "chat-service" || all_services_up=false
check_service "$AUTH_SERVICE_URL" "auth-service" || all_services_up=false

echo ""

if [ "$all_services_up" = false ]; then
    echo -e "${YELLOW}Warning: Not all subgraph services are running${NC}"
    echo "Make sure services are started before composing:"
    echo "  task start_infra"
    echo "  task start_apps"
    echo ""
    echo -e "${YELLOW}Attempting composition anyway...${NC}"
    echo ""
fi

# Compose supergraph
echo -e "${BLUE}Composing supergraph schema...${NC}"

# Accept ELv2 license if not already accepted
export APOLLO_ELV2_LICENSE=accept

if rover supergraph compose --config "$SUPERGRAPH_CONFIG" > "$OUTPUT_FILE"; then
    echo -e "${GREEN}✓ Supergraph schema composed successfully${NC}"
    echo ""
    echo -e "${BLUE}Output:${NC} $OUTPUT_FILE"

    # Show schema stats
    lines=$(wc -l < "$OUTPUT_FILE")
    size=$(du -h "$OUTPUT_FILE" | cut -f1)
    echo -e "${BLUE}Schema size:${NC} $lines lines ($size)"

    # Count types
    types=$(grep -c "^type " "$OUTPUT_FILE" || true)
    echo -e "${BLUE}Types:${NC} $types"

    echo ""
    echo -e "${GREEN}Supergraph composition complete!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Start the GraphQL Gateway:"
    echo "     docker-compose -f deployments/docker-compose/graphql-gateway.yaml up"
    echo ""
    echo "  2. Access GraphQL endpoint:"
    echo "     http://localhost:4000/graphql"

    exit 0
else
    echo -e "${RED}✗ Supergraph composition failed${NC}"
    echo ""
    echo "Troubleshooting:"
    echo "  1. Ensure all subgraph services are running"
    echo "  2. Check subgraph GraphQL endpoints are accessible"
    echo "  3. Verify schema compatibility"
    echo "  4. Check Rover logs above for detailed errors"

    exit 1
fi