#!/bin/bash

set -e

COMMAND=$1
SERVICE_ARG=$2

DB_HOST="localhost"

SERVICES=(
  "auth-service"
  "product-service"
  "order-service"
  "notification-service"
)


get_db_port() {
  case "$1" in
    "auth-service") echo 5000 ;;
    "product-service") echo 5001 ;;
    "order-service") echo 5002 ;;
    "notification-service") echo 5003 ;;
    *) echo "Unknown service '$1'" >&2; exit 1 ;;
  esac
}

run_migration() {
  SERVICE=$1
  DB_PORT=$(get_db_port "$SERVICE")
  MIGRATION_PATH="./services/$SERVICE/internal/infra/db/migrations"

  echo "Migrating $SERVICE on port $DB_PORT..."

  migrate -path "$MIGRATION_PATH" \
    -database "postgresql://postgres:postgres@$DB_HOST:$DB_PORT/api_db?sslmode=disable" \
    -verbose "$COMMAND"
}

if [ "$COMMAND" != "up" ] && [ "$COMMAND" != "down" ]; then
  echo "Usage: ./scripts/migrate.sh [up|down] [optional-service-name]"
  exit 1
fi

if [ -n "$SERVICE_ARG" ]; then
  run_migration "$SERVICE_ARG"
else
  for svc in "${SERVICES[@]}"; do
    run_migration "$svc"
  done
fi
