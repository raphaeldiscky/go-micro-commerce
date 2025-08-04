#!/bin/bash

set -e

SERVICE=$1

if [ -z "$SERVICE" ]; then
  echo "Please specify a service to run (e.g., product-service)"
  exit 1
fi

AIR_CONFIG="./$SERVICE/.air.toml"

if [ ! -f "$AIR_CONFIG" ]; then
  echo ".air.toml not found in $SERVICE, using default air config"
  cd "$SERVICE"
  air
else
  echo "Starting $SERVICE with air..."
  cd "$SERVICE"
  air -c .air.toml
fi
