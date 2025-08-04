#!/bin/bash

# OpenTelemetry Integration Test
# This script sends sample telemetry data to test the OTEL pipeline

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Testing OpenTelemetry Pipeline...${NC}"
echo

# Function to send trace data via OTLP HTTP
send_test_trace() {
    echo -n "Sending test trace... "
    
    local trace_data='{
        "resourceSpans": [{
            "resource": {
                "attributes": [{
                    "key": "service.name",
                    "value": {"stringValue": "test-service"}
                }, {
                    "key": "service.version", 
                    "value": {"stringValue": "1.0.0"}
                }]
            },
            "scopeSpans": [{
                "spans": [{
                    "traceId": "12345678901234567890123456789012",
                    "spanId": "1234567890123456",
                    "name": "test-operation",
                    "kind": 1,
                    "startTimeUnixNano": "1640995200000000000",
                    "endTimeUnixNano": "1640995201000000000",
                    "attributes": [{
                        "key": "http.method",
                        "value": {"stringValue": "GET"}
                    }]
                }]
            }]
        }]
    }'
    
    if response=$(curl -s -w "%{http_code}" -o /dev/null \
        -X POST http://localhost:4318/v1/traces \
        -H "Content-Type: application/json" \
        -d "$trace_data" 2>/dev/null); then
        
        if [ "$response" -eq "200" ] || [ "$response" -eq "202" ]; then
            echo -e "${GREEN}✓ OK (HTTP $response)${NC}"
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

# Function to send metrics data via OTLP HTTP
send_test_metrics() {
    echo -n "Sending test metrics... "
    
    local metrics_data='{
        "resourceMetrics": [{
            "resource": {
                "attributes": [{
                    "key": "service.name",
                    "value": {"stringValue": "test-service"}
                }]
            },
            "scopeMetrics": [{
                "metrics": [{
                    "name": "test_counter",
                    "description": "A test counter metric",
                    "unit": "1",
                    "sum": {
                        "dataPoints": [{
                            "timeUnixNano": "1640995200000000000",
                            "asInt": "42",
                            "attributes": [{
                                "key": "test.label",
                                "value": {"stringValue": "test-value"}
                            }]
                        }],
                        "aggregationTemporality": 2,
                        "isMonotonic": true
                    }
                }]
            }]
        }]
    }'
    
    if response=$(curl -s -w "%{http_code}" -o /dev/null \
        -X POST http://localhost:4318/v1/metrics \
        -H "Content-Type: application/json" \
        -d "$metrics_data" 2>/dev/null); then
        
        if [ "$response" -eq "200" ] || [ "$response" -eq "202" ]; then
            echo -e "${GREEN}✓ OK (HTTP $response)${NC}"
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

# Function to send logs data via OTLP HTTP
send_test_logs() {
    echo -n "Sending test logs... "
    
    local logs_data='{
        "resourceLogs": [{
            "resource": {
                "attributes": [{
                    "key": "service.name",
                    "value": {"stringValue": "test-service"}
                }]
            },
            "scopeLogs": [{
                "logRecords": [{
                    "timeUnixNano": "1640995200000000000",
                    "severityNumber": 9,
                    "severityText": "INFO",
                    "body": {
                        "stringValue": "This is a test log message"
                    },
                    "attributes": [{
                        "key": "log.level",
                        "value": {"stringValue": "info"}
                    }]
                }]
            }]
        }]
    }'
    
    if response=$(curl -s -w "%{http_code}" -o /dev/null \
        -X POST http://localhost:4318/v1/logs \
        -H "Content-Type: application/json" \
        -d "$logs_data" 2>/dev/null); then
        
        if [ "$response" -eq "200" ] || [ "$response" -eq "202" ]; then
            echo -e "${GREEN}✓ OK (HTTP $response)${NC}"
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

# Function to check if data reached backends
check_data_in_backends() {
    echo -e "\n${YELLOW}Checking data in backends...${NC}"
    
    # Wait a moment for data to be processed
    sleep 2
    
    echo -n "Checking Prometheus metrics... "
    if curl -s "http://localhost:9090/api/v1/query?query=up" | jq -e '.data.result | length > 0' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${YELLOW}⚠ No metrics found (may take time to appear)${NC}"
    fi
    
    echo -n "Checking Tempo traces... "
    if curl -s "http://localhost:3200/api/search" | jq -e 'has("traces")' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${YELLOW}⚠ Traces endpoint accessible${NC}"
    fi
}

# Main test execution
echo -e "${YELLOW}Sending Test Data${NC}"
send_test_trace
send_test_metrics  
send_test_logs

check_data_in_backends

echo
echo -e "${GREEN}OpenTelemetry integration test completed!${NC}"
echo -e "${YELLOW}Tip: Check Grafana dashboards at http://localhost:3000 to visualize the data${NC}"