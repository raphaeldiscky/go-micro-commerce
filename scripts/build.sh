#!/bin/bash

set -e

SERVICES=()

for dir in */ ; do
  dir="${dir%/}"  
  if [[ -f "$dir/Dockerfile" ]]; then
    SERVICES+=("$dir")
  fi
done

# Configuration
REGISTRY="${REGISTRY:-localhost:5000}"
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
    echo -e "${BLUE} $1${NC}"
}

print_success() {
    echo -e "${GREEN} $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}  $1${NC}"
}

print_error() {
    echo -e "${RED} $1${NC}"
}

build_image() {
  local service=$1
  local dockerfile="$CURDIR/$service/Dockerfile"
  local image_name="${REGISTRY}/${service}:${TAG}"

  if [ ! -f "$dockerfile" ]; then
    print_warning "Skipping $service: $dockerfile not found"
    return 1
  fi

  print_status "Building Docker image for $service..."
  
  # Build the Docker image using root context (like CI does)
  # This matches the CI workflow: context: . and file: ./${{ matrix.service }}/Dockerfile
  if docker build \
    --tag "$image_name" \
    --file "$dockerfile" \
    "$CURDIR"; then
    print_success "Successfully built image: $image_name"
  else
    print_error "failed to build image for $service"
    return 1
  fi
  
  return 0
}

show_usage() {
  echo "Usage: $0 [OPTIONS] [SERVICE]"
  echo ""
  echo "Build Docker images for microservices (local development)"
  echo ""
  echo "OPTIONS:"
  echo "  -r, --registry REGISTRY   Docker registry (default: localhost:5000)"
  echo "  -t, --tag TAG            Image tag (default: latest)"
  echo "  -h, --help               Show this help message"
  echo ""
  echo "SERVICES:"
  for service in "${SERVICES[@]}"; do
    echo "  - $service"
  done
  echo ""
  echo "EXAMPLES:"
  echo "  $0                       # Build all services"
  echo "  $0 api-gateway          # Build only api-gateway"
  echo "  $0 -t dev               # Build all with tag dev"
  echo ""
  echo "NOTE: This script only builds images locally. Use push.sh to build and push to registry."
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

print_status "Using registry: $REGISTRY"
print_status "Using tag: $TAG"
echo ""

# Single service build
if [ -n "$SERVICE_NAME" ]; then
  print_status "Building image for service: $SERVICE_NAME"
  if build_image "$SERVICE_NAME"; then
    print_success "Image build completed successfully for $SERVICE_NAME"
    echo ""
    print_status "Image available locally: ${REGISTRY}/${SERVICE_NAME}:${TAG}"
  else
    print_error "failed to build image for $SERVICE_NAME"
    exit 1
  fi
else
  print_status "Building Docker images for all services..."
  failed_services=()
  successful_services=()
  
  for service in "${SERVICES[@]}"; do
    echo ""
    print_status "Processing service: $service"
    if build_image "$service"; then
      successful_services+=("$service")
    else
      failed_services+=("$service")
    fi
  done
  
  echo ""
  echo "==================== SUMMARY ===================="
  
  if [ ${#successful_services[@]} -gt 0 ]; then
    print_success "Successfully built ${#successful_services[@]} service(s):"
    for service in "${successful_services[@]}"; do
      echo " ${REGISTRY}/${service}:${TAG}"
    done
  fi
  
  if [ ${#failed_services[@]} -gt 0 ]; then
    echo ""
    print_error "failed to build ${#failed_services[@]} service(s):"
    for service in "${failed_services[@]}"; do
      echo " $service"
    done
    exit 1
  else
    echo ""
    print_success "All Docker images built successfully!"
  fi
fi

# Show built images
echo ""
print_status "Current images:"
docker images | grep -E "(REPOSITORY|${REGISTRY})" || echo "No images found with registry ${REGISTRY}"