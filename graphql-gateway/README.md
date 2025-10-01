# GraphQL Gateway (Apollo Router)

Apollo Router-based GraphQL Federation gateway for the go-micro-commerce platform.

## Architecture

The GraphQL Gateway composes multiple GraphQL subgraphs into a unified supergraph:

- **auth-service**: Owns User entity, handles authentication
- **chat-service**: Owns Conversation, Message, Participant entities
- (Future: product-service, order-service, etc.)

## Local Development

### Prerequisites

- [Rover CLI](https://www.apollographql.com/docs/rover/getting-started) installed
- Docker and Docker Compose
- Running subgraph services (auth-service, chat-service)

### Supergraph Schema Management

The supergraph schema uses a **commit-based workflow**:

#### Automatic Composition (CI/CD)

When you push changes to subgraph GraphQL schemas (auth-service, chat-service), the CI/CD pipeline automatically:

1. Composes the supergraph schema using Rover
2. Commits the updated `supergraph-schema.graphql` if changed
3. Pushes the commit back to the repository

#### Manual Composition (Local Development)

Before committing subgraph schema changes, **you should compose locally**:

```bash
# Using the script
bash scripts/compose-supergraph.sh

# Or using task
task compose_graphql
```

This generates `supergraph-schema.graphql` from running subgraph services.

**Important**: Always commit the generated schema file along with your subgraph changes.

### Start Gateway

```bash
# Start infrastructure and services first
task start_infra
task start_apps

# Compose supergraph
task compose_graphql

# Start gateway
docker-compose -f deployments/docker-compose/graphql-gateway.yaml up
```

## Endpoints

- **GraphQL**: `http://localhost:4000/graphql`
- **Health Check**: `http://localhost:8088/health`
- **Metrics**: `http://localhost:9091/metrics`

## Configuration

### router.yaml

Main router configuration:

- CORS settings
- Header propagation
- OpenTelemetry integration
- Rate limiting
- Timeouts

### supergraph.yaml

Subgraph composition configuration:

- Subgraph routing URLs
- Schema sources

## Adding New Subgraphs

1. Enable federation in subgraph's `gqlgen.yml`
2. Add `@key` directives to entity types
3. Implement entity resolvers
4. Add subgraph to `supergraph.yaml`
5. Recompose supergraph locally: `task compose_graphql`
6. Commit both subgraph schema and `supergraph-schema.graphql`
7. CI/CD will validate and update if needed
8. Restart gateway

## Monitoring

The gateway exports:

- **Traces**: to OpenTelemetry Collector → Tempo
- **Metrics**: Prometheus format on port 9091
- **Logs**: JSON structured logs

View in Grafana: `http://localhost:3000`

## Troubleshooting

### Composition Fails

- Ensure all subgraph services are running
- Check subgraph GraphQL endpoints are accessible
- Verify schema compatibility with `rover subgraph check`

### Gateway Won't Start

- Check `supergraph-schema.graphql` exists
- Verify router.yaml syntax
- Check logs: `docker logs graphql-gateway`

### Entity Resolution Errors

- Verify `@key` directives match between services
- Check entity resolver implementations
- Review subgraph schema for conflicts
