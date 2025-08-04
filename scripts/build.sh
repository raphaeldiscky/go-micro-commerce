#!/bin/bash

set -e

SERVICES=(
  "auth-service"
  "notification-service"
  "order-service"
  "product-service"
  "api-gateway"
)
CURDIR=$(pwd)

build_service() {
  local service=$1
  local service_dir="$CURDIR/$service"
  local main_file="$service_dir/cmd/main.go"
  local service_bin_dir="$service_dir/bin"
  local output_file="$service_bin_dir/main"

  if [ ! -f "$main_file" ]; then
    echo "Skipping $service: $main_file not found"
    return
  fi

  echo "Building $service..."
  mkdir -p "$service_bin_dir"
  cd "$service_dir"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-w -s" \
    -tags "sonic avx" \
    -v -o "$output_file" "./cmd/main.go"
  cd "$CURDIR"
  echo "Built $output_file"
}

# Single service build
if [ -n "$1" ]; then
  build_service "$1"
else
  echo "Building all services..."
  for service in "${SERVICES[@]}"; do
    build_service "$service"
  done
  echo "All builds complete."
fi
