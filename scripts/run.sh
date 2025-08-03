#!/bin/bash

# Usage: ./run.sh [service-name]
# Examples:
#   ./run.sh                    # Run all services with air
#   ./run.sh api-gateway        # Run api-gateway with air
#   ./run.sh product            # Run product-service with air
#   ./run.sh product-service    # Run product-service with air

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to check if air is installed
check_air() {
    if ! command -v air &> /dev/null; then
        echo -e "${RED}Air is not installed. Please install it first:${NC}"
        echo -e "${YELLOW}go install github.com/cosmtrek/air@latest${NC}"
        exit 1
    fi
}

# Function to run service with air
run_service_with_air() {
    local service_dir="$1"
    local service_name="$2"
    
    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}Service directory $service_dir does not exist${NC}"
        return 1
    fi
    
    if [ ! -f "$service_dir/.air.toml" ]; then
        echo -e "${RED}Air configuration not found in $service_dir${NC}"
        return 1
    fi
    
    echo -e "${GREEN}Starting $service_name with air...${NC}"
    echo -e "${BLUE}Service directory: $service_dir${NC}"
    
    cd "$service_dir"
    air
    cd - > /dev/null
}

# Function to run all services concurrently
run_all_services() {
    echo -e "${YELLOW}Starting all services with air...${NC}"
    
    # Array to store PIDs of background processes
    pids=""
    
    # Run api-gateway in background
    if [ -d "api-gateway" ] && [ -f "api-gateway/.air.toml" ]; then
        echo -e "${GREEN}Starting api-gateway...${NC}"
        cd api-gateway
        air &
        pids="$pids $!"
        cd - > /dev/null
    fi
    
    # Run all services in services directory
    if [ -d "services" ]; then
        for service_dir in services/*/; do
            if [ -d "$service_dir" ] && [ -f "$service_dir/.air.toml" ]; then
                service_name=$(basename "$service_dir")
                echo -e "${GREEN}Starting $service_name...${NC}"
                cd "$service_dir"
                air &
                pids="$pids $!"
                cd - > /dev/null
            fi
        done
    fi
    
    if [ -z "$pids" ]; then
        echo -e "${RED}No services with air configuration found${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}All services started. Press Ctrl+C to stop all services.${NC}"
    
    # Function to handle cleanup on exit
    cleanup() {
        echo -e "\n${YELLOW}Stopping all services...${NC}"
        for pid in $pids; do
            if kill -0 "$pid" 2>/dev/null; then
                kill "$pid"
            fi
        done
        wait
        echo -e "${GREEN}All services stopped${NC}"
        exit 0
    }
    
    # Set up signal handlers
    trap cleanup SIGINT SIGTERM
    
    # Wait for all background processes
    wait
}

# Function to get service directory from service name
get_service_directory() {
    local service_name="$1"
    
    case "$service_name" in
        "api-gateway")
            echo "api-gateway"
            ;;
        "product"|"product-service")
            echo "services/product-service"
            ;;
        *)
            # Try to find in services directory
            if [ -d "services/$service_name" ]; then
                echo "services/$service_name"
            elif [ -d "services/${service_name}-service" ]; then
                echo "services/${service_name}-service"
            else
                echo ""
            fi
            ;;
    esac
}

# Main execution
main() {
    echo -e "${BLUE}Go Microservices Live Reload Runner${NC}"
    echo -e "${BLUE}====================================${NC}"
    
    # Check if air is installed
    check_air
    
    # Get service name from argument
    service_name="$1"
    
    if [ -z "$service_name" ]; then
        # No service specified, run all services
        run_all_services
    else
        # Specific service requested
        service_dir=$(get_service_directory "$service_name")
        
        if [ -z "$service_dir" ]; then
            echo -e "${RED}Service '$service_name' not found${NC}"
            echo -e "${YELLOW}Available services:${NC}"
            echo -e "  - api-gateway"
            if [ -d "services" ]; then
                for dir in services/*/; do
                    if [ -d "$dir" ]; then
                        basename "$dir"
                    fi
                done | sed 's/^/  - /'
            fi
            exit 1
        fi
        
        run_service_with_air "$service_dir" "$service_name"
    fi
}

# Run main function with all arguments
main "$@"
