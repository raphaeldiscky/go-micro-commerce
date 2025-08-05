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

lint_pkg() {
  echo "Linting shared package (pkg)..."
  (cd "pkg" && golangci-lint run ./... --fix --timeout 5m --config ../.golangci.yml)
  echo "Lint completed for pkg"
}


lint_service() {
  local dir="$1"
  echo "Linting $dir..."
  (cd "$dir" && golangci-lint run ./... --fix --timeout 5m --config ../.golangci.yml)
  echo "Lint completed for $dir"
}

lint_pkg

if [ -n "$1" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    lint_service "$1"
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
else
  for service in "${SERVICES[@]}"; do
    lint_service "$service"
  done
fi
