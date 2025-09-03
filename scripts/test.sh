#!/bin/bash

set -e

SERVICE=${1:-"all"}  # default to all

# Services with Go code
SERVICES=(
  "auth-service"
  "product-service"
  "order-service"
  "notification-service"
  "payment-service"
  "fulfillment-service"
  "api-gateway"
  "pkg"
)

run_tests() {
  local dir="$1"

  if [ ! -f "$dir/go.mod" ]; then
    echo "Skipping $dir — no go.mod found"
    return
  fi

  if ! find "$dir" -name "*_test.go" | grep -q .; then
    echo "Skipping $dir — no test files found"
    return
  fi

  echo "Running tests in $dir..."
  (cd "$dir" && go test ./... -v -coverprofile=coverage.out -covermode=atomic)
  echo "Tests completed for $dir"

  echo "Coverage report for $dir:"
  (cd "$dir" && go tool cover -func=coverage.out)
  echo
}

if [ "$SERVICE" == "all" ]; then
  for svc in "${SERVICES[@]}"; do
    run_tests "$svc"
  done
else
  run_tests "$SERVICE"
fi
