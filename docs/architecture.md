# Go-DDD with Event-Driven Architecture

This is an enhanced version of the Go DDD marketplace application that includes:

- **gRPC** APIs alongside REST APIs
- **Redis** caching for improved performance
- **Kafka** for event-driven architecture
- **Domain Events** for decoupled communication

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐
│   gRPC Client   │    │   REST Client   │
└─────────┬───────┘    └─────────┬───────┘
          │                      │
          └──────────┬───────────┘
                     │
┌────────────────────▼────────────────────┐
│              Interface Layer            │
│  ┌─────────────┐    ┌─────────────┐    │
│  │ gRPC Server │    │ REST Server │    │
│  └─────────────┘    └─────────────┘    │
└────────────────────┬────────────────────┘
                     │
┌────────────────────▼────────────────────┐
│           Application Layer             │
│  ┌─────────────────────────────────────┐ │
│  │        Domain Services              │ │
│  │     (with Event Publishing)         │ │
│  └─────────────────────────────────────┘ │
└────────────────────┬────────────────────┘
                     │
┌────────────────────▼────────────────────┐
│            Domain Layer                 │
│  ┌─────────────┐    ┌─────────────┐    │
│  │  Entities   │    │   Events    │    │
│  └─────────────┘    └─────────────┘    │
└────────────────────┬────────────────────┘
                     │
┌────────────────────▼────────────────────┐
│         Infrastructure Layer            │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │PostgreSQL│ │  Redis  │ │ Kafka   │   │
│  │         │ │ (Cache) │ │(Events) │   │
│  └─────────┘ └─────────┘ └─────────┘   │
└─────────────────────────────────────────┘
```

## New Features

### 1. Event-Driven Architecture

Domain events are published when entities are created, updated, or deleted:

- `ProductCreated` - Published when a new product is created
- `ProductUpdated` - Published when a product is updated
- `ProductDeleted` - Published when a product is deleted
- `SellerCreated` - Published when a new seller is created
- `SellerUpdated` - Published when a seller is updated
- `SellerDeleted` - Published when a seller is deleted

### 2. gRPC APIs

Protocol Buffer definitions are in the `proto/` directory:

- `product.proto` - Product service definitions
- `seller.proto` - Seller service definitions

Generated Go code is also in the `proto/` directory.

### 3. Redis Caching

Repository decorators add caching capabilities:

- `CachedProductRepository` - Caches products with TTL
- `CachedSellerRepository` - Caches sellers with TTL
- Cache invalidation on write operations

### 4. Kafka Event Bus

- **Publisher**: `KafkaEventPublisher` publishes domain events to Kafka topics
- **Subscriber**: `KafkaEventSubscriber` consumes events from Kafka topics
- **Topics**: Events are published to topics prefixed with `marketplace.`

## Prerequisites

1. **Go 1.23+**
2. **PostgreSQL** - for data persistence
3. **Redis** - for caching (optional, gracefully degrades)
4. **Kafka** - for event streaming (optional, gracefully degrades)
5. **Protocol Buffers compiler** (`protoc`)

## Setup Instructions

### 1. Install Dependencies

```bash
# Install protoc-gen-go tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Download dependencies
go mod tidy
```

### 2. Generate Protocol Buffers

```bash
chmod +x generate_proto.sh
./generate_proto.sh
```

### 3. Start Infrastructure Services

#### PostgreSQL

```bash
# Using Docker
docker run --name postgres-ddd \
  -e POSTGRES_USER=gorm \
  -e POSTGRES_PASSWORD=gorm \
  -e POSTGRES_DB=gorm \
  -p 9920:5432 \
  -d postgres:13
```

#### Redis (Optional)

```bash
# Using Docker
docker run --name redis-ddd \
  -p 6379:6379 \
  -d redis:7-alpine
```

#### Kafka (Optional)

```bash
# Using Docker Compose
cat > docker-compose.kafka.yml << EOF
version: '3.8'
services:
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
EOF

docker-compose -f docker-compose.kafka.yml up -d
```

### 4. Run the Application

```bash
go run cmd/marketplace/main.go
```

The application will start:

- **HTTP Server**: `http://localhost:8080`
- **gRPC Server**: `localhost:8090`

## API Usage

### REST APIs

Create a seller:

```bash
curl -X POST http://localhost:8080/api/v1/sellers \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe"}'
```

Create a product:

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"name": "iPhone 15", "price": 999.99, "seller_id": "SELLER_ID_HERE"}'
```

List products:

```bash
curl http://localhost:8080/api/v1/products
```

### gRPC APIs

You can use tools like [grpcurl](https://github.com/fullstorydev/grpcurl) or [Evans](https://github.com/ktr0731/evans) to test gRPC APIs:

```bash
# List services
grpcurl -plaintext localhost:8090 list

# Call ProductService.ListProducts
grpcurl -plaintext localhost:8090 marketplace.v1.ProductService/ListProducts
```

## Configuration

The application can be configured through environment variables or by modifying the `Config` struct in `main.go`:

```go
config := Config{
    Database: DatabaseConfig{
        DSN: "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable",
    },
    Redis: RedisConfig{
        Host:       "localhost",
        Port:       6379,
        // ...
    },
    Kafka: KafkaConfig{
        Brokers:     []string{"localhost:9092"},
        TopicPrefix: "marketplace",
        // ...
    },
    // ...
}
```

## Event Handling

The application includes an example Kafka consumer that logs received events. In a real-world scenario, you would:

1. **Update search indices** when products are created/updated
2. **Send notifications** when orders are placed
3. **Update analytics** based on user actions
4. **Trigger workflows** for business processes

## Caching Strategy

The Redis cache uses a write-through strategy:

- **Read**: Check cache first, fallback to database
- **Write**: Update database, then update cache
- **Delete**: Remove from database, then invalidate cache

Cache keys follow these patterns:

- Products: `marketplace:product:{id}`
- Sellers: `marketplace:seller:{id}`
- Lists: `marketplace:products:all`, `marketplace:sellers:all`

## Monitoring and Observability

For production use, consider adding:

1. **Metrics** - Prometheus metrics for latency, throughput, error rates
2. **Tracing** - OpenTelemetry for distributed tracing
3. **Logging** - Structured logging with correlation IDs
4. **Health checks** - HTTP endpoints for service health

## Testing

Run tests:

```bash
go test ./...
```

For integration tests with real infrastructure:

```bash
# Set up test dependencies first
docker-compose up -d postgres redis kafka

# Run integration tests
go test -tags=integration ./...
```

## Production Considerations

1. **Security**

   - Enable TLS for gRPC
   - Add authentication/authorization
   - Validate all inputs
   - Use secrets management

2. **Scalability**

   - Implement connection pooling
   - Add load balancing
   - Use Redis Cluster for cache
   - Partition Kafka topics

3. **Reliability**

   - Add circuit breakers
   - Implement retry policies
   - Set up monitoring alerts
   - Plan for disaster recovery

4. **Performance**
   - Optimize database queries
   - Tune cache TTL values
   - Configure Kafka batching
   - Profile application bottlenecks

## Contributing

1. Follow the DDD principles
2. Add tests for new features
3. Update documentation
4. Use conventional commit messages
5. Ensure backward compatibility for APIs
