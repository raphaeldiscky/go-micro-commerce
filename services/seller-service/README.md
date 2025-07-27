# Seller Service

A microservice for managing sellers in the marketplace application, built with Domain-Driven Design (DDD) principles.

## Architecture

This service follows clean architecture and DDD patterns with the following layers:

- **Domain Layer**: Contains business entities, domain events, and repository interfaces
- **Application Layer**: Contains services, DTOs, and business logic orchestration
- **Infrastructure Layer**: Contains database implementations, messaging, and external integrations
- **Interface Layer**: Contains HTTP REST API handlers and server configuration

## Features

- ✅ CRUD operations for sellers
- ✅ Email uniqueness validation
- ✅ Seller activation/deactivation
- ✅ PostgreSQL database with pgx driver
- ✅ Kafka event publishing for domain events
- ✅ RESTful HTTP API
- ✅ Clean architecture with dependency injection
- ✅ Docker containerization
- ✅ Graceful shutdown

## API Endpoints

### Sellers

- `POST /api/v1/sellers` - Create a new seller
- `GET /api/v1/sellers` - Get all sellers (with pagination and filtering)
- `GET /api/v1/sellers/{id}` - Get a seller by ID
- `GET /api/v1/sellers/email/{email}` - Get a seller by email
- `PUT /api/v1/sellers/{id}` - Update a seller
- `PATCH /api/v1/sellers/{id}/status` - Update seller status (activate/deactivate)
- `DELETE /api/v1/sellers/{id}` - Delete a seller

### Health Checks

- `GET /health` - Health check endpoint

## Configuration

The service is configured via environment variables:

| Variable        | Default               | Description            |
| --------------- | --------------------- | ---------------------- |
| `HTTP_PORT`     | `8081`                | HTTP server port       |
| `DB_HOST`       | `localhost`           | Database host          |
| `DB_PORT`       | `5432`                | Database port          |
| `DB_USER`       | `postgres`            | Database user          |
| `DB_PASSWORD`   | ``                    | Database password      |
| `DB_NAME`       | `marketplace_sellers` | Database name          |
| `DB_SSL_MODE`   | `disable`             | Database SSL mode      |
| `KAFKA_BROKERS` | `localhost:9092`      | Kafka broker addresses |
| `KAFKA_TOPIC`   | `seller-events`       | Kafka topic for events |

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
export DB_NAME=marketplace_sellers
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

- Seller service on port 8081
- PostgreSQL on port 5433
- Kafka and Zookeeper

### Database Schema

The service automatically creates the required table on startup:

```sql
CREATE TABLE IF NOT EXISTS sellers (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(254) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL,
    address VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

## Domain Events

The service publishes the following domain events to Kafka:

- `SellerCreated` - When a new seller is created
- `SellerUpdated` - When a seller is updated
- `SellerActivated` - When a seller is activated
- `SellerDeactivated` - When a seller is deactivated
- `SellerDeleted` - When a seller is deleted

## Business Rules

- **Email Uniqueness**: Each seller must have a unique email address
- **Validation**: Name (2-100 chars), Email (valid format), Phone (10-20 chars), Address (10-255 chars)
- **Status Management**: Sellers can be activated/deactivated
- **Data Integrity**: All required fields must be provided

## Testing

Run tests with:

```bash
go test ./...
```

## Project Structure

```
services/seller-service/
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
