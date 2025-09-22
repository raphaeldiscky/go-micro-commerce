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

# Install govulncheck if not present
if ! command -v govulncheck &> /dev/null; then
  echo "Installing govulncheck..."
  go install golang.org/x/vuln/cmd/govulncheck@latest
fi

security_scan() {
  local dir="$1"
  echo "Running security scan for $dir..."
  if [ -d "$dir" ] && [ -f "$dir/go.mod" ]; then
    (cd "$dir" && govulncheck ./...)
    echo "Security scan completed for $dir"
  else
    echo "Skipping $dir - no go.mod found"
  fi
}

if [ -n "$1" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    security_scan "$1"
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
else
  for service in "${SERVICES[@]}"; do
    security_scan "$service"
  done
fi
