#!/bin/bash

# Monitoring Stack Health Check Script
# This script tests all monitoring components to ensure they're working properly

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🔍 Testing Monitoring Stack Health...${NC}"
echo

# Function to test HTTP endpoint
test_endpoint() {
    local name=$1
    local url=$2
    local expected_status=${3:-200}
    
    echo -n "Testing $name... "
    
    if response=$(curl -s -w "%{http_code}" -o /dev/null "$url" 2>/dev/null); then
        if [ "$response" -eq "$expected_status" ]; then
            echo -e "${GREEN}✓ OK${NC}"
            return 0
        else
            echo -e "${RED}✗ FAIL (HTTP $response)${NC}"
            return 1
        fi
    else
        echo -e "${RED}✗ FAIL (Connection error)${NC}"
        return 1
    fi
}

# Function to test JSON endpoint and extract data
test_json_endpoint() {
    local name=$1
    local url=$2
    local jq_filter=$3
    
    echo -n "Testing $name... "
    
    if response=$(curl -s "$url" 2>/dev/null); then
        if echo "$response" | jq -e "$jq_filter" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ OK${NC}"
            return 0
        else
            echo -e "${RED}✗ FAIL (Invalid response)${NC}"
            return 1
        fi
    else
        echo -e "${RED}✗ FAIL (Connection error)${NC}"
        return 1
    fi
}

# Function to test metrics endpoint
test_metrics() {
    local name=$1
    local url=$2
    local metric_name=$3
    
    echo -n "Testing $name... "
    
    if response=$(curl -s "$url" 2>/dev/null); then
        if echo "$response" | grep -q "$metric_name"; then
            echo -e "${GREEN}✓ OK (found $metric_name)${NC}"
            return 0
        else
            echo -e "${RED}✗ FAIL (metric $metric_name not found)${NC}"
            return 1
        fi
    else
        echo -e "${RED}✗ FAIL (Connection error)${NC}"
        return 1
    fi
}

# Test basic connectivity
echo -e "${YELLOW}Basic Connectivity Tests${NC}"
test_endpoint "Prometheus" "http://localhost:9090/-/healthy"
test_endpoint "Grafana" "http://localhost:3000/api/health"
test_endpoint "Loki" "http://localhost:3100/ready"
test_endpoint "Tempo" "http://localhost:3200/ready"
test_endpoint "OpenTelemetry Collector" "http://localhost:13133/health"
echo

# Test Prometheus targets
echo -e "${YELLOW}Prometheus Targets${NC}"
test_json_endpoint "Prometheus Targets" "http://localhost:9090/api/v1/targets" '.status == "success"'
echo

# Test Prometheus metrics
echo -e "${YELLOW}Prometheus Metrics${NC}"
test_metrics "Prometheus Self Metrics" "http://localhost:9090/metrics" "prometheus_build_info"
test_metrics "OpenTelemetry Collector Metrics" "http://localhost:8889/metrics" "otelcol_process_uptime"
echo

# Test Grafana data sources
echo -e "${YELLOW}Grafana Data Sources${NC}"
test_json_endpoint "Grafana Data Sources" "http://admin:admin@localhost:3000/api/datasources" 'length > 0'
echo

# Test Loki labels
echo -e "${YELLOW}Loki Labels${NC}"
test_json_endpoint "Loki Labels" "http://localhost:3100/loki/api/v1/labels" '.status == "success"'
echo

# Test OpenTelemetry Collector endpoints
echo -e "${YELLOW}OpenTelemetry Collector${NC}"
# Note: gRPC port 4317 doesn't respond to HTTP requests, which is expected
test_endpoint "OTLP HTTP Receiver" "http://localhost:4318" 404  # Should return 404 for GET on POST endpoint
test_endpoint "Collector Health" "http://localhost:13133/health" 
echo

# Test Docker containers
echo -e "${YELLOW}Docker Containers Status${NC}"
containers=("prometheus" "grafana" "loki" "tempo" "otel-collector")

for container in "${containers[@]}"; do
    echo -n "Testing $container container... "
    if docker ps --filter "name=$container" --filter "status=running" --format "{{.Names}}" | grep -q "^${container}$"; then
        echo -e "${GREEN}✓ Running${NC}"
    else
        echo -e "${RED}✗ Not running or unhealthy${NC}"
    fi
done
echo

# Summary
echo -e "${YELLOW}Test Summary${NC}"
echo "All basic monitoring stack components tested."
echo "For more detailed testing, run individual component tests."
echo
echo -e "${GREEN}Monitoring stack health check completed!${NC}"