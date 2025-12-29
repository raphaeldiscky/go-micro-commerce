# Apollo Router with pre-composed supergraph schema
# Note: The supergraph schema must be pre-generated using:
#   task supergraph
# This runs scripts/compose-supergraph.sh which uses Rover CLI locally
FROM ghcr.io/apollographql/router:v2.7.0

WORKDIR /app

# Copy router configuration and pre-generated supergraph schema
COPY graphql-gateway/router.yaml /app/router.yaml
COPY graphql-gateway/main.rhai /app/main.rhai
COPY graphql-gateway/supergraph-schema.graphql /app/supergraph-schema.graphql

# Expose router port
EXPOSE 4000

# Expose health check port
EXPOSE 4088

# Expose Prometheus metrics port
EXPOSE 9091

# The router will use the environment variables:
# APOLLO_ROUTER_SUPERGRAPH_PATH and APOLLO_ROUTER_CONFIG_PATH
# No CMD needed - will use default entrypoint
