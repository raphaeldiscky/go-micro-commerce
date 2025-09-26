#!/bin/bash

set -euo pipefail

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

if [ -n "${1-}" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    format_service "$1"
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
else
  pids=()

  for service in "${SERVICES[@]}"; do
    format_service "$service" &
    pids+=($!)
  done

  for pid in "${pids[@]}"; do
    wait "$pid" || {
      echo "Formatting failed in one of the services."
      exit 1
    }
  done

  echo "All format checks completed successfully!"
fi
