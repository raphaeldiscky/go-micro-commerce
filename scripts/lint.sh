#!/bin/bash

set -euo pipefail

SERVICES=()

for dir in */ ; do
  dir="${dir%/}"  
  if [[ -f "$dir/go.mod" ]]; then
    SERVICES+=("$dir")
  elif [[ -f "$dir/package.json" ]]; then
    SERVICES+=("$dir")
  fi
done

lint_service() {
  local dir="$1"


  # Check if this is a Node.js project
  if [ -f "$dir/package.json" ]; then
    echo "Linting $dir(node project)..."
    (
      cd "$dir"
      if [ -f "pnpm-lock.yaml" ]; then
        pnpm run lint
      elif [ -f "package-lock.json" ]; then
        npm run lint
      else
        echo "Warning: No lock file found, using npm"
        npm run lint
      fi
    )
    return $?
  fi

  echo "Linting $dir..."

  # Go project linting
  (
    cd "$dir" && golangci-lint run ./... --fix --timeout 5m
  )
  # The exit code of the subshell is returned automatically
}

# If a specific service is provided as an argument
if [ -n "${1-}" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    lint_service "$1"
    echo "Lint completed for $1."
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
    lint_service "$service" &
    pids+=($!)
  done
  
  # Variable to track if any process fails
  exit_code=0 

  # Wait for all background jobs to finish
  for pid in "${pids[@]}"; do
    # 'wait $pid' will return the exit code of the process.
    # If a command fails, its exit code will be non-zero.
    if ! wait "$pid"; then
      echo "A lint process failed."
      # Record that at least one failure occurred
      exit_code=1 
    fi
  done

  # After checking all processes, decide the final outcome
  if [ "$exit_code" -ne 0 ]; then
    echo "Linting failed in one or more services."
    exit 1
  fi

  echo "All lint checks completed successfully!"
fi