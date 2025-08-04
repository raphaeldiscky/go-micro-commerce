#!/bin/bash

set -e

COMMAND=$1                       # "up" or "down"
SERVICE=${2:-"all"}              # default to all

if [[ "$COMMAND" != "up" && "$COMMAND" != "down" ]]; then
  echo "Invalid command: '$COMMAND'. Use 'up' or 'down'."
  exit 1
fi

# Default environment variables
DB_HOST="${DB_HOST:-localhost}"
DB_USER="${DB_USER:-postgres}"
DB_PASS="${DB_PASS:-postgres}"
DB_NAME="${DB_NAME:-postgres}"
SSL_MODE="${SSL_MODE:-disable}"

# Map services to host ports (from Docker Compose)
declare -A SERVICE_PORTS
SERVICE_PORTS[product-service]=15432
SERVICE_PORTS[order-service]=25433
# Add more services if needed

# If SERVICE == "all", loop through all
if [ "$SERVICE" == "all" ]; then
  echo "📦 Running '$COMMAND' migrations for all services..."
  for svc in "${!SERVICE_PORTS[@]}"; do
    MIGRATIONS_DIR="${svc}/db/migrations"
    if [ ! -d "$MIGRATIONS_DIR" ]; then
      echo "Skipping $svc — no migrations found at $MIGRATIONS_DIR"
      continue
    fi
    bash "$0" "$COMMAND" "$svc"
  done
  exit 0
fi

PORT=${SERVICE_PORTS[$SERVICE]}

if [ -z "$PORT" ]; then
  echo "❌ Unknown service: '$SERVICE'"
  echo "Available services: ${!SERVICE_PORTS[@]}"
  exit 1
fi

MIGRATIONS_DIR="${SERVICE}/db/migrations"
if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo " No migration directory found for $SERVICE — skipping"
  exit 0
fi

DATABASE_URL="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${PORT}/${DB_NAME}?sslmode=${SSL_MODE}"

echo "Running '$COMMAND' migrations for $SERVICE"
echo "Migrations directory: $MIGRATIONS_DIR"
echo "Database URL: $DATABASE_URL"

migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" "$COMMAND"
