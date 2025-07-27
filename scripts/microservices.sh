#!/bin/bash

# Microservices Setup Script
# This script helps set up and run the Go DDD microservices marketplace

set -e

echo "🚀 Go DDD Microservices Marketplace Setup"
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
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

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_success "Docker and Docker Compose are installed"
}

# Check if Go is installed (for local development)
check_go() {
    if ! command -v go &> /dev/null; then
        print_warning "Go is not installed. You'll need Go 1.23+ for local development."
    else
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go $GO_VERSION is installed"
    fi
}

# Build services locally
build_services() {
    print_status "Building services locally..."
    
    # Build Product Service
    print_status "Building Product Service..."
    cd services/product-service
    go mod tidy
    go build -o ../../bin/product-service ./cmd/main.go
    cd ../..
    
    # Build Seller Service
    print_status "Building Seller Service..."
    cd services/seller-service
    go mod tidy
    go build -o ../../bin/seller-service ./cmd/main.go
    cd ../..
    
    # Build API Gateway
    print_status "Building API Gateway..."
    go mod tidy
    go build -o bin/api-gateway ./cmd/api-gateway/main.go
    
    print_success "All services built successfully"
}

# Start microservices with Docker Compose
start_microservices() {
    print_status "Starting microservices with Docker Compose..."
    
    # Create bin directory if it doesn't exist
    mkdir -p bin
    
    # Stop any running containers
    docker-compose -f docker-compose.microservices.yml down --remove-orphans 2>/dev/null || true
    
    # Build and start services
    docker-compose -f docker-compose.microservices.yml up --build -d
    
    print_success "Microservices started successfully"
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 10
    
    # Check service health
    check_service_health
}

# Check service health
check_service_health() {
    print_status "Checking service health..."
    
    # Check API Gateway
    if curl -s http://localhost:8000/health > /dev/null; then
        print_success "API Gateway is healthy (http://localhost:8000)"
    else
        print_warning "API Gateway is not responding"
    fi
    
    # Check Product Service
    if curl -s http://localhost:8080/health > /dev/null; then
        print_success "Product Service is healthy (http://localhost:8080)"
    else
        print_warning "Product Service is not responding"
    fi
    
    # Check Seller Service
    if curl -s http://localhost:8081/health > /dev/null; then
        print_success "Seller Service is healthy (http://localhost:8081)"
    else
        print_warning "Seller Service is not responding"
    fi
}

# Show service information
show_service_info() {
    echo ""
    echo "🌐 Service URLs:"
    echo "==============="
    echo "📡 API Gateway:    http://localhost:8000"
    echo "🛍️  Product Service: http://localhost:8080"
    echo "👤 Seller Service:  http://localhost:8081"
    echo "📊 Kafka UI:       http://localhost:9090"
    echo ""
    echo "💾 Databases:"
    echo "============="
    echo "🗄️  Product DB:    localhost:5432 (marketplace_products)"
    echo "🗄️  Seller DB:     localhost:5433 (marketplace_sellers)"
    echo "🔴 Redis:          localhost:6379"
    echo ""
    echo "📡 Example API calls:"
    echo "===================="
    echo "# Health check"
    echo "curl http://localhost:8000/health"
    echo ""
    echo "# Create a seller"
    echo "curl -X POST http://localhost:8000/api/v1/sellers \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -d '{\"name\":\"John Doe\",\"email\":\"john@example.com\",\"phone\":\"1234567890\",\"address\":\"123 Main St\"}'"
    echo ""
    echo "# Get marketplace stats"
    echo "curl http://localhost:8000/api/v1/marketplace/stats"
    echo ""
}

# Stop microservices
stop_microservices() {
    print_status "Stopping microservices..."
    docker-compose -f docker-compose.microservices.yml down
    print_success "Microservices stopped"
}

# Show logs
show_logs() {
    SERVICE=${1:-""}
    if [ -z "$SERVICE" ]; then
        docker-compose -f docker-compose.microservices.yml logs -f
    else
        docker-compose -f docker-compose.microservices.yml logs -f "$SERVICE"
    fi
}

# Main script
case "${1:-start}" in
    "start")
        check_docker
        start_microservices
        show_service_info
        ;;
    "stop")
        stop_microservices
        ;;
    "restart")
        stop_microservices
        start_microservices
        show_service_info
        ;;
    "build")
        check_go
        build_services
        ;;
    "health")
        check_service_health
        ;;
    "logs")
        show_logs "${2:-}"
        ;;
    "info")
        show_service_info
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  start     Start all microservices (default)"
        echo "  stop      Stop all microservices"
        echo "  restart   Restart all microservices"
        echo "  build     Build services locally"
        echo "  health    Check service health"
        echo "  logs      Show logs for all services"
        echo "  logs <service>  Show logs for specific service"
        echo "  info      Show service information"
        echo "  help      Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0 start          # Start all services"
        echo "  $0 logs           # Show all logs"
        echo "  $0 logs api-gateway  # Show API Gateway logs"
        echo "  $0 health         # Check service health"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
