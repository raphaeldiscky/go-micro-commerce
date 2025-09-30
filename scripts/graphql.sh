#!/bin/bash

set -euo pipefail

SERVICES=(
  "auth-service"
  "notification-service"
  "order-service"
  "product-service"
  "payment-service"
  "fulfillment-service"
  "search-service"
  "chat-service"
)

generate_graphql() {
  local service="$1"
  local dir="$service"

  echo "Generating GraphQL files for $service..."

  if [ -f "$dir/go.mod" ] && [ -f "$dir/gqlgen.yml" ]; then
    (
      cd "$dir"
      echo "Running gqlgen in $(pwd)..."

      echo "Installing gqlgen dependencies..."
      printf '//go:build tools\npackage tools\nimport (_ "github.com/99designs/gqlgen"\n _ "github.com/99designs/gqlgen/graphql/introspection")' | gofmt > tools.go
      go mod tidy

      echo "Generating GraphQL code..."
      go run github.com/99designs/gqlgen generate --verbose

      echo "Remove tools.go"
      rm tools.go
    )
  else
    echo "Skipping $service (missing go.mod or gqlgen.yml)."
  fi
}

# If a specific service is provided as an argument
if [ -n "${1-}" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    generate_graphql "$1"
    echo "Done for $1."
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
# If no argument, run for all services concurrently
else
  pids=()
  for service in "${SERVICES[@]}"; do
    generate_graphql "$service" &
    pids+=($!)
  done

  exit_code=0
  for pid in "${pids[@]}"; do
    if ! wait "$pid"; then
      echo "A gqlgen process failed."
      exit_code=1
    fi
  done

  if [ "$exit_code" -ne 0 ]; then
    echo "GraphQL codegen failed in one or more services."
    exit 1
  fi

  echo "GraphQL codegen completed successfully for all services!"
fi
