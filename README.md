<h1 align="center">Go Micro Commerce</h1>

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24.7-blue.svg)](https://golang.org/) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This application is primarily intended for exploring technical concepts. My goal is to experiment with different technologies, software architecture designs, and all the essential components involved in building distributed systems in Golang.

## Features :sparkles:

- Event-driven architecture using `Kafka` for event streaming, `Redis Pub/Sub` for message broadcasting, and `Asynq` for distributed task queues
- Each service implemented with Domain-Driven Design and Hexagonal architecture
- Custom saga workflow with saga states stored in `Postgres` and managed service using `Temporal`
- Use RS256 as the asymmetric JWT algorithm for microservices authentication
- `GraphQL Federation` for API specification and type-safety between client and server
- Synchronous `gRPC` for internal service-to-service communication
- Database migrations using `golang-migrate`
- Input validation with `go-playground/validator`
- CI/CD pipeline using `GitHub Actions`

## Technologies - Libraries 🛠️

- **[labstack/echo](https://github.com/labstack/echo)** - high performance, minimalist go web framework
- **[connectrpc/connect-go](https://github.com/connectrpc/connect-go)** - protobuf RPC
- **[99designs/glqgen](https://github.com/99designs/gqlgen)** - go generate based graphql server library
- **[jackc/pgx/v5](https://github.com/jackc/pgx)** - postgres driver and toolkit for Go
- **[ibm/sarama](https://github.com/IBM/sarama)** - go library for Apache Kafka.
- **[redis/go-redis](https://github.com/redis/go-redis)** - redis go client
- **[bsm/redislock](https://github.com/bsm/redislock)** - distributed locking implementation using Redis
- **[hibiken/asynq](https://github.com/hibiken/asynq)** - simple, reliable, and efficient distributed task queue in Go
- **[stretchr/testify](https://github.com/stretchr/testify)** - testing toolkit
- **[testcontainers/testcontainers-go](https://github.com/testcontainers/testcontainers-go)** - testcontainers for go
- **[spf13/viper](https://github.com/spf13/viper)** - go configuration with fangs
- **[spf13/cobra](https://github.com/spf13/cobra)** - a commander for modern go CLI interactions
- **[hashicorp/consul](https://github.com/hashicorp/consul)** - service discovery
- **[docker](https://www.docker.com/)** - container platform
- **[go-playground/validator/v10](https://github.com/go-playground/validator)** - go struct and field validation
- **[golang/crypto](https://github.com/golang/crypto)** - cryptographic functions
- **[golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)** - go implementation of JWT
- **[gorilla/websocket](https://github.com/gorilla/websocket)** - websocket implementation for go
- **[temporal](https://github.com/temporalio/temporal)** - workflow service

## Architecture Overview 🏗️

### 1. Auth Service

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

- User lifecycle management and service-to-service authentication
- Secure session and token handling
- Entities: `users`, `sessions`, `addresses`

**Key features**:

- JWT-based authentication with RS256 (asymmetric) algorithm, short-lived access + long-lived refresh tokens
- User verification via email with time-limited token (24h)
- Resend capability

### 2. Product Service

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

- Full product lifecycle management
- Inventory reservation and deduction with concurrency control
- Entities: `products`

**Key features**:

- Optimistic locking using a `version` column to prevent lost updates
- Stock reservation during order placement (via `gRPC`)
- Stock deduction only after payment confirmation (idempotent)
- Event publishing via `outbox pattern`
- Cache-aside pattern with `Redis` for high-read performance

### 3. Order Service

**Order Overview**:

```mermaid
graph TD
  A[Place Order] --> B[Schedule Reminders]
  A --> C[Start 24h Countdown]
  A --> D[Create Payment Intent]
  A --> E[Fetch Checkout Session
  +
  Shipping Cost]
  D --> F[Return Gateway Metadata]
  B --> G[4h Reminder]
  B --> H[12h Reminder]
  B --> I[22h Reminder]
  C --> J[Expiration Check]

  G --> K[Send Email
   via Notification Service]
  H --> K
  I --> K
  J --> L{Payment Status?}

  L -->|Completed| M[Mark Order as Paid]
  M --> N[Trigger Async
  Post-Payment Saga]
  N --> O[Skip Payment Reminder +
  Fulfillment +
  Stock Update
  ]
  L -->|Timeout| P[Mark Order as Expired]
  P --> Q[Restock Inventory]
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
      OrderService->>PostgreSQL: Update order → EXPIRED
      OrderService->>Kafka: PUBLISH OrderExpired
      Kafka-->>OrderService: Ack
      OrderService-->>Asynq: Task completed
  else Order status = PAID
      OrderService-->>Asynq: Skip expiration
  end
```

**Responsibilities**:

- Orchestrate the order creation
- Entities: `orders`, `order_items`, `inbox_events`, `outbox_events`, `saga_states`

**Key features**:

- Distributed locking for idempotency with Redis
-

### 4. Fulfillment Service

```mermaid
sequenceDiagram
  autonumber
  participant FulfillmentService as Fulfillment Service
  participant FullSvc as Shipping Provider
  participant PostgreSQL
  participant Kafka

  Kafka-->>FulfillmentService: CONSUME OrderPaid
  FulfillmentService->>PostgreSQL: INSERT fulfillment (status=PENDING)
  FulfillmentService->>FullSvc: Create shipping order
  FullSvc-->>FulfillmentService: TrackingID + ETA
  FulfillmentService->>PostgreSQL: UPDATE fulfillment (status=IN_PROGRESS)
  FulfillmentService->>Kafka: PUBLISH FulfillmentCreated
```

**Responsibilities**:

- Delivery service
- Shipping cost processing

**Key features**:

### 5. Payment Service

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

- Payment processing

**Key features**:

- Stripe, Xendit, etc (Factory)

### 6. Notification Service

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

- Async notification processing

**Key features**:

- Push notification with SSE
- Async email processing
- SMS notification

### 7. Chat Service

```mermaid
sequenceDiagram
  autonumber
  participant UserA
  participant ChatService as Chat Service
  participant PostgreSQL
  participant Redis
  participant UserB

  UserA->>ChatService: Send message via WebSocket
  ChatService->>PostgreSQL: Persist message
  ChatService->>Redis: Publish message to channel (userB)
  Redis-->>ChatService: Deliver message to UserB connection
  ChatService-->>UserB: WebSocket push (real-time)

```

**Responsibilities**:

- Live chat implementation

**Key features**:

- with WebSocket

### 8. Cart Service

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

- Cart persistence
- Checkout Session

**Key features**:

### 9. Search Service

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

- Full text search

**Key features**:

- Elasticsearch

### 10. API Gateway

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

**Responsibilities**: Unified entry point, auth, rate limiting, service discovery, protocol routing (REST/gRPC/WebSocket/SSE)

**Key features**:

- JWT validation middleware
- Rate limiting

### 11. GraphQL Gateway

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

- Unified graphql

**Key features**:

- graphql federation

### 12. Observability

**Responsibilities**:

**Key features**:

- `Prometheus` - Real-time metrics for each service
- `Tempo` - End-to-end distributed tracing across services
- `Loki` - Centralized logging with structured labels
- `Grafana` - Performance dashboards for each service

```mermaid
graph TD
  A[Microservices] --> B[OpenTelemetry SDKs]
  B --> C[Metrics]
  B --> D[Traces]
  B --> E[Logs]
  C --> F[Prometheus]
  D --> G[Tempo]
  E --> H[Loki]
  F --> I[Grafana]
  G --> I
  H --> I
```

### 13. Frontend

```mermaid
sequenceDiagram
  autonumber
  participant User
  participant FrontendApp as React/Next.js App
  participant APIGateway as API Gateway
  participant GraphQLGateway as GraphQL Gateway

  User->>FrontendApp: Browse / Add to cart / Checkout
  FrontendApp->>APIGateway: REST/gRPC calls (auth, cart, orders)
  FrontendApp->>GraphQLGateway: GraphQL queries for aggregated data
  GraphQLGateway-->>FrontendApp: Unified response
  FrontendApp-->>User: Render UI with real-time updates
```
