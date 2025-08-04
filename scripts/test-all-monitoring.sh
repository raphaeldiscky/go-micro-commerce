#!/bin/bash

# Master Monitoring Test Script
# Runs comprehensive tests for the entire monitoring stack

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${BLUE}Starting Comprehensive Monitoring Stack Tests${NC}"
echo -e "${BLUE}=================================================${NC}"
echo

# Function to run a test script and capture results
run_test() {
    local test_name=$1
    local script_path=$2
    
    echo -e "${YELLOW}Running $test_name...${NC}"
    echo -e "${YELLOW}$(printf '=%.0s' {1..50})${NC}"
    
    if [ -f "$script_path" ]; then
        if bash "$script_path"; then
            echo -e "${GREEN}$test_name PASSED${NC}"
            return 0
        else
            echo -e "${RED}$test_name FAILED${NC}"
            return 1
        fi
    else
        echo -e "${RED}Test script not found: $script_path${NC}"
        return 1
    fi
}

# Track test results
total_tests=0
passed_tests=0
failed_tests=0

# Run monitoring stack health tests
total_tests=$((total_tests + 1))
if run_test "Monitoring Stack Health Check" "$SCRIPT_DIR/test-monitoring.sh"; then
    passed_tests=$((passed_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi

echo

# Run OpenTelemetry integration tests
total_tests=$((total_tests + 1))
if run_test "OpenTelemetry Pipeline Test" "$SCRIPT_DIR/test-otel.sh"; then
    passed_tests=$((passed_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi

echo

# Run API Gateway monitoring tests
total_tests=$((total_tests + 1))
if run_test "API Gateway Monitoring Test" "$SCRIPT_DIR/test-api-gateway.sh"; then
    passed_tests=$((passed_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi

echo

# Summary
echo -e "${BLUE}Test Results Summary${NC}"
echo -e "${BLUE}======================${NC}"
echo -e "Total Tests: $total_tests"
echo -e "${GREEN}Passed: $passed_tests${NC}"
echo -e "${RED}Failed: $failed_tests${NC}"

if [ $failed_tests -eq 0 ]; then
    echo
    echo -e "${GREEN}All monitoring tests passed successfully!${NC}"
    echo -e "${GREEN}Your monitoring stack is fully operational.${NC}"
    echo
    echo -e "${YELLOW}Quick Links:${NC}"
    echo "  • Grafana:    http://localhost:3000 (admin/admin)"
    echo "  • Prometheus: http://localhost:9090"
    echo "  • API Gateway: http://localhost:8080"
    echo "  • Health:     http://localhost:8080/health"
    echo
    exit 0
else
    echo
    echo -e "${RED}Some tests failed. Please check the logs above.${NC}"
    echo -e "${YELLOW}Common issues:${NC}"
    echo "  • Ensure all services are running: docker-compose -f monitoring.yaml ps"
    echo "  • Check service logs: docker-compose -f monitoring.yaml logs [service]"
    echo "  • Verify network connectivity between services"
    echo
    exit 1
fi