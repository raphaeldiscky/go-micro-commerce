#!/bin/bash

set -e

SERVICES=(
  "auth-service"
  "notification-service"
  "order-service"
  "product-service"
  "api-gateway"
  "payment-service"
  "fulfillment-service"
  "search-service"
  "chat-service"
)

deadcode_service() {
  local dir="$1"
  echo "Running deadcode analysis on $dir..."
  
  # Check if deadcode is installed
  if ! command -v deadcode &> /dev/null; then
    echo "Installing deadcode..."
    go install golang.org/x/tools/cmd/deadcode@latest
  fi
  
  # Run deadcode from cmd/api main package
  if [ -f "$dir/cmd/api/main.go" ]; then
    (cd "$dir" && deadcode ./cmd/api)
  else
    echo "Skipping $dir - no cmd/api/main.go found"
  fi
  echo "Deadcode analysis completed for $dir"
}

if [ -n "$1" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    deadcode_service "$1"
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
else
  for service in "${SERVICES[@]}"; do
    deadcode_service "$service"
  done
fi