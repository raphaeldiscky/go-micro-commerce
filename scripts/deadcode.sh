#!/bin/bash

# deadcode.sh - Run deadcode analysis on all microservices
# This script runs the official Go deadcode tool on each microservice to find unreachable functions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Microservices list - directories with main.go files
MICROSERVICES=(
    "api-gateway"
    "auth-service"
    "fulfillment-service"
    "notification-service"
    "order-service"
    "payment-service"
    "product-service"
    "search-service"
)

# Function to print colored output
print_header() {
    echo -e "${BLUE}=====================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=====================================${NC}"
}

print_success() {
    echo -e "${GREEN} $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}� $1${NC}"
}

print_error() {
    echo -e "${RED} $1${NC}"
}

# Function to check if deadcode is installed
check_deadcode() {
    if ! command -v deadcode &> /dev/null; then
        print_error "deadcode is not installed"
        echo -e "Installing deadcode..."
        go install golang.org/x/tools/cmd/deadcode@latest
        if ! command -v deadcode &> /dev/null; then
            print_error "Failed to install deadcode"
            exit 1
        fi
        print_success "deadcode installed successfully"
    else
        print_success "deadcode is available"
    fi
}

# Function to run deadcode on a specific service
run_deadcode() {
    local service=$1
    local service_dir="$ROOT_DIR/$service"
    
    if [ ! -d "$service_dir" ]; then
        print_warning "Directory $service_dir does not exist, skipping..."
        return 0
    fi

    # Check if service has main.go
    if [ ! -f "$service_dir/cmd/api/main.go" ]; then
        print_warning "No main.go found in $service/cmd/api/, skipping..."
        return 0
    fi

    echo -e "\n${YELLOW}Analyzing $service...${NC}"
    
    cd "$service_dir" || exit 1
    
    # Run deadcode from the main package entry point
    local deadcode_output
    deadcode_output=$(deadcode ./cmd/api 2>&1) || true
    
    if [ -n "$deadcode_output" ]; then
        echo -e "${RED}Found unreachable code in $service:${NC}"
        echo "$deadcode_output"
        echo ""
        return 1
    else
        print_success "No deadcode found in $service"
        return 0
    fi
}

# Function to run deadcode on all services
run_all() {
    local has_deadcode=false
    local failed_services=()
    
    for service in "${MICROSERVICES[@]}"; do
        if ! run_deadcode "$service"; then
            has_deadcode=true
            failed_services+=("$service")
        fi
    done
    
    echo ""
    print_header "DEADCODE ANALYSIS SUMMARY"
    
    if [ "$has_deadcode" = true ]; then
        print_error "Deadcode found in ${#failed_services[@]} service(s): ${failed_services[*]}"
        echo ""
        echo -e "${YELLOW}To fix deadcode issues:${NC}"
        echo "1. Review the reported functions and determine if they're truly unused"
        echo "2. Remove unused functions or add them to your API if they should be exposed"
        echo "3. Consider if functions are used by tests or external packages"
        echo "4. Re-run this script to verify fixes"
        return 1
    else
        print_success "No deadcode found in any microservice!"
        return 0
    fi
}

# Function to run deadcode on specific service
run_service() {
    local service=$1
    
    if [[ ! " ${MICROSERVICES[@]} " =~ " $service " ]]; then
        print_error "Unknown service: $service"
        echo -e "Available services: ${MICROSERVICES[*]}"
        exit 1
    fi
    
    run_deadcode "$service"
}

# Function to show help
show_help() {
    cat << EOF
Usage: $0 [OPTIONS] [SERVICE]

Run deadcode analysis on microservices to find unreachable functions.

OPTIONS:
    -h, --help      Show this help message
    -l, --list      List available services
    -a, --all       Run on all services (default)
    -s, --service   Run on specific service

EXAMPLES:
    $0                          # Run on all services
    $0 -a                       # Run on all services  
    $0 -s search-service        # Run on search-service only
    $0 search-service           # Run on search-service only (shorthand)

SERVICES:
    ${MICROSERVICES[*]}

ABOUT DEADCODE:
The deadcode tool finds unreachable functions by analyzing the call graph from
main packages. It only reports functions that are truly unreachable from your
application entry points.

EOF
}

# Function to list services
list_services() {
    print_header "AVAILABLE MICROSERVICES"
    for service in "${MICROSERVICES[@]}"; do
        if [ -d "$ROOT_DIR/$service" ] && [ -f "$ROOT_DIR/$service/cmd/api/main.go" ]; then
            print_success "$service"
        else
            print_warning "$service (not found or missing main.go)"
        fi
    done
}

# Main script logic
main() {
    print_header "DEADCODE ANALYSIS"
    
    # Check if deadcode is installed
    check_deadcode
    
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -l|--list)
            list_services
            exit 0
            ;;
        -a|--all|"")
            run_all
            ;;
        -s|--service)
            if [ -z "${2:-}" ]; then
                print_error "Service name required with -s option"
                echo "Use -l to list available services"
                exit 1
            fi
            run_service "$2"
            ;;
        *)
            # Assume it's a service name
            run_service "$1"
            ;;
    esac
}

# Run main function with all arguments
main "$@"