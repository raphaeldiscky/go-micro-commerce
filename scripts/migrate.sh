#!/bin/sh

set -e

COMMAND=$1                       # "up", "down", or "force"
SERVICE=${2:-"all"}              # default to all
VERSION=$3                       # version for force command (required)

show_usage() {
  echo "Usage: $0 <command> [service] [version]"
  echo ""
  echo "Commands:"
  echo "  up     - Run migrations up"
  echo "  down   - Run migrations down"
  echo "  force  - Force set migration version (requires version parameter)"
  echo ""
  echo "Services:"
  echo "  all            - Run on all services (default)"
  echo "  auth-service   - Run on auth service only"
  echo "  product-service- Run on product service only"
  echo "  order-service  - Run on order service only"
  echo ""
  echo "Examples:"
  echo "  $0 up                           # Run up migrations for all services"
  echo "  $0 down product-service         # Run down migrations for product service"
  echo "  $0 force order-service 000001   # Force set order service to version 000001"
  echo "  $0 force all 000002             # Force set all services to version 000002"
}

# Validate command
if [ "$COMMAND" != "up" ] && [ "$COMMAND" != "down" ] && [ "$COMMAND" != "force" ]; then
  echo "Invalid command: '$COMMAND'. Use 'up', 'down', or 'force'."
  show_usage
  exit 1
fi

# Check if force command has version parameter
if [ "$COMMAND" = "force" ] && [ -z "$VERSION" ]; then
  echo "Error: 'force' command requires a version parameter."
  show_usage
  exit 1
fi

# Default environment variables
DB_HOST="${DB_HOST:-localhost}"
DB_USER="${DB_USER:-postgres}"
DB_PASS="${DB_PASS:-postgres}"
DB_NAME="${DB_NAME:-postgres}"
SSL_MODE="${SSL_MODE:-disable}"

# Postgres path
POSTGRES_MIGRATION_PATH="db/migrations"

# Function to get port for service
get_service_port() {
  case "$1" in
    "auth-service") echo "15432" ;;
    "product-service") echo "25432" ;;
    "order-service") echo "35432" ;;
    "payment-service") echo "45432" ;;
    "fulfillment-service") echo "55432" ;;
    "notification-service") echo "65432" ;;
    *) echo "" ;;
  esac
}

# Function to get all services
get_all_services() {
  echo "auth-service product-service order-service payment-service fulfillment-service notification-service" 
}

# Function to run migration for a single service
run_migration_for_service() {
  svc=$1
  port=$(get_service_port "$svc")
  
  if [ -z "$port" ]; then
    echo "Unknown service: '$svc'"
    echo "Available services: $(get_all_services)"
    return 1
  fi

  migrations_dir="${svc}/${POSTGRES_MIGRATION_PATH}"
  if [ ! -d "$migrations_dir" ]; then
    echo "Migrations directory '$migrations_dir' does not exist for service '$svc'."
    echo "No migration directory found for $svc — skipping"
    return 0
  fi

  database_url="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${port}/${DB_NAME}?sslmode=${SSL_MODE}"

  echo "----------------------------------------"
  echo "Running '$COMMAND' migrations for $svc"
  echo "Migrations directory: $migrations_dir"
  echo "Database URL: postgres://${DB_USER}:***@${DB_HOST}:${port}/${DB_NAME}?sslmode=${SSL_MODE}"
  
  case "$COMMAND" in
    "up")
      migrate -path "$migrations_dir" -database "$database_url" up
      ;;
    "down")
      migrate -path "$migrations_dir" -database "$database_url" down
      ;;
    "force")
      echo "Forcing migration version to: $VERSION"
      echo "WARNING: This will mark the database as being at version $VERSION without running migrations!"
      printf "Are you sure you want to continue? (y/N): "
      read -r reply
      case "$reply" in
        [Yy]|[Yy][Ee][Ss]) ;;
        *) echo "Aborted."; return 0 ;;
      esac
      migrate -path "$migrations_dir" -database "$database_url" force "$VERSION"
      ;;
  esac
  
  echo "Migration completed for $svc"
  echo "----------------------------------------"
}

# If SERVICE == "all", loop through all services
if [ "$SERVICE" = "all" ]; then
  echo "Running '$COMMAND' migrations for all services..."
  
  # For force command on all services, confirm once
  if [ "$COMMAND" = "force" ]; then
    echo "WARNING: This will force ALL services to version $VERSION!"
    printf "Are you sure you want to continue? (y/N): "
    read -r reply
    case "$reply" in
      [Yy]|[Yy][Ee][Ss]) ;;
      *) echo "Aborted."; exit 0 ;;
    esac
  fi
  
  for svc in $(get_all_services); do
    run_migration_for_service "$svc"
  done
  
  echo "All migrations completed successfully!"
  exit 0
fi

# Run migration for single service
run_migration_for_service "$SERVICE"

echo " Migration completed successfully!"