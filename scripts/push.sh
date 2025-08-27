#!/bin/bash

set -e

SERVICES=(
  "auth-service"
  "notification-service"
  "order-service"
  "product-service"
  "payment-service"
  "api-gateway"
)

# Configuration
REGISTRY="${REGISTRY:-ghcr.io/raphaeldiscky/go-micro-template}"
TAG="${TAG:-latest}"
CURDIR=$(pwd)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_warning() {
    echo -e "${YELLOW}$1${NC}"
}

print_error() {
    echo -e "${RED}$1${NC}"
}

build_and_push_image() {
  local service=$1
  local dockerfile="$CURDIR/$service/Dockerfile"
  local image_name="${REGISTRY}/${service}:${TAG}"

  if [ ! -f "$dockerfile" ]; then
    print_warning "Skipping $service: $dockerfile not found"
    return 1
  fi

  if [ ! -f "$CURDIR/$service/cmd/api/main.go" ]; then
    print_warning "Skipping $service: main.go not found in $CURDIR/$service/cmd/api/"
    return 1
  fi

  print_status "Building Docker image for $service..."
  
  # Build the Docker image using root context
  if docker build \
    --tag "$image_name" \
    --file "$dockerfile" \
    "$CURDIR"; then
    print_success "Successfully built image: $image_name"
  else
    print_error "failed to build image for $service"
    return 1
  fi
  
  # Push to registry
  print_status "Pushing $image_name to registry..."
  if docker push "$image_name"; then
    print_success "Successfully pushed image: $image_name"
  else
    print_error "failed to push image for $service"
    return 1
  fi
  
  return 0
}

check_docker_login() {
  print_status "Checking Docker registry authentication..."
  
  # Check if we can access the registry
  if [[ "$REGISTRY" == ghcr.io* ]]; then
    # For GitHub Container Registry
    if ! docker info | grep -q "Username"; then
      print_warning "Not logged into Docker registry"
      print_status "To login to GitHub Container Registry, run:"
      echo "  export CR_PAT=YOUR_GITHUB_TOKEN"
      echo "  echo \$CR_PAT | docker login ghcr.io -u USERNAME --password-stdin"
      echo ""
      read -p "Continue anyway? (y/N): " -n 1 -r
      echo
      if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
      fi
    else
      print_success "Docker registry authentication verified"
    fi
  else
    print_status "Using custom registry: $REGISTRY"
  fi
}

show_usage() {
  echo "Usage: $0 [OPTIONS] [SERVICE]"
  echo ""
  echo "Build and push Docker images for microservices"
  echo ""
  echo "OPTIONS:"
  echo "  -r, --registry REGISTRY   Docker registry (default: ghcr.io/raphaeldiscky/go-micro-template)"
  echo "  -t, --tag TAG            Image tag (default: latest)"
  echo "  -h, --help               Show this help message"
  echo ""
  echo "SERVICES:"
  for service in "${SERVICES[@]}"; do
    echo "  - $service"
  done
  echo ""
  echo "EXAMPLES:"
  echo "  $0                       # Build and push all services"
  echo "  $0 api-gateway          # Build and push only api-gateway"
  echo "  $0 -t v1.0.0            # Build and push all with tag v1.0.0"
  echo "  $0 -r my-registry.com   # Use custom registry"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -r|--registry)
      REGISTRY="$2"
      shift 2
      ;;
    -t|--tag)
      TAG="$2"
      shift 2
      ;;
    -h|--help)
      show_usage
      exit 0
      ;;
    -*)
      print_error "Unknown option: $1"
      show_usage
      exit 1
      ;;
    *)
      # This is a service name
      SERVICE_NAME="$1"
      shift
      ;;
  esac
done

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
  print_error "Docker is not running or not accessible"
  exit 1
fi

# Check Docker registry authentication
check_docker_login

print_status "Using registry: $REGISTRY"
print_status "Using tag: $TAG"
echo ""

# Single service build and push
if [ -n "$SERVICE_NAME" ]; then
  print_status "Building and pushing image for service: $SERVICE_NAME"
  if build_and_push_image "$SERVICE_NAME"; then
    print_success "Build and push completed successfully for $SERVICE_NAME"
    echo ""
    print_status "Image available at: ${REGISTRY}/${SERVICE_NAME}:${TAG}"
  else
    print_error "failed to build and push image for $SERVICE_NAME"
    exit 1
  fi
else
  print_status "Building and pushing Docker images for all services..."
  failed_services=()
  successful_services=()
  
  for service in "${SERVICES[@]}"; do
    echo ""
    print_status "Processing service: $service"
    if build_and_push_image "$service"; then
      successful_services+=("$service")
    else
      failed_services+=("$service")
    fi
  done
  
  echo ""
  echo "==================== SUMMARY ===================="
  
  if [ ${#successful_services[@]} -gt 0 ]; then
    print_success "Successfully built and pushed ${#successful_services[@]} service(s):"
    for service in "${successful_services[@]}"; do
      echo " ${REGISTRY}/${service}:${TAG}"
    done
  fi
  
  if [ ${#failed_services[@]} -gt 0 ]; then
    echo ""
    print_error "failed to build and push ${#failed_services[@]} service(s):"
    for service in "${failed_services[@]}"; do
      echo " $service"
    done
    exit 1
  else
    echo ""
    print_success "All Docker images built and pushed successfully!"
  fi
fi

echo ""
print_status "Deployment commands:"
echo "  docker pull ${REGISTRY}/api-gateway:${TAG}"
echo "  docker pull ${REGISTRY}/auth-service:${TAG}"
echo "  # ... etc"
echo ""
print_status "To update your docker-compose.yml files, change the image references to:"
echo "  image: ${REGISTRY}/SERVICE_NAME:${TAG}"
