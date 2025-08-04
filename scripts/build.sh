#!/bin/bash

set -e

SERVICES=(
  "auth-service"
  "notification-service"
  "order-service"
  "product-service"
  "pkg"
  "api-gateway"
)
CURDIR=$(pwd)
OUTPUT_DIR="$CURDIR/bin"

mkdir -p "$OUTPUT_DIR"

build_service() {
  local service=$1
  local main_file="$CURDIR/$service/cmd/main.go"
  local output_file="$OUTPUT_DIR/$service"

  if [ ! -f "$main_file" ]; then
    echo "Skipping $service: $main_file not found"
    return
  fi

  echo "Building $service..."
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-w -s" \
    -tags "sonic avx" \
    -v -o "$output_file" "$main_file"
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
