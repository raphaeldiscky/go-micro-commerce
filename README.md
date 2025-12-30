<h1 align="center">Go Micro Commerce</h1>

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25.5-blue.svg)](https://golang.org/) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This application is primarily intended for exploring technical concepts. My goal is to experiment with different technologies, software architecture designs, and all the essential components involved in building distributed systems in Golang.

## Features :sparkles:

- `Event-driven architecture` using `Kafka` for event streaming, `Redis PubSub` for real-time message broadcasting, and `Asynq` for distributed task scheduling (order and cart services)
- `Clean Architecture` (entity, repository, service, handler) with `Domain-Driven Design (DDD)` principles across all services
- Each microservice has its own dedicated `PostgreSQL` database instance
- 3-node `Kafka Cluster` running on `KRaft mode` (ZooKeeper-free)
- 6-node `Redis Cluster` (3 masters + 3 replicas)
- Central instrumentation using `OpenTelemetry` combined with LGTM stack (`Loki, Grafana, Tempo, Prometheus`) and `Alloy` as telemetry collector
- Two local development options:
  - `Docker Compose` setup for rapid development (infrastructure + 7 core services)
  - `Kubernetes Cluster` with `Tilt + (Kind or MicroK8s)` for hot reload in a production-like environment with all 11 services
- `Kubernetes Gateway API` with `Traefik` as gateway controller for production-grade traffic routing
- CI pipeline using `GitHub Actions` to automate build, test, and push images to a registry
- `GitOps` with `ArgoCD` for declarative, pull-based continuous deployment
- Infrastructure as Code with `Terraform` for GKE cluster provisioning on GCP
- `Kubernetes` for robust, scalable container orchestration in production environments
- Secure authentication implemented via `JWT` with `RS256` asymmetric algorithm and refresh token rotation
- Unified REST `API Gateway` and `GraphQL Federation` for type-safe client-server communication
- Internal communication via synchronous `gRPC calls` for microservices to interact with each other.
- Database Management with schema migrations handled by `golang-migrate`
- Validation using `go-playground/validator` for input sanitization
- Order creation is implemented using two saga orchestration options:
  - A custom `PostgreSQL-based saga` implementation
  - A managed workflow service engine using `Temporal`
- Implemented `message inbox pattern` for idempotent event consumption and `transactional outbox pattern` for publishing domain events
- `Server-Sent Events (SSE)` for real-time push notification delivery in the notification-service.
- `WebSocket` support in the chat-service for bi-directional communication.
- Uses [bytedance/sonic](https://github.com/bytedance/sonic) instead of the standard Go library for serialization/deserialization, offering up to 5x faster unmarshalling and significant marshaling performance improvements, as demonstrated in this [benchmark](https://github.com/centralci/go-benchmarks/tree/b647c45272c7dc371fd4337cb3b6546356d967d1/json)

## Technology Stack 🛠️

- **[labstack/echo](https://github.com/labstack/echo)** - high performance, minimalist go web framework
- **[jackc/pgx/v5](https://github.com/jackc/pgx)** - PostgreSQL driver and toolkit for Go
- **[ibm/sarama](https://github.com/IBM/sarama)** - go library for Apache Kafka
- **[redis/go-redis](https://github.com/redis/go-redis)** - redis go client for cache and Pub/Sub
- **[bsm/redislock](https://github.com/bsm/redislock)** - distributed locking implementation using Redis
- **[elastic/go-elasticsearch](https://github.com/elastic/go-elasticsearch)** - official Go client for Elasticsearch
- **[hibiken/asynq](https://github.com/hibiken/asynq)** - simple, reliable, and efficient distributed task queue in Go using Redis
- **[google.golang.org/protobuf](https://github.com/protocolbuffers/protobuf-go)** - Protocol Buffers for Go
- **[connectrpc/connect-go](https://github.com/connectrpc/connect-go)** - protobuf RPC framework
- **[bufbuild/buf](https://github.com/bufbuild/buf)** - linter, formatter, generator for protobuf
- **[99designs/gqlgen](https://github.com/99designs/gqlgen)** - Go code generation-based GraphQL server library
- **[apollographql/router](https://github.com/apollographql/router)** - routing runtime for Apollo Federation
- **[oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)** - generate Go client and server boilerplate from OpenAPI 3 specifications
- **[bytedance/sonic](https://github.com/bytedance/sonic)** - a blazingly fast JSON serializing & deserializing library
- **[stretchr/testify](https://github.com/stretchr/testify)** - testing toolkit
- **[testcontainers/testcontainers-go](https://github.com/testcontainers/testcontainers-go)** - testcontainers for go
- **[spf13/viper](https://github.com/spf13/viper)** - go configuration with fangs
- **[spf13/cobra](https://github.com/spf13/cobra)** - a commander for modern go CLI interactions
- **[hashicorp/consul](https://github.com/hashicorp/consul)** - service registration and discovery
- **[docker](https://www.docker.com/)** - container platform
- **[go-playground/validator/v10](https://github.com/go-playground/validator)** - go struct and field validation
- **[golang/crypto](https://github.com/golang/crypto)** - cryptographic functions
- **[golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)** - go implementation of JWT
- **[gorilla/websocket](https://github.com/gorilla/websocket)** - websocket implementation for go
- **[temporal](https://github.com/temporalio/temporal)** - workflow engine service
- **[sony/gobreaker](https://github.com/sony/gobreaker)** - circuit breaker implemented in go
- **[prometheus/client_golang](https://github.com/prometheus/client_golang)** - prometheus instrumentation lib for go apps
- **[opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go)** - opentelemetry go API and SDK
- **[grafana/alloy](https://github.com/grafana/alloy)** - vendor-agnostic OpenTelemetry Collector distribution for telemetry data collection and forwarding
- **[shopspring/decimal](https://github.com/shopspring/decimal)** - precision fixed-point decimal numbers in go
- **[google/uuid](https://github.com/google/uuid)** - go package for UUIDs

## Architecture Overview 🏗️

The system follows a microservices architecture where each service represents an independent business domain. Services communicate through both synchronous gRPC calls and asynchronous event-driven patterns. Data consistency across distributed transactions is maintained using Saga orchestration patterns, with support for both custom PostgreSQL-based implementations and Temporal-managed workflows.

### 1. Authentication Service

Manages user identity, authentication, and session lifecycle with secure token-based authentication.

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant Gateway
  participant AuthService as Auth Service
  participant PostgreSQL
  participant Kafka
  participant NotificationService as Notification Service
  participant EmailProvider as Email Provider

  Client->>Gateway: POST /v1/register
  Gateway->>AuthService: Forward (REST/gRPC)
  AuthService->>PostgreSQL: INSERT new user (status=UNVERIFIED)
  AuthService->>Kafka: PUBLISH UserRegistered
  Kafka-->>NotificationService: CONSUME UserRegistered
  NotificationService->>NotificationService: Render verification email template
  NotificationService->>EmailProvider: Send verification email
  EmailProvider-->>NotificationService: 202 Accepted
  AuthService-->>Gateway: Return 201 Created
  Gateway-->>Client: 201 Created (User pending verification)
```

**Responsibilities**:

- User lifecycle management (registration, verification, profile updates)
- Service-to-service authentication and authorization
- Secure session and token management

**Entities**: `users`, `sessions`, `addresses`

**Key Features**:

- JWT-based authentication with RS256 asymmetric algorithm
- Short-lived access tokens (15-30 minutes) and long-lived refresh tokens (7-30 days)
- Email verification with time-limited tokens (24-hour expiry)
- Resend verification capability with rate limiting

### 2. Product Service

Manages the product catalog, inventory tracking, and pricing.

```mermaid
sequenceDiagram
  autonumber
  participant Admin as Admin Client
  participant ProductService as Product Service
  participant PostgreSQL
  participant Redis
  participant Kafka
  participant OrderService as Order Service

  Admin->>ProductService: POST /v1/products
  ProductService->>PostgreSQL: INSERT product (version=1)
  ProductService->>Redis: Cache product details
  ProductService->>Kafka: PUBLISH ProductCreated

  Admin->>ProductService: PUT /v1/products/:id
  ProductService->>PostgreSQL: UPDATE product with optimistic lock (version++)
  ProductService->>Redis: Invalidate cache
  ProductService->>Kafka: PUBLISH ProductUpdated

  OrderService->>ProductService: gRPC ReserveProducts()
  ProductService->>PostgreSQL: Lock rows + decrement available stock
  ProductService-->>OrderService: Reservation confirmed

  OrderService->>ProductService: gRPC DeductStock() (after payment success)
  ProductService->>PostgreSQL: Deduct stock permanently
```

**Responsibilities**:

- Full product lifecycle management (CRUD operations)
- Inventory reservation and deduction with concurrency control
- Price management and versioning

**Entities**: `products`, `outbox_events`

**Key Features**:

- Optimistic locking using version column to prevent lost updates
- Stock reservation during order placement (via gRPC)
- Idempotent stock deduction after payment confirmation
- Cache-aside pattern with Redis for high-read performance

### 3. Order Service

Orchestrates complex order workflows using a hybrid approach combining synchronous and asynchronous communication patterns.

**Order Lifecycle Overview**:

```mermaid
graph TD
  A[Place Order] --> B[Schedule Reminders]
  A --> C[Start 24h Countdown]
  A --> D[Create Payment Intent]
  A --> E[Fetch Checkout Session + Shipping Cost]
  D --> F[Return Gateway Metadata]
  B --> G[4h Reminder]
  B --> H[12h Reminder]
  B --> I[22h Reminder]
  C --> J[Expiration Check]

  G --> K[Send Email via Notification Service]
  H --> K
  I --> K
  J --> L{Payment Status?}

  L -->|Completed| M[Mark Order as Paid]
  M --> N[Trigger Async Post-Payment Saga]
  N --> O[Skip Payment Reminder + Fulfillment + Stock Update]
  L -->|Timeout| P[Mark Order as Expired]
  P --> Q[Restock Inventory]
  L -->|Canceled| S[Mark Order as Canceled]
  S --> Q
```

**Order Placement Flow**:

```mermaid
sequenceDiagram
  autonumber
  participant User
  participant OrderSvc as Order Service
  participant CartSvc as Cart Service
  participant FulfillSvc as Fulfillment Service
  participant PaySvc as Payment Service
  participant Kafka

  User->>OrderSvc: POST /orders (checkoutSessionId)
  OrderSvc->>CartSvc: gRPC getCheckoutSession()
  CartSvc-->>OrderSvc: CheckoutData
  OrderSvc->>FulfillSvc: gRPC getShippingCost()
  FulfillSvc-->>OrderSvc: ShippingCost
  OrderSvc->>PaySvc: gRPC createPaymentIntent()
  PaySvc-->>OrderSvc: PaymentIntent (client_secret)
  OrderSvc-->>User: { orderId, client_secret }

  PaySvc-->>Kafka: PaymentSucceeded
  Kafka-->>OrderSvc: PaymentSucceeded
  OrderSvc-->>Kafka: OrderConfirmed
  Kafka-->>FulfillSvc: OrderConfirmed
```

**Payment Reminder Flow**:

```mermaid
sequenceDiagram
  autonumber
  participant Asynq as Task Queue
  participant OrderService
  participant PostgreSQL
  participant NotificationService as Notification Service

  Asynq->>OrderService: Trigger payment reminder (after 4h)
  OrderService->>PostgreSQL: Fetch order by ID
  alt Order status = PENDING_PAYMENT
      OrderService->>NotificationService: Send "Payment Reminder" message/email
      NotificationService-->>OrderService: Acknowledged
      OrderService-->>Asynq: Task completed
  else Order status = PAID or EXPIRED
      OrderService-->>Asynq: Skip reminder (no action)
  end
```

**Order Expiration Flow**:

```mermaid
sequenceDiagram
  autonumber
  participant Asynq as Task Queue
  participant OrderService as Order Service
  participant PostgreSQL
  participant Kafka

  Asynq->>OrderService: Trigger 24h expiration task
  OrderService->>PostgreSQL: Fetch order by ID
  alt Order status = PENDING_PAYMENT
      OrderService->>PostgreSQL: Update order -> EXPIRED
      OrderService->>Kafka: PUBLISH OrderExpired
      Kafka-->>OrderService: Ack
      OrderService-->>Asynq: Task completed
  else Order status = PAID
      OrderService-->>Asynq: Skip expiration
  end
```

**Responsibilities**:

- Order lifecycle orchestration and state management
- Coordination between cart, payment, and fulfillment services
- Saga pattern implementation for distributed transactions

**Entities**: `orders`, `order_items`, `inbox_events`, `outbox_events`, `saga_states`

**Key Features**:

- Distributed locking with Redis for operation idempotency
- Dual saga implementation (custom PostgreSQL-based and Temporal-managed)
- Automated payment reminders and order expiration
- Support for order modifications and cancellations

### 4. Fulfillment Service

Manages order fulfillment operations, shipping coordination, and delivery tracking.

```mermaid
sequenceDiagram
  autonumber
  participant FulfillmentService as Fulfillment Service
  participant ShipSvc as Shipping Provider
  participant PostgreSQL
  participant Kafka

  Kafka-->>FulfillmentService: CONSUME OrderPaid
  FulfillmentService->>PostgreSQL: INSERT fulfillment (status=PENDING)
  FulfillmentService->>ShipSvc: Create shipping order
  ShipSvc-->>FulfillmentService: TrackingID + ETA
  FulfillmentService->>PostgreSQL: UPDATE fulfillment (status=IN_PROGRESS)
  FulfillmentService->>Kafka: PUBLISH FulfillmentCreated
```

**Responsibilities**:

- Delivery and shipping management
- Shipping cost calculation and processing

**Entities**: `fulfillments`

### 5. Payment Service

Handles payment processing with multiple gateway integrations.

```mermaid
sequenceDiagram
  autonumber
  participant OrderService
  participant PaymentService
  participant Stripe as Payment Gateway
  participant PostgreSQL
  participant Kafka

  OrderService->>PaymentService: gRPC createPaymentIntent()
  PaymentService->>Stripe: Create payment intent (amount, metadata)
  Stripe-->>PaymentService: client_secret
  PaymentService->>PostgreSQL: Store payment tx
  PaymentService-->>OrderService: Return client_secret

  Stripe-->>PaymentService: Webhook (payment_succeeded)
  PaymentService->>PostgreSQL: Update payment status = SUCCEEDED
  PaymentService->>Kafka: PUBLISH PaymentSucceeded

  Stripe-->>PaymentService: Webhook (payment_failed)
  PaymentService->>PostgreSQL: Update payment status = FAILED
  PaymentService->>Kafka: PUBLISH PaymentFailed
```

**Responsibilities**:

- Payment processing and transaction management
- Multiple payment gateway integrations
- Webhook handling for payment status updates
- Refund processing and dispute management

**Entities**: `payments`, `outbox_events`, `inbox_events`

**Key Features**:

- Payment gateway factory pattern supporting Stripe and additional payment providers
- Idempotent payment processing with idempotency keys
- Secure webhook verification and handling
- Comprehensive payment analytics and reporting

### 6. Notification Service

Processes and delivers notifications across multiple communication channels.

```mermaid
sequenceDiagram
  autonumber
  participant Kafka
  participant NotificationService as Notification Service
  participant PostgreSQL
  participant Redis
  participant EmailProvider as Email Provider
  participant SMSProvider as SMS Provider
  participant Client

  Kafka-->>NotificationService: CONSUME events (UserRegistered, PaymentReminder, OrderConfirmed, etc.)
  NotificationService->>PostgreSQL: Log notification event
  NotificationService->>Redis: Publish via Redis Pub/Sub
  Redis-->>Client: SSE Push (real-time)
  NotificationService->>EmailProvider: Send email (async)
  NotificationService->>SMSProvider: Send SMS (optional)
```

**Responsibilities**:

- Asynchronous notification processing and delivery
- Multi-channel notification support (email, SMS, push)
- Notification template management

**Entities**: `notifications`, `inbox_events`

**Key Features**:

- Real-time push notifications with Server-Sent Events (SSE)
- Async email processing
- SMS notification support with failover providers

### 7. Chat Service

Provides real-time customer support and bi-directional communication capabilities.

```mermaid
sequenceDiagram
  autonumber
  participant UserA
  participant ChatService as Chat Service
  participant PostgreSQL
  participant Redis as Redis Pub/Sub
  participant UserB

  UserA->>ChatService: Send message via WebSocket
  ChatService->>PostgreSQL: Persist message
  ChatService->>Redis: Publish message to channel (userB)
  Redis-->>ChatService: Message received on subscribed channel
  ChatService-->>UserB: WebSocket push (real-time)
```

**Responsibilities**:

- Live chat implementation for customer support
- Real-time message delivery and persistence
- Chat room management and moderation
- File sharing and rich media support

**Entities**: `conversations`, `messages`, `participants`, `connections`

**Key Features**:

- WebSocket-based real-time communication
- Conversation history and search
- Typing indicators and online status
- Support for group chats and channels

### 8. Cart Service

Manages shopping cart lifecycle and checkout session preparation.

```mermaid
sequenceDiagram
  autonumber
  participant User
  participant CartService as Cart Service
  participant PostgreSQL
  participant Redis
  participant OrderService as Order Service

  User->>CartService: Add item to cart
  CartService->>PostgreSQL: INSERT cart_item
  CartService->>Redis: Cache updated cart state

  User->>CartService: POST /checkout
  CartService->>PostgreSQL: Create checkout session
  CartService-->>User: CheckoutSessionID

  OrderService->>CartService: gRPC getCheckoutSession(sessionId)
  CartService-->>OrderService: Checkout details (items, subtotal, vouchers)
```

**Responsibilities**:

- Shopping cart lifecycle management
- Checkout session generation and validation
- Cart abandonment tracking and recovery
- Promotional code, payment gateway, and courier selection

**Entities**: `carts`, `cart_items`, `checkout_sessions`, `outbox_events`

**Key Features**:

- Cart synchronization and persistence
- Promotional code validation and application
- Cart expiration and cleanup

### 9. Search Service

Provides full-text search functionality and advanced filtering capabilities.

```mermaid
sequenceDiagram
  autonumber
  participant Kafka
  participant SearchService as Search Service
  participant Elasticsearch
  participant Client

  Kafka-->>SearchService: CONSUME ProductCreated / ProductUpdated / OrderCreated
  SearchService->>Elasticsearch: Index / Update document

  Client->>SearchService: GET /search?q=keyword
  SearchService->>Elasticsearch: Query index
  Elasticsearch-->>SearchService: Matched results
  SearchService-->>Client: Return search results
```

**Responsibilities**:

- Document indexing and search functionality

**Entities**: `inbox_events`

**Key Features**:

- Real-time indexing via Kafka events
- Advanced full-text search with fuzzy matching
- Faceted search with filters and aggregations
- Search relevance scoring and boosting
- Search analytics and popular queries

### 10. API Gateway

Serves as the unified entry point for all client requests with routing capabilities.

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant APIGateway as API Gateway
  participant AuthService
  participant ProductService
  participant OrderService
  participant ChatService
  participant NotificationService

  Client->>APIGateway: HTTP Request (REST/gRPC/WebSocket)
  APIGateway->>AuthService: Validate JWT / Session
  alt Auth success
      APIGateway->>ProductService: Route request (if /products)
      APIGateway->>OrderService: Route request (if /orders)
      APIGateway->>ChatService: WebSocket connection
      APIGateway->>NotificationService: Subscribe to SSE
      APIGateway-->>Client: Forward response
  else Invalid token
      APIGateway-->>Client: 401 Unauthorized
  end
```

**Responsibilities**:

- Unified API entry point and request routing
- Authentication and authorization middleware
- Rate limiting and request throttling
- Protocol translation (REST/gRPC/WebSocket/SSE)
- Service discovery and load balancing

**Key Features**:

- JWT validation middleware
- Configurable rate limiting per endpoint and user
- Request/response transformation and validation
- Circuit breaker pattern for fault tolerance
- Comprehensive request logging and metrics
- CORS and security headers management

### 11. GraphQL Gateway

Provides a federated GraphQL interface for unified data access and querying.

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant GraphQLGateway as GraphQL Gateway
  participant ProductService
  participant OrderService
  participant AuthService

  Client->>GraphQLGateway: GraphQL query/mutation
  GraphQLGateway->>ProductService: Resolve product field (Federation)
  GraphQLGateway->>OrderService: Resolve order field
  GraphQLGateway->>AuthService: Resolve user field
  GraphQLGateway-->>Client: Combined federated response
```

**Responsibilities**:

- Unified GraphQL schema federation
- Client-specific schema customization

**Key Features**:

- Apollo Federation for schema composition
- Custom JWT Authentication
- Real-time subscriptions support
- Schema validation and versioning

### 12. Observability Stack

Comprehensive monitoring, distributed tracing, and centralized logging for end-to-end system observability.

```mermaid
graph TD
  A[Microservices] --> B[OpenTelemetry SDKs]
  B --> C[Metrics]
  B --> D[Traces]
  B --> E[Logs]
  C --> F[Prometheus]
  D --> G[Tempo]
  E --> H[Alloy]
  H --> I[Loki]
  F --> J[Grafana]
  G --> J
  I --> J
```

**Responsibilities**:

- System-wide monitoring and alerting
- Distributed tracing for request flow analysis
- Centralized logging and log aggregation
- Performance metrics collection and visualization

**Key Features**:

- **Prometheus** - Real-time metrics collection with service-level indicators
- **Tempo** - End-to-end distributed tracing across service boundaries
- **Alloy** - Vendor-agnostic telemetry collector deployed as DaemonSet for Kubernetes log aggregation
- **Loki** - Centralized logging with structured labels and efficient storage
- **Grafana** - Unified dashboards for metrics, traces, and logs correlation

### 13. Frontend Application

Modern React-based user interface with real-time communication capabilities.

```mermaid
sequenceDiagram
  autonumber
  participant User
  participant FrontendApp as React Vite + TanStack
  participant APIGateway as API Gateway
  participant GraphQLGateway as GraphQL Gateway

  User->>FrontendApp: Browse / Add to cart / Checkout
  FrontendApp->>APIGateway: REST/gRPC calls (auth, cart, orders)
  FrontendApp->>GraphQLGateway: GraphQL queries for aggregated data
  GraphQLGateway-->>FrontendApp: Unified response
  FrontendApp-->>User: Render UI with real-time updates
```

**Responsibilities**:

- User interface rendering and interaction handling
- State management and data synchronization
- Real-time updates via WebSocket and SSE

**Key Features**:

- Use Tanstack Query, Form, and Router
- TypeScript for type safety and developer experience
- Real-time updates for chat, notifications, and order status
- Performance optimization with code splitting and lazy loading

### 14. Production Deployment

The production environment uses a two-phase deployment strategy combining Infrastructure as Code (IaC) with GitOps.

#### Infrastructure Provisioning (Terraform)

Terraform provisions the complete GKE cluster infrastructure:

```text
terraform/
├── modules/                    # Reusable infrastructure modules
│   ├── gcp-network/            # VPC, subnets, Cloud NAT
│   ├── gke-cluster/            # 5-tier node pool architecture
│   ├── cloudnative-pg-operator/
│   ├── strimzi-kafka-operator/
│   ├── redis-operator/
│   ├── monitoring/             # Prometheus, Grafana, Loki, Tempo, Alloy
│   ├── argocd/                 # GitOps controller
│   └── traefik/                # Ingress controller
└── environments/prod/          # Production configuration
```

**Key Features**:

- 5-tier node pool architecture (stateful, stateless, monitoring, infra, gateway)
- Spot VMs for stateless workloads (~60% cost savings)
- Automated TLS with cert-manager and Let's Encrypt
- External Secrets Operator for GCP Secret Manager integration

**See**: [`terraform/README.md`](./terraform/README.md) for complete infrastructure setup.

#### Application Deployment (GitOps)

ArgoCD manages application deployments using Kustomize overlays:

```text
deployments/k8s/
├── apps/applicationsets/      # ArgoCD ApplicationSet definitions
├── infrastructure/            # Platform services (databases, monitoring)
│   ├── base/                  # Environment-agnostic configs
│   └── overlays/prod/         # Production-specific patches
└── workloads/                 # Microservices
    ├── base/                  # Shared configurations
    └── overlays/prod/         # Production replicas, resources, HPA
```

**Deployment Flow**:

1. Developer pushes to `main` branch
2. ArgoCD detects changes and syncs to cluster
3. Kustomize builds manifests with production overlays
4. Applications deployed with HA configuration

**See**: [`deployments/k8s/README.md`](./deployments/k8s/README.md) for GitOps workflow details.

#### Frontend Deployment (Cloudflare Pages)

**Note**: The frontend application is deployed separately from the Kubernetes infrastructure via Cloudflare Pages for optimal cost and performance:

**Deployment Architecture**:

```text
GitHub (main branch) → Cloudflare Pages → Global CDN (300+ locations)
```

**Benefits**:

- Zero infrastructure cost (Cloudflare Pages free tier)
- Global edge network with sub-50ms latency worldwide
- Automatic deployments on git push
- Preview deployments for every pull request
- Built-in SPA routing and asset optimization
- Automatic HTTPS with Cloudflare certificates

**Configuration**: See `terraform/modules/cloudflare-pages/` for IaC setup

**Domain**: Frontend served from Cloudflare-managed domain, backend APIs from GKE cluster

This separation allows the backend microservices to scale independently while keeping the static frontend optimally distributed at the edge.
