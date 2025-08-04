#!/bin/bash

set -e

SERVICES=(
  "auth-service"
  "notification-service"
  "order-service"
  "product-service"
  "api-gateway"
)

# Configuration
REGISTRY="${REGISTRY:-localhost:5000}"
TAG="${TAG:-latest}"
CURDIR=$(pwd)

build_image() {
  local service=$1
  local dockerfile="$CURDIR/$service/Dockerfile"
  local image_name="${REGISTRY}/${service}:${TAG}"

  if [ ! -f "$dockerfile" ]; then
    echo "Skipping $service: $dockerfile not found"
    return 1
  fi

  if [ ! -f "$CURDIR/$service/cmd/main.go" ]; then
    echo "Skipping $service: main.go not found in $CURDIR/$service/cmd/"
    return 1
  fi

  echo "Building Docker image for $service..."
  
  # Build the Docker image using root context (like CI does)
  # This matches the CI workflow: context: . and file: ./${{ matrix.service }}/Dockerfile
  docker build \
    --tag "$image_name" \
    --file "$dockerfile" \
    "$CURDIR"
  
  echo "Successfully built image: $image_name"
  
  # Optional: Push to registry (uncomment if needed)
  # echo "Pushing $image_name to registry..."
  # docker push "$image_name"
  
  return 0
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
  echo "Error: Docker is not running or not accessible"
  exit 1
fi

# Single service build
if [ -n "$1" ]; then
  echo "Building image for service: $1"
  if build_image "$1"; then
    echo "Image build completed successfully for $1"
  else
    echo "Failed to build image for $1"
    exit 1
  fi
else
  echo "Building Docker images for all services..."
  failed_services=()
  
  for service in "${SERVICES[@]}"; do
    if ! build_image "$service"; then
      failed_services+=("$service")
    fi
  done
  
  if [ ${#failed_services[@]} -eq 0 ]; then
    echo "All Docker images built successfully!"
    echo ""
    echo "Built images:"
    for service in "${SERVICES[@]}"; do
      echo "  - ${REGISTRY}/${service}:${TAG}"
    done
  else
    echo ""
    echo "Failed to build images for: ${failed_services[*]}"
    exit 1
  fi
fi

# Show built images
echo ""
echo "Current images:"
docker images | grep -E "(REPOSITORY|${REGISTRY})" || echo "No images found with registry ${REGISTRY}"