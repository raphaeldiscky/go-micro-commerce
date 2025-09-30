FROM ghcr.io/apollographql/router:v1.59.1

# Copy router configuration
COPY router.yaml /app/router.yaml

# Copy supergraph schema (will be generated before build)
COPY supergraph-schema.graphql /app/supergraph-schema.graphql

# Expose router port
EXPOSE 4000

# Expose health check port
EXPOSE 8088

# Expose Prometheus metrics port
EXPOSE 9091

# Set working directory
WORKDIR /app

# Start the router
CMD ["-c", "/app/router.yaml", "-s", "/app/supergraph-schema.graphql"]