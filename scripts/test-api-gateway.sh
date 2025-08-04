#!/bin/bash

# API Gateway Monitoring Test Script
# Tests all monitoring endpoints and validates responses

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

API_GATEWAY_URL="http://localhost:8080"

echo -e "${YELLOW}Testing API Gateway Monitoring Endpoints...${NC}"
echo

# Function to test JSON endpoint and validate structure
test_json_endpoint() {
    local name=$1
    local url=$2
    local expected_fields=$3
    
    echo -n "Testing $name... "
    
    if response=$(curl -s "$url" 2>/dev/null); then
        # Check if response is valid JSON
        if echo "$response" | jq . > /dev/null 2>&1; then
            # Check for expected fields
            missing_fields=""
            for field in $expected_fields; do
                if ! echo "$response" | jq -e "has(\"$field\")" > /dev/null 2>&1; then
                    missing_fields="$missing_fields $field"
                fi
            done
            
            if [ -z "$missing_fields" ]; then
                echo -e "${GREEN}✓ OK${NC}"
                return 0
            else
                echo -e "${RED}✗ FAIL (missing fields:$missing_fields)${NC}"
                return 1
            fi
        else
            echo -e "${RED}✗ FAIL (invalid JSON)${NC}"
            return 1
        fi
    else
        echo -e "${RED}✗ FAIL (connection error)${NC}"
        return 1
    fi
}

# Function to test endpoint with expected HTTP status
test_endpoint_status() {
    local name=$1
    local url=$2
    local expected_status=$3
    
    echo -n "Testing $name... "
    
    if response=$(curl -s -w "%{http_code}" -o /dev/null "$url" 2>/dev/null); then
        if [ "$response" -eq "$expected_status" ]; then
            echo -e "${GREEN}✓ OK (HTTP $response)${NC}"
            return 0
        else
            echo -e "${RED}✗ FAIL (HTTP $response, expected $expected_status)${NC}"
            return 1
        fi
    else
        echo -e "${RED}✗ FAIL (connection error)${NC}"
        return 1
    fi
}

# Function to test and display endpoint response
test_and_display() {
    local name=$1
    local url=$2
    
    echo -e "\n${YELLOW}📋 $name Response:${NC}"
    if response=$(curl -s "$url" 2>/dev/null); then
        echo "$response" | jq . 2>/dev/null || echo "$response"
    else
        echo -e "${RED}✗ Failed to get response${NC}"
    fi
}

# Test basic monitoring endpoints
echo -e "${YELLOW}Health & Readiness Endpoints${NC}"
test_json_endpoint "Health Check" "$API_GATEWAY_URL/health" "status timestamp uptime version checks"
test_json_endpoint "Readiness Check" "$API_GATEWAY_URL/ready" "status timestamp"
test_json_endpoint "Service Info" "$API_GATEWAY_URL/info" "service version timestamp uptime"
test_json_endpoint "App Metrics" "$API_GATEWAY_URL/app-metrics" "timestamp uptime memory_usage_bytes goroutine_count"
echo

# Test Prometheus metrics endpoint
echo -e "${YELLOW}Prometheus Metrics${NC}"
test_endpoint_status "Prometheus Metrics" "$API_GATEWAY_URL/metrics" 200
echo

# Test monitoring validation endpoints  
echo -e "${YELLOW}Monitoring Validation Endpoints${NC}"
test_json_endpoint "Test Trace" "$API_GATEWAY_URL/test/trace" "status message trace_id timestamp"
test_endpoint_status "Test Error" "$API_GATEWAY_URL/test/error" 500
echo

# Display detailed responses
test_and_display "Health Check Details" "$API_GATEWAY_URL/health"
test_and_display "Service Info Details" "$API_GATEWAY_URL/info"

# Test trace creation and get trace ID
echo -e "\n${YELLOW}Testing Trace Creation${NC}"
if trace_response=$(curl -s "$API_GATEWAY_URL/test/trace" 2>/dev/null); then
    trace_id=$(echo "$trace_response" | jq -r '.trace_id // empty' 2>/dev/null)
    if [ -n "$trace_id" ] && [ "$trace_id" != "null" ]; then
        echo -e "${GREEN}✓ Trace created successfully${NC}"
        echo "  Trace ID: $trace_id"
        
        # Wait a moment for trace to be processed
        sleep 2
        
        # Try to find the trace in Tempo (if available)
        echo -n "Checking trace in Tempo... "
        if tempo_response=$(curl -s "http://localhost:3200/api/traces/$trace_id" 2>/dev/null); then
            if echo "$tempo_response" | jq . > /dev/null 2>&1; then
                echo -e "${GREEN}✓ Found in Tempo${NC}"
            else
                echo -e "${YELLOW}⚠ Not yet available in Tempo${NC}"
            fi
        else
            echo -e "${YELLOW}⚠ Tempo not accessible${NC}"
        fi
    else
        echo -e "${RED}✗ Failed to create trace${NC}"
    fi
else
    echo -e "${RED}✗ Failed to call trace endpoint${NC}"
fi

# Test error creation and logging
echo -e "\n${YELLOW}Testing Error Creation${NC}"
if error_response=$(curl -s "$API_GATEWAY_URL/test/error" 2>/dev/null); then
    error_trace_id=$(echo "$error_response" | jq -r '.trace_id // empty' 2>/dev/null)
    if [ -n "$error_trace_id" ] && [ "$error_trace_id" != "null" ]; then
        echo -e "${GREEN}✓ Error trace created successfully${NC}"
        echo "  Error Trace ID: $error_trace_id"
    else
        echo -e "${YELLOW}⚠ Error created but no trace ID${NC}"
    fi
else
    echo -e "${RED}✗ Failed to call error endpoint${NC}"
fi

echo
echo -e "${GREEN}API Gateway monitoring test completed!${NC}"
echo -e "${YELLOW}Tips:${NC}"
echo "  - Check Grafana dashboards at http://localhost:3000"
echo "  - View traces in Tempo via Grafana"  
echo "  - Monitor metrics in Prometheus at http://localhost:9090"
echo "  - API Gateway available at $API_GATEWAY_URL"