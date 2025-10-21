# Federated GraphQL Implementation for Chat Service

## Overview

Implement **Apollo Federation v2** for the **chat-service** as the first subgraph, with a separate **GraphQL gateway service** to compose multiple subgraphs into a unified API.

---

## Architecture

**Apollo Federation Pattern:**

- **Subgraph (chat-service):** Exposes GraphQL schema with federation directives
- **Gateway Service:** Aggregates subgraphs into unified supergraph
- **API Gateway:** Proxies GraphQL requests to the gateway service

---

## Phase 1: Chat Service Subgraph Implementation

### 1. Dependencies & Tools

- Add `gqlgen` for Go GraphQL code generation
- Add `@apollo/subgraph` package for federation directives
- Create GraphQL schema with federation directives (`@key`, `@shareable`, `@external`)

### 2. GraphQL Schema Design (`chat-service/graph/schema.graphql`)

```graphql
# Federation directives
directive @key(fields: String!) on OBJECT
directive @extends on OBJECT
directive @external on FIELD_DEFINITION
directive @requires(fields: String!) on FIELD_DEFINITION
directive @provides(fields: String!) on FIELD_DEFINITION

# Extend User entity from auth-service
extend type User @key(fields: "id") {
  id: UUID! @external
  conversations: [Conversation!]!
  onlineStatus: OnlineStatus
}

type Conversation @key(fields: "id") {
  id: UUID!
  subject: String
  status: ConversationStatus!
  priority: Int!
  participants: [Participant!]!
  messages(limit: Int = 50, offset: Int = 0): MessageConnection!
  createdAt: Time!
  updatedAt: Time!
  endedAt: Time
}

type Message @key(fields: "id") {
  id: UUID!
  conversationId: UUID!
  senderId: UUID
  content: String!
  messageType: MessageType!
  isSystem: Boolean!
  createdAt: Time!
}

type Participant {
  id: UUID!
  conversationId: UUID!
  userId: UUID!
  userType: UserType!
  role: ParticipantRole!
  joinedAt: Time!
  leftAt: Time
  isActive: Boolean!
}
```

### 3. File Structure (New Files to Create)

```
chat-service/
├── graph/
│   ├── schema.graphql          # GraphQL schema with federation
│   ├── generated/              # gqlgen generated code
│   │   ├── generated.go
│   │   └── models.go
│   └── resolver.go             # Root resolver
├── internal/
│   ├── graph/
│   │   ├── resolver/           # GraphQL resolvers
│   │   │   ├── conversation_resolver.go
│   │   │   ├── message_resolver.go
│   │   │   ├── participant_resolver.go
│   │   │   └── user_resolver.go (federation)
│   │   └── model/              # GraphQL-specific models
│   │       └── mapper.go       # Entity to GraphQL model mapping
│   ├── config/
│   │   └── graphql.go          # GraphQL server config
│   └── routes/
│       └── graphql_routes.go   # GraphQL endpoint routes
├── gqlgen.yml                  # gqlgen configuration
```

### 4. Implementation Steps

1. **Install gqlgen and setup**

   ```sh
   go get github.com/99designs/gqlgen
   go run github.com/99designs/gqlgen init
   ```

2. **Configure `gqlgen.yml`**
   - Set federation mode: `version: 2`
   - Configure model bindings to existing entities
   - Set resolver paths

3. **Create GraphQL Resolvers**
   - `ConversationResolver`: Implements conversation queries/mutations
   - `MessageResolver`: Implements message queries
   - `ParticipantResolver`: Implements participant operations
   - `UserResolver`: Federation resolver for User entity references

4. **Add GraphQL Route**
   - Add `/graphql` endpoint in `chat_routes.go`
   - Handler wraps gqlgen's GraphQL handler
   - Apply auth middleware for protected operations

5. **Update Service Layer**
   - Reuse existing `ChatService` methods
   - Add any missing methods for GraphQL field resolvers
   - Ensure proper error handling for GraphQL context

---

## Phase 2: GraphQL Gateway Service

### 1. Create New Service (`graphql-gateway/`)

```
graphql-gateway/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   ├── gateway.go
│   │   └── discovery.go
│   ├── gateway/
│   │   ├── supergraph.go      # Compose subgraphs
│   │   ├── router.go          # Apollo Router integration
│   │   └── health.go
│   └── server/
│       └── http_server.go
├── supergraph.yaml            # Subgraph configuration
└── go.mod
```

### 2. Gateway Implementation Options

- **Option A: Apollo Router (Recommended)**
  - Use official Apollo Router (Rust binary)
  - Configuration via YAML
  - Auto-compose subgraphs via introspection
  - Best performance and official support

- **Option B: Pure Go Gateway**
  - Use [`github.com/nautilus/gateway`](https://github.com/nautilus/gateway)
  - All Go implementation
  - More control but less feature-complete

### 3. Supergraph Configuration (`supergraph.yaml`)

```yaml
federation_version: 2
subgraphs:
  chat:
    routing_url: http://chat-service:8003/graphql
    schema:
      subgraph_url: http://chat-service:8003/graphql
  # Future subgraphs:
  # auth:
  #   routing_url: http://auth-service:8001/graphql
  # product:
  #   routing_url: http://product-service:8002/graphql
```

### 4. Gateway Features

- Service discovery integration (Consul)
- Query planning and execution
- Error handling and aggregation
- Authentication/authorization propagation
- Caching and performance optimization

---

## Phase 3: API Gateway Integration

### 1. Add GraphQL Proxy Route (`api-gateway/internal/routes/gateway_routes.go`)

```go
// GraphQL endpoint
public.POST("/graphql", gw.ProxyToService("graphql-gateway", "/graphql"))
public.GET("/graphql/playground", gw.ProxyToService("graphql-gateway", "/playground"))
```

### 2. CORS Configuration

- Add GraphQL-specific headers
- Support introspection queries
- Enable playground for development

---

## Phase 4: Infrastructure & Deployment

### 1. Docker Compose (`deployments/docker-compose/graphql-gateway.yml`)

- GraphQL gateway service
- Health checks
- Consul registration

### 2. Taskfile Commands (`taskfile.yml`)

```yaml
graphql-generate:
  desc: "Generate GraphQL code"
  cmds:
    - cd chat-service && go run github.com/99designs/gqlgen generate

graphql-compose:
  desc: "Compose supergraph schema"
  cmds:
    - cd graphql-gateway && rover supergraph compose --config supergraph.yaml
```

---

## Files to Create/Modify

**New Files (Chat Service):**

1. `chat-service/graph/schema.graphql` – GraphQL schema with federation
2. `chat-service/gqlgen.yml` – gqlgen configuration
3. `chat-service/internal/graph/resolver/*.go` – GraphQL resolvers
4. `chat-service/internal/config/graphql.go` – GraphQL config
5. `chat-service/internal/routes/graphql_routes.go` – GraphQL routes

**New Service (Gateway):** 6. `graphql-gateway/` – Entire new service directory

**Modified Files:** 7. `chat-service/internal/routes/chat_routes.go` – Add GraphQL endpoint 8. `chat-service/go.mod` – Add gqlgen dependency 9. `api-gateway/internal/routes/gateway_routes.go` – Add GraphQL proxy 10. `deployments/docker-compose/services.yml` – Add gateway service 11. `taskfile.yml` – Add GraphQL tasks

---

## Benefits

1. Unified API: Single GraphQL endpoint for all services
2. Type Safety: Strong typing across services
3. Flexible Queries: Clients request exactly what they need
4. Real-time: GraphQL subscriptions for WebSocket events
5. Gradual Migration: Add services incrementally
6. Federation: Each service owns its domain
7. Introspection: Auto-generated documentation

---

## Future Subgraphs

After chat-service:

1. **auth-service:** User entity owner, authentication
2. **product-service:** Product catalog
3. **order-service:** Orders and transactions
4. **notification-service:** Notifications

---

## Notes

- REST API remains available alongside GraphQL
- WebSocket connections stay separate (not via GraphQL subscriptions initially)
- Authentication via JWT tokens in `Authorization` header
- Use **Dataloader** for N+1 query optimization
- Consider GraphQL caching strategies (Apollo Client, Redis)
