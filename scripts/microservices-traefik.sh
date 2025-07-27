#!/bin/bash

# Microservices Management Script with Traefik
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.traefik.yml"
PROJECT_NAME="go-ddd"

# Print colored output
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

# Check if Docker and Docker Compose are installed
check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_success "Dependencies check passed"
}

# Start all services with Traefik
start_services() {
    print_status "Starting microservices with Traefik API Gateway..."
    
    # Create network if it doesn't exist
    docker network create ${PROJECT_NAME}_traefik 2>/dev/null || true
    
    # Start services
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d
    
    print_success "All services started!"
    print_status "Waiting for services to be ready..."
    
    # Wait for Traefik to be ready
    sleep 10
    
    show_service_info
}

# Stop all services
stop_services() {
    print_status "Stopping all microservices..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down
    print_success "All services stopped!"
}

# Restart all services
restart_services() {
    print_status "Restarting all microservices..."
    stop_services
    sleep 2
    start_services
}

# Show service status and endpoints
show_service_info() {
    print_status "Service Information:"
    echo
    echo -e "${GREEN}🌐 API Gateway (Traefik)${NC}"
    echo -e "  Dashboard: ${BLUE}http://localhost:8080${NC}"
    echo -e "  API Endpoint: ${BLUE}http://localhost${NC}"
    echo
    echo -e "${GREEN}📦 Product Service${NC}"
    echo -e "  Endpoint: ${BLUE}http://localhost/api/v1/products${NC}"
    echo -e "  Health: ${BLUE}http://localhost/api/v1/products/health${NC}"
    echo
    echo -e "${GREEN}👤 Seller Service${NC}"
    echo -e "  Endpoint: ${BLUE}http://localhost/api/v1/sellers${NC}"
    echo -e "  Health: ${BLUE}http://localhost/api/v1/sellers/health${NC}"
    echo
    echo -e "${GREEN}📊 Monitoring${NC}"
    echo -e "  Prometheus: ${BLUE}http://localhost:9091${NC}"
    echo -e "  Grafana: ${BLUE}http://localhost:3001${NC} (admin/admin)"
    echo
    echo -e "${GREEN}📨 Kafka Management${NC}"
    echo -e "  Kafka UI: ${BLUE}http://localhost:8081${NC}"
    echo
    echo -e "${GREEN}🗄️ Databases${NC}"
    echo -e "  Product DB: ${BLUE}localhost:5433${NC}"
    echo -e "  Seller DB: ${BLUE}localhost:5434${NC}"
    echo -e "  Redis Cache: ${BLUE}localhost:6379${NC}"
}

# Check health of all services
check_health() {
    print_status "Checking service health..."
    
    # Function to check HTTP endpoint
    check_endpoint() {
        local name=$1
        local url=$2
        local expected_code=${3:-200}
        
        if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "$expected_code"; then
            print_success "$name is healthy"
        else
            print_error "$name is not responding"
        fi
    }
    
    echo
    print_status "Health Check Results:"
    echo
    
    # Check Traefik
    check_endpoint "Traefik Dashboard" "http://localhost:8080/api/rawdata"
    
    # Check Product Service
    check_endpoint "Product Service" "http://localhost/api/v1/products/health"
    
    # Check Seller Service
    check_endpoint "Seller Service" "http://localhost/api/v1/sellers/health"
    
    # Check Prometheus
    check_endpoint "Prometheus" "http://localhost:9091/-/healthy"
    
    # Check Grafana
    check_endpoint "Grafana" "http://localhost:3001/api/health"
    
    # Check Kafka UI
    check_endpoint "Kafka UI" "http://localhost:8081/actuator/health"
}

# Show logs for specific service
show_logs() {
    local service=$1
    if [ -z "$service" ]; then
        print_status "Available services:"
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps --services
        echo
        print_status "Usage: $0 logs <service-name>"
        return 1
    fi
    
    print_status "Showing logs for $service..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f "$service"
}

# Clean up everything
cleanup() {
    print_status "Cleaning up all resources..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down -v --remove-orphans
    docker network rm ${PROJECT_NAME}_traefik 2>/dev/null || true
    print_success "Cleanup completed!"
}

# Build services
build_services() {
    print_status "Building microservices..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME build
    print_success "Build completed!"
}

# Scale a service
scale_service() {
    local service=$1
    local replicas=$2
    
    if [ -z "$service" ] || [ -z "$replicas" ]; then
        print_error "Usage: $0 scale <service-name> <replicas>"
        return 1
    fi
    
    print_status "Scaling $service to $replicas replicas..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d --scale "$service=$replicas"
    print_success "$service scaled to $replicas replicas!"
}

# Show usage information
show_help() {
    echo "Microservices Management Script with Traefik"
    echo
    echo "Usage: $0 <command> [options]"
    echo
    echo "Commands:"
    echo "  start       Start all microservices with Traefik"
    echo "  stop        Stop all microservices"
    echo "  restart     Restart all microservices"
    echo "  status      Show service information and endpoints"
    echo "  health      Check health of all services"
    echo "  logs        Show logs for a specific service"
    echo "  build       Build all services"
    echo "  scale       Scale a specific service"
    echo "  cleanup     Remove all containers, volumes, and networks"
    echo "  help        Show this help message"
    echo
    echo "Examples:"
    echo "  $0 start                    # Start all services"
    echo "  $0 logs product-service     # Show product service logs"
    echo "  $0 scale seller-service 3   # Scale seller service to 3 replicas"
    echo "  $0 health                   # Check all service health"
}

# Main script logic
main() {
    case "${1:-}" in
        start)
            check_dependencies
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        status|info)
            show_service_info
            ;;
        health)
            check_health
            ;;
        logs)
            show_logs "$2"
            ;;
        build)
            build_services
            ;;
        scale)
            scale_service "$2" "$3"
            ;;
        cleanup)
            cleanup
            ;;
        help|--help|-h)
            show_help
            ;;
        "")
            print_error "No command specified"
            show_help
            exit 1
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
