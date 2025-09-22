#!/bin/bash

set -e

SERVICES=(
  "auth-service"
  "notification-service"
  "order-service"
  "product-service"
  "pkg"
  "proto"
  "api-gateway"
  "payment-service"
  "fulfillment-service"
  "search-service"
  "chat-service"
)

format_service() {
  local dir="$1"
  echo "Formatting $dir..."
  gofumpt -w "$dir"
  goimports -w "$dir"
  echo "Format complete for $dir"
}

if [ -n "$1" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    format_service "$1"
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
else
  for service in "${SERVICES[@]}"; do
    format_service "$service"
  done
fi
