#!/bin/bash

set -euo pipefail

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
  "graphql-gateway"
)

MAX_CONCURRENT=4

deadcode_service() {
  local dir="$1"
  echo "Running deadcode analysis on $dir..."

  # Skip Node.js projects (no Go code)
  if [ -f "$dir/package.json" ]; then
    echo "Skipping $dir - node project (no Go code to analyze)"
    return 0
  fi

  # Run deadcode from cmd/api main package
  if [ -f "$dir/cmd/api/main.go" ]; then
    # Run in a subshell to isolate the cd command
    (cd "$dir" && deadcode ./cmd/api)
  else
    echo "Skipping $dir - no cmd/api/main.go found"
  fi
}

# Check and install dependency once at the start ---
if ! command -v deadcode &> /dev/null; then
  echo "Installing deadcode..."
  go install golang.org/x/tools/cmd/deadcode@latest
fi

# If a specific service is provided as an argument
if [ -n "${1-}" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    deadcode_service "$1"
    echo "Deadcode analysis completed for $1."
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
# If no argument, run for all services concurrently
else
  pids=()
  exit_code=0 # Variable to track if any process fails

  for service in "${SERVICES[@]}"; do
    # If the number of running jobs has reached the limit...
    if [ ${#pids[@]} -ge $MAX_CONCURRENT ]; then
      # ...wait for the *next* job to finish
      # Handle failures from finished jobs ---
      if ! wait -n; then
        echo "A deadcode process failed."
        exit_code=1
      fi
      # Remove all finished PIDs from the list
      pids=($(jobs -pr))
    fi

    echo "Starting deadcode analysis for $service..."
    deadcode_service "$service" &
    pids+=($!)
  done

  # Wait for all remaining jobs and check for failures ---
  echo "Waiting for remaining jobs to finish..."
  for pid in "${pids[@]}"; do
    if ! wait "$pid"; then
      echo "A deadcode process failed (PID: $pid)."
      exit_code=1
    fi
  done

  # After checking all processes, decide the final outcome
  if [ "$exit_code" -ne 0 ]; then
    echo "Deadcode analysis failed in one or more services."
    exit 1
  fi

  echo "All deadcode checks completed successfully!"
fi