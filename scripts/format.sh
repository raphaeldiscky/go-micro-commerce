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
  "graphql-gateway"
)

format_service() {
  local dir="$1"

  # Check if this is a Node.js project
  if [ -f "$dir/package.json" ]; then
    echo "Formatting $dir(node project)..."
    (
      cd "$dir"
      if [ -f "pnpm-lock.yaml" ]; then
        pnpm run format
      elif [ -f "package-lock.json" ]; then
        npm run format
      else
        echo "Warning: No lock file found, using npm"
        npm run format
      fi
    )
    return $?
  fi

  echo "Formatting $dir..."


  # Go project formatting
  (
    gofumpt -w "$dir"
    goimports -w "$dir"
  )
  # The exit code of the last command in the subshell is returned
}

# If a specific service is provided as an argument
if [ -n "${1-}" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    format_service "$1"
    echo "Format complete for $1."
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
# If no argument, run for all services concurrently
else
  pids=()
  for service in "${SERVICES[@]}"; do
    # Run in the background and store the process ID
    format_service "$service" &
    pids+=($!)
  done

  # Variable to track if any process fails
  exit_code=0

  # Wait for all background jobs to finish
  for pid in "${pids[@]}"; do
    # 'wait "$pid"' will return the exit code of the process.
    # If a command fails, its exit code will be non-zero.
    if ! wait "$pid"; then
      echo "A format process failed."
      # Record that at least one failure occurred
      exit_code=1
    fi
  done

  # After checking all processes, decide the final outcome
  if [ "$exit_code" -ne 0 ]; then
    echo "Formatting failed in one or more services."
    exit 1
  fi

  echo "All format checks completed successfully!"
fi