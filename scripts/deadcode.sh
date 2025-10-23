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

# Load ignore patterns from .deadcodeignore file
load_ignore_patterns() {
  local ignore_file=".deadcodeignore"

  if [[ -f "$ignore_file" ]]; then
    # Read patterns, skip empty lines and comments
    grep -v '^#' "$ignore_file" | grep -v '^[[:space:]]*$' || true
  fi
}

# Filter deadcode output based on ignore patterns
filter_ignored_patterns() {
  local output="$1"
  local patterns="$2"

  if [[ -z "$patterns" ]]; then
    # No patterns to filter, return original output
    echo "$output"
    return
  fi

  # Filter out each pattern
  local filtered_output="$output"
  while IFS= read -r pattern; do
    if [[ -n "$pattern" ]]; then
      # Use grep -v -F for fixed-string matching to filter out the pattern
      filtered_output=$(echo "$filtered_output" | grep -v -F "$pattern" || true)
    fi
  done <<< "$patterns"

  echo "$filtered_output"
}

deadcode_service() {
  local dir="$1"
  local ignore_patterns="$2"
  echo -e "${YELLOW}Running deadcode analysis on $dir...${RESET}"

  if [ -f "$dir/cmd/api/main.go" ]; then
    (cd "$dir" && {
      output=$(deadcode --test ./cmd/api 2>&1 || true)

      # Track ignored items
      local ignored_items=""
      if [[ -n "$output" && -n "$ignore_patterns" ]]; then
        while IFS= read -r pattern; do
          if [[ -n "$pattern" ]]; then
            local matched=$(echo "$output" | grep -F "$pattern" || true)
            if [[ -n "$matched" ]]; then
              ignored_items+="$matched"$'\n'
            fi
          fi
        done <<< "$ignore_patterns"
      fi

      # Filter ignored patterns
      filtered_output=$(filter_ignored_patterns "$output" "$ignore_patterns")

      # Show ignored items if any
      if [[ -n "$ignored_items" ]]; then
        echo -e "${YELLOW}ℹ Ignored deadcode (via .deadcodeignore):${RESET}"
        echo -n "$ignored_items"
      fi

      if [[ -n "$filtered_output" ]]; then
        echo -e "${RED}❌ Deadcode found in $dir:${RESET}"
        echo "$filtered_output"

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

# --- Load ignore patterns once at the start ---
IGNORE_PATTERNS=$(load_ignore_patterns)

# --- If a specific service is provided as an argument ---
if [ -n "${1-}" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    deadcode_service "$1" "$IGNORE_PATTERNS"
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

    deadcode_service "$service" "$IGNORE_PATTERNS" &
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