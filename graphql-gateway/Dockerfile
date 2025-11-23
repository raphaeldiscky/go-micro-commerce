# ---- Build Stage ----
# Use Alpine for Rover CLI installation and supergraph composition
FROM alpine:latest AS builder

# Install dependencies
RUN apk add --no-cache \
    curl \
    ca-certificates

# Install Rover CLI
SHELL ["/bin/ash", "-o", "pipefail", "-c"]
RUN curl -sSL https://rover.apollo.dev/nix/latest | sh

# Add Rover to PATH
ENV PATH="/root/.rover/bin:${PATH}"

# Set working directory
WORKDIR /build

# Copy supergraph configuration
COPY graphql-gateway/supergraph.yaml ./supergraph.yaml
COPY graphql-gateway/router.yaml ./router.yaml
COPY graphql-gateway/main.rhai ./main.rhai

# Accept Apollo ELv2 license
ENV APOLLO_ELV2_LICENSE=accept

# Note: Supergraph composition requires subgraph services to be running
# In production, you should either:
# 1. Pre-compose the schema and copy it here, OR
# 2. Use schema files instead of introspection URLs
# For now, we'll copy a pre-generated schema if it exists
COPY graphql-gateway/supergraph-schema.graphql ./supergraph-schema.graphql

# ---- Final Stage ----
FROM ghcr.io/apollographql/router:v2.7.0

# Set working directory
WORKDIR /app

# Copy router configuration and supergraph schema from builder
COPY --from=builder /build/router.yaml /app/router.yaml
COPY --from=builder /build/main.rhai /app/main.rhai
COPY --from=builder /build/supergraph-schema.graphql /app/supergraph-schema.graphql

# Expose router port
EXPOSE 4000

# Expose health check port
EXPOSE 4088

# Expose Prometheus metrics port
EXPOSE 9091

# The router will use the environment variables:
# APOLLO_ROUTER_SUPERGRAPH_PATH and APOLLO_ROUTER_CONFIG_PATH
# No CMD needed - will use default entrypoint