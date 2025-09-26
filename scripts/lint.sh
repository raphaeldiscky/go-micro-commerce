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

lint_service() {
  local dir="$1"
  echo "Linting $dir..."
  (
    cd "$dir" && golangci-lint run ./... --fix --timeout 5m
  )
  echo "Lint completed for $dir"
}

if [ -n "${1-}" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    lint_service "$1"
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
else
  pids=()

  for service in "${SERVICES[@]}"; do
    # run in background
    lint_service "$service" &
    pids+=($!)
  done

  # wait for all jobs to finish
  for pid in "${pids[@]}"; do
    wait "$pid" || {
      echo "Lint failed in one of the services."
      exit 1
    }
  done

  echo "All lint checks completed successfully!"
fi
