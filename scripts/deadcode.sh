#!/bin/bash

set -euo pipefail

SERVICES=()

for dir in */ ; do
  dir="${dir%/}"
  if [[ -f "$dir/go.mod" ]]; then
    SERVICES+=("$dir")
  fi
done

MAX_CONCURRENT=2

# --- Colored output helpers ---
RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
RESET="\033[0m"

deadcode_service() {
  local dir="$1"
  echo -e "${YELLOW}Running deadcode analysis on $dir...${RESET}"

  if [ -f "$dir/cmd/api/main.go" ]; then
    (cd "$dir" && {
      output=$(deadcode ./cmd/api 2>&1 || true)

      if [[ -n "$output" ]]; then
        echo -e "${RED}❌ Deadcode found in $dir:${RESET}"
        echo "$output"

        # Only fail in CI (GitHub Actions sets CI=true automatically)
        if [[ "${CI:-}" == "true" ]]; then
          exit 1
        fi
      else
        echo -e "${GREEN}✅ No deadcode in $dir${RESET}"
      fi
    })
  else
    echo "Skipping $dir - no cmd/api/main.go found"
  fi
}

# --- Check and install dependency once at the start ---
if ! command -v deadcode &> /dev/null; then
  echo "Installing deadcode..."
  go install golang.org/x/tools/cmd/deadcode@latest
fi

# --- If a specific service is provided as an argument ---
if [ -n "${1-}" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    deadcode_service "$1"
    echo "Deadcode analysis completed for $1."
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
else
  # --- Run for all services concurrently ---
  pids=()
  exit_code=0

  for service in "${SERVICES[@]}"; do
    # Limit concurrency
    if [ ${#pids[@]} -ge $MAX_CONCURRENT ]; then
      if ! wait -n; then
        exit_code=1
      fi
      pids=($(jobs -pr))
    fi

    deadcode_service "$service" &
    pids+=($!)
  done

  echo "Waiting for remaining deadcode checks to finish..."
  for pid in "${pids[@]}"; do
    if ! wait "$pid"; then
      exit_code=1
    fi
  done

  if [ "$exit_code" -ne 0 ]; then
    echo -e "${RED}Deadcode analysis failed in one or more services.${RESET}"
    exit 1
  fi

  echo -e "${GREEN}All deadcode checks completed successfully!${RESET}"
fi

exit 0