#!/bin/bash

# Build Script for Go Microservices Template
# ==========================================
# 
# This script builds all microservices in the project or individual services.
# It supports cross-compilation and provides clean build artifacts management.
#
# Features:
# - Build all services or individual services
# - Cross-compilation support (GOOS/GOARCH environment variables)
# - Optimized builds with link-time optimizations (-w -s flags)
# - Sonic and AVX tags for better performance
# - Colored output for better visibility
# - Build summary with success/failure status
# - Clean functionality to remove build artifacts
# - Binary size reporting
#
# Usage:
#   ./scripts/build.sh                    # Build all services
#   ./scripts/build.sh build api-gateway  # Build specific service
#   ./scripts/build.sh clean              # Clean all build artifacts
#   ./scripts/build.sh help               # Show help

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Build configuration
BUILD_LDFLAGS="-w -s"
BUILD_TAGS="sonic avx"
TARGET_OS="${GOOS:-linux}"
TARGET_ARCH="${GOARCH:-amd64}"

# Services to build
SERVICES=(
    "api-gateway"
    "services/auth-service"
    "services/notification-service"
    "services/order-service"
    "services/product-service"
)

# Function to build a single service
build_service() {
    local service_path="$1"
    local service_name=$(basename "$service_path")
    
    print_info "Building $service_name..."
    
    # Check if service directory exists
    if [ ! -d "$service_path" ]; then
        print_error "Service directory $service_path does not exist"
        return 1
    fi
    
    # Check if main.go exists
    local main_file="$service_path/cmd/main.go"
    if [ ! -f "$main_file" ]; then
        print_error "Main file not found at $main_file"
        return 1
    fi
    
    # Create bin directory if it doesn't exist
    mkdir -p "$service_path/bin"
    
    # Change to service directory
    cd "$service_path"
    
    # Build the service
    print_info "Compiling $service_name with CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH"
    
    if CGO_ENABLED=0 GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build \
        -ldflags="$BUILD_LDFLAGS" \
        -tags="$BUILD_TAGS" \
        -v \
        -o "./bin/main" \
        "./cmd/main.go"; then
        print_success "Successfully built $service_name"
        
        # Print binary info
        if [ -f "./bin/main" ]; then
            local binary_size=$(du -h "./bin/main" | cut -f1)
            print_info "Binary size: $binary_size"
        fi
    else
        print_error "Failed to build $service_name"
        return 1
    fi
    
    # Return to root directory
    cd - > /dev/null
}

# Function to build all services
build_all() {
    print_info "Starting build process for all services..."
    
    local failed_services=()
    local successful_services=()
    
    for service in "${SERVICES[@]}"; do
        if build_service "$service"; then
            successful_services+=("$service")
        else
            failed_services+=("$service")
        fi
        echo ""
    done
    
    # Print summary
    echo "=================================="
    print_info "Build Summary:"
    
    if [ ${#successful_services[@]} -gt 0 ]; then
        print_success "Successfully built ${#successful_services[@]} service(s):"
        for service in "${successful_services[@]}"; do
            echo "  ✓ $(basename "$service")"
        done
    fi
    
    if [ ${#failed_services[@]} -gt 0 ]; then
        print_error "Failed to build ${#failed_services[@]} service(s):"
        for service in "${failed_services[@]}"; do
            echo "  ✗ $(basename "$service")"
        done
        exit 1
    fi
    
    print_success "All services built successfully!"
}

# Function to clean build artifacts
clean() {
    print_info "Cleaning build artifacts..."
    
    for service in "${SERVICES[@]}"; do
        if [ -d "$service/bin" ]; then
            rm -rf "$service/bin"
            print_info "Cleaned $service/bin"
        fi
    done
    
    print_success "Clean completed"
}

# Function to show help
show_help() {
    echo "Usage: $0 [COMMAND] [SERVICE]"
    echo ""
    echo "Commands:"
    echo "  build [SERVICE]  Build all services or a specific service"
    echo "  clean           Clean all build artifacts"
    echo "  help            Show this help message"
    echo ""
    echo "Available services:"
    for service in "${SERVICES[@]}"; do
        echo "  - $(basename "$service")"
    done
    echo ""
    echo "Examples:"
    echo "  $0                    # Build all services"
    echo "  $0 build              # Build all services"
    echo "  $0 build api-gateway  # Build only api-gateway"
    echo "  $0 clean              # Clean all build artifacts"
    echo ""
    echo "Environment variables:"
    echo "  GOOS    Target operating system (default: linux)"
    echo "  GOARCH  Target architecture (default: amd64)"
}

# Main script logic
main() {
    # Store current directory
    local root_dir=$(pwd)
    
    # Ensure we're in the project root
    if [ ! -f "go.mod" ]; then
        print_error "This script must be run from the project root directory"
        exit 1
    fi
    
    local command="${1:-build}"
    local service="$2"
    
    case "$command" in
        "build")
            if [ -n "$service" ]; then
                # Find the service path
                local service_path=""
                for s in "${SERVICES[@]}"; do
                    if [ "$(basename "$s")" = "$service" ]; then
                        service_path="$s"
                        break
                    fi
                done
                
                if [ -z "$service_path" ]; then
                    print_error "Service '$service' not found"
                    echo "Available services:"
                    for s in "${SERVICES[@]}"; do
                        echo "  - $(basename "$s")"
                    done
                    exit 1
                fi
                
                build_service "$service_path"
            else
                build_all
            fi
            ;;
        "clean")
            clean
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            # If first argument is not a command, treat it as a service name
            if [ -n "$1" ]; then
                local service_path=""
                for s in "${SERVICES[@]}"; do
                    if [ "$(basename "$s")" = "$1" ]; then
                        service_path="$s"
                        break
                    fi
                done
                
                if [ -n "$service_path" ]; then
                    build_service "$service_path"
                else
                    print_error "Unknown command or service: $1"
                    show_help
                    exit 1
                fi
            else
                build_all
            fi
            ;;
    esac
}

# Run main function with all arguments
main "$@"