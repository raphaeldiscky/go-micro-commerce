#!/bin/bash

# Quick Start Script for Consul and Traefik Setup
# This script helps you get started with the enhanced API Gateway setup

set -e

echo "🚀 Starting Consul and Traefik Setup..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE} $1${NC}"
}

print_success() {
    echo -e "${GREEN} $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}  $1${NC}"
}

print_error() {
    echo -e "${RED} $1${NC}"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

print_success "Docker is running"

# Create network if it doesn't exist
if ! docker network ls | grep -q go-micro-template; then
    print_status "Creating Docker network 'go-micro-template'..."
    docker network create go-micro-template
    print_success "Network created"
else
    print_success "Network 'go-micro-template' already exists"
fi

# Navigate to the docker-compose directory
cd "$(dirname "$0")/../deployments/docker-compose"

# Stop any existing containers
print_status "Stopping existing containers..."
docker-compose -f api-gateway.yaml down 2>/dev/null || true

# Start Consul and Traefik
print_status "Starting Consul and Traefik..."
docker-compose -f api-gateway.yaml up -d

# Wait for services to be healthy
print_status "Waiting for services to be healthy..."

# Wait for Consul
echo -n "Waiting for Consul to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:8500/v1/status/leader > /dev/null 2>&1; then
        echo ""
        print_success "Consul is ready"
        break
    fi
    echo -n "."
    sleep 2
    if [ $i -eq 30 ]; then
        echo ""
        print_error "Consul failed to start properly"
        exit 1
    fi
done

# Wait for Traefik
echo -n "Waiting for Traefik to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:9000/api/overview > /dev/null 2>&1; then
        echo ""
        print_success "Traefik is ready"
        break
    fi
    echo -n "."
    sleep 2
    if [ $i -eq 30 ]; then
        echo ""
        print_error "Traefik failed to start properly"
        exit 1
    fi
done

print_success "🎉 Setup complete!"

echo ""
echo "📋 Access Information:"
echo "  • Consul UI:    http://localhost:8500"
echo "  • Traefik UI:   http://localhost:9000"
echo ""

echo "🔍 Testing the setup..."

# Test Consul
print_status "Testing Consul API..."
CONSUL_LEADER=$(curl -s http://localhost:8500/v1/status/leader)
if [ ! -z "$CONSUL_LEADER" ]; then
    print_success "Consul API is working (Leader: $CONSUL_LEADER)"
else
    print_warning "Consul API test failed"
fi

# Test Traefik
print_status "Testing Traefik API..."
TRAEFIK_RESPONSE=$(curl -s http://localhost:9000/api/overview | jq -r '.http.routers.total' 2>/dev/null || echo "unknown")
if [ "$TRAEFIK_RESPONSE" != "unknown" ]; then
    print_success "Traefik API is working (Routers: $TRAEFIK_RESPONSE)"
else
    print_warning "Traefik API test failed or jq not installed"
fi

echo ""
echo "🧪 Next Steps:"
echo "  1. Register a test service with Consul:"
echo "     curl -X PUT http://localhost:8500/v1/agent/service/register \\"
echo "       -d '{\"ID\": \"test-service-1\", \"Name\": \"test-service\", \"Tags\": [\"api\"], \"Address\": \"127.0.0.1\", \"Port\": 8080}'"
echo ""
echo "  2. Check if Traefik discovered the service:"
echo "     curl http://localhost:9000/api/http/services"
echo ""
echo "  3. Start your microservices with Consul registration enabled"
echo ""

print_status "Setup logs are available via:"
echo "  docker-compose -f api-gateway-enhanced.yaml logs -f"

echo ""
print_success "🚀 Consul and Traefik are ready for your microservices!"

# Optional: Open browser windows
if command -v open > /dev/null 2>&1; then  # macOS
    print_status "Opening Consul and Traefik dashboards..."
    sleep 2
    open http://localhost:8500 2>/dev/null &
    open http://localhost:9000 2>/dev/null &
elif command -v xdg-open > /dev/null 2>&1; then  # Linux
    print_status "Opening Consul and Traefik dashboards..."
    sleep 2
    xdg-open http://localhost:8500 2>/dev/null &
    xdg-open http://localhost:9000 2>/dev/null &
fi
