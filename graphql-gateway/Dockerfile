# ---- Build Stage ----
# Use Ubuntu for Rover CLI installation and supergraph composition
FROM ubuntu:22.04 AS builder

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Rover CLI
RUN curl -sSL https://rover.apollo.dev/nix/latest | sh

# Add Rover to PATH
ENV PATH="/root/.rover/bin:${PATH}"

# Set working directory
WORKDIR /build

# Copy supergraph configuration
COPY graphql-gateway/supergraph.yaml ./supergraph.yaml
COPY graphql-gateway/router.yaml ./router.yaml

# Accept Apollo ELv2 license
ENV APOLLO_ELV2_LICENSE=accept

# Note: Supergraph composition requires subgraph services to be running
# In production, you should either:
# 1. Pre-compose the schema and copy it here, OR
# 2. Use schema files instead of introspection URLs
# For now, we'll copy a pre-generated schema if it exists
COPY graphql-gateway/supergraph-schema.graphql ./supergraph-schema.graphql

# ---- Final Stage ----
FROM ghcr.io/apollographql/router:v1.59.1

# Set working directory
WORKDIR /app

# Copy router configuration and supergraph schema from builder
COPY --from=builder /build/router.yaml /app/router.yaml
COPY --from=builder /build/supergraph-schema.graphql /app/supergraph-schema.graphql

# Expose router port
EXPOSE 4000

# Expose health check port
EXPOSE 8088

# Expose Prometheus metrics port
EXPOSE 9091

# Start the router
CMD ["-c", "/app/router.yaml", "-s", "/app/supergraph-schema.graphql"]