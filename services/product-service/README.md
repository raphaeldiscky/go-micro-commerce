# Product Service

A microservice for managing products in the marketplace application, built with Domain-Driven Design (DDD) principles.

## Architecture

This service follows clean architecture and DDD patterns with the following layers:

- **Domain Layer**: Contains business entities, domain events, and repository interfaces
- **Application Layer**: Contains services, DTOs, and business logic orchestration
- **Infrastructure Layer**: Contains database implementations, messaging, and external integrations
- **Interface Layer**: Contains HTTP REST API handlers and server configuration

## Features

- ✅ CRUD operations for products
- ✅ PostgreSQL database with pgx driver
- ✅ Kafka event publishing for domain events
- ✅ RESTful HTTP API
- ✅ Clean architecture with dependency injection
- ✅ Docker containerization
- ✅ Graceful shutdown

## API Endpoints

### Products

- `POST /api/v1/products` - Create a new product
- `GET /api/v1/products` - Get all products (with pagination and filtering)
- `GET /api/v1/products/{id}` - Get a product by ID
- `PUT /api/v1/products/{id}` - Update a product
- `DELETE /api/v1/products/{id}` - Delete a product

### Health Checks

- `GET /health` - Health check endpoint

## Configuration

The service is configured via environment variables:

| Variable        | Default                | Description            |
| --------------- | ---------------------- | ---------------------- |
| `HTTP_PORT`     | `8080`                 | HTTP server port       |
| `DB_HOST`       | `localhost`            | Database host          |
| `DB_PORT`       | `5432`                 | Database port          |
| `DB_USER`       | `postgres`             | Database user          |
| `DB_PASSWORD`   | ``                     | Database password      |
| `DB_NAME`       | `marketplace_products` | Database name          |
| `DB_SSL_MODE`   | `disable`              | Database SSL mode      |
| `KAFKA_BROKERS` | `localhost:9092`       | Kafka broker addresses |
| `KAFKA_TOPIC`   | `product-events`       | Kafka topic for events |

## Development

### Prerequisites

- Go 1.23+
- PostgreSQL
- Kafka (optional, for event publishing)

### Running Locally

1. Install dependencies:

```bash
go mod download
```

2. Set up environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=marketplace_products
```

3. Run the service:

```bash
go run ./cmd/main.go
```

### Running with Docker Compose

1. Build and start all services:

```bash
docker-compose up --build
```

This will start:

- Product service on port 8080
- PostgreSQL on port 5432
- Kafka and Zookeeper

### Database Schema

The service automatically creates the required table on startup:

```sql
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    seller_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

## Domain Events

The service publishes the following domain events to Kafka:

- `ProductCreated` - When a new product is created
- `ProductUpdated` - When a product is updated
- `ProductDeleted` - When a product is deleted

## Testing

Run tests with:

```bash
go test ./...
```

## Project Structure

```
services/product-service/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── application/
│   │   ├── dto/                   # Data Transfer Objects
│   │   └── services/              # Application services
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── domain/
│   │   ├── entities/              # Domain entities
│   │   ├── events/                # Domain events
│   │   └── repositories/          # Repository interfaces
│   ├── infrastructure/
│   │   ├── database/              # Database connection
│   │   ├── messaging/             # Kafka implementation
│   │   └── persistence/           # Repository implementations
│   └── interface/
│       └── http/                  # HTTP handlers and server
├── docker-compose.yml             # Development environment
├── Dockerfile                     # Container definition
├── go.mod                        # Go module definition
└── README.md                     # This file
```

## Contributing

1. Follow Go best practices and conventions
2. Maintain clean architecture principles
3. Write tests for new functionality
4. Update documentation as needed
