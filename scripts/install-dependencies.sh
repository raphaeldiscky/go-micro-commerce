#!/bin/bash

set -e

SERVICES=()

for dir in */ ; do
  dir="${dir%/}"  
  if [[ -f "$dir/go.mod" ]]; then
    SERVICES+=("$dir")
done

install_deps() {
  local dir="$1"
  echo "Installing dependencies in $dir..."
  (cd "$dir" && go mod tidy)
  echo "Done for $dir"
}

# If an argument is passed, install only for that service
if [ -n "$1" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    install_deps "$1"
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
else
  # No argument passed — install for all services
  for service in "${SERVICES[@]}"; do
    install_deps "$service"
  done
fi
