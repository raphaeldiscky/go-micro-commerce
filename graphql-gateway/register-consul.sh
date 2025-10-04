#!/bin/sh
#
# Consul registration sidecar for graphql-gateway
# Registers the service with Consul and keeps running to maintain registration

set -e

CONSUL_HTTP_ADDR="${CONSUL_HTTP_ADDR:-consul:8500}"
SERVICE_NAME="${SERVICE_NAME:-graphql-gateway}"
SERVICE_ID="${SERVICE_ID:-graphql-gateway-1}"
SERVICE_ADDRESS="${SERVICE_ADDRESS:-graphql-gateway}"
SERVICE_PORT="${SERVICE_PORT:-4000}"
HEALTH_CHECK_URL="${HEALTH_CHECK_URL:-http://graphql-gateway:4088/health}"

# Deregister on exit
cleanup() {
    echo "Deregistering $SERVICE_NAME from Consul..."
    curl -X PUT "http://$CONSUL_HTTP_ADDR/v1/agent/service/deregister/$SERVICE_ID" || true
    echo "Deregistered successfully"
    exit 0
}

trap cleanup SIGTERM SIGINT EXIT

echo "Starting Consul registration sidecar for $SERVICE_NAME..."

# Wait for Consul to be ready
until curl -sf "http://$CONSUL_HTTP_ADDR/v1/status/leader" > /dev/null 2>&1; do
    echo "Waiting for Consul to be ready..."
    sleep 2
done

echo "Consul is ready. Registering $SERVICE_NAME..."

# Register service with Consul
register_service() {
    curl -sf -X PUT "http://$CONSUL_HTTP_ADDR/v1/agent/service/register" \
      -H "Content-Type: application/json" \
      -d "{
        \"ID\": \"$SERVICE_ID\",
        \"Name\": \"$SERVICE_NAME\",
        \"Tags\": [\"graphql\", \"federation\", \"gateway\"],
        \"Address\": \"$SERVICE_ADDRESS\",
        \"Port\": $SERVICE_PORT,
        \"Check\": {
          \"HTTP\": \"$HEALTH_CHECK_URL\",
          \"Interval\": \"10s\",
          \"Timeout\": \"5s\",
          \"DeregisterCriticalServiceAfter\": \"30s\"
        }
      }" > /dev/null
}

# Initial registration
if register_service; then
    echo "Successfully registered $SERVICE_NAME with Consul"
else
    echo "Failed to register $SERVICE_NAME with Consul"
    exit 1
fi

# Keep the sidecar running and check registration every 30 seconds
echo "Sidecar running. Monitoring registration..."
while true; do
    sleep 30

    # Check if still registered
    if ! curl -sf "http://$CONSUL_HTTP_ADDR/v1/agent/service/$SERVICE_ID" > /dev/null 2>&1; then
        echo "Service not registered. Re-registering..."
        register_service || echo "Re-registration failed, will retry..."
    fi
done
