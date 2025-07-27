# Go-DDD: Domain Driven Design Template

> A reference implementation demonstrating [Domain Driven Design (DDD)](https://en.wikipedia.org/wiki/Domain-driven_design) patterns in Go, featuring a simple marketplace where sellers can sell products.

## Table of Contents

- [Go-DDD: Domain Driven Design Template](#go-ddd-domain-driven-design-template)
  - [Table of Contents](#table-of-contents)
  - [Why Domain Driven Design?](#why-domain-driven-design)
  - [🚀 Quick Start](#-quick-start)
    - [Option 1: Microservices with Traefik API Gateway (Recommended)](#option-1-microservices-with-traefik-api-gateway-recommended)
    - [Option 2: Microservices with Custom API Gateway](#option-2-microservices-with-custom-api-gateway)
    - [Option 3: Monolithic Application (Legacy)](#option-3-monolithic-application-legacy)
  - [Architecture Overview](#architecture-overview)
  - [Project Structure](#project-structure)
  - [Layer Principles](#layer-principles)
    - [Domain Layer](#domain-layer)
    - [Application Layer](#application-layer)
    - [Infrastructure Layer](#infrastructure-layer)
    - [Interface Layer](#interface-layer)
  - [Best Practices](#best-practices)
    - [Repository Patterns](#repository-patterns)
    - [Data Management](#data-management)
  - [Development](#development)
    - [Mock Generation](#mock-generation)
    - [Available Commands](#available-commands)

## Why Domain Driven Design?

DDD helps build maintainable enterprise software by connecting implementation to business models:

**Ubiquitous Language**

- Common vocabulary between developers and stakeholders

**Clean Architecture**

- Business logic isolated from infrastructure concerns

**Scalability**

- Easier transition to microservices architecture

## 🚀 Quick Start

### Option 1: Microservices with Traefik API Gateway (Recommended)

```bash
# Start all services with Traefik
./scripts/microservices-traefik.sh start

# Check service status
./scripts/microservices-traefik.sh status

# Test the APIs
./deployments/traefik/test-api.sh
```

**Access Points:**

- **Traefik Dashboard:** http://localhost:8080
- **Product API:** http://localhost/api/v1/products
- **Seller API:** http://localhost/api/v1/sellers
- **Prometheus:** http://localhost:9091
- **Grafana:** http://localhost:3001 (admin/admin)
- **Kafka UI:** http://localhost:8081

### Option 2: Microservices with Custom API Gateway

```bash
# Start microservices
./scripts/microservices.sh start

# Check service health
./scripts/microservices.sh health
```

### Option 3: Monolithic Application (Legacy)

```bash
# Start infrastructure
docker-compose -f deployments/docker-compose/infra.yml up -d

# Run the application
make run
```

## Architecture Overview

This project follows the **Onion Architecture** pattern with clear layer separation:

```
┌─────────────────────────────────────┐
│            Interface Layer          │  ← REST APIs, gRPC
├─────────────────────────────────────┤
│          Application Layer          │  ← Use cases, Commands, Queries
├─────────────────────────────────────┤
│         Infrastructure Layer        │  ← Database, Cache, Messaging
├─────────────────────────────────────┤
│            Domain Layer             │  ← Business Logic & Rules
└─────────────────────────────────────┘
```

## Project Structure

```
go-ddd/
├── cmd/                   # Application entry points
│   └── marketplace/
├── internal/
│   ├── domain/            # Core business logic
│   │   ├── entities/      # Business entities (Product, Seller)
│   │   ├── events/        # Domain events
│   │   └── repositories/  # Repository interfaces
│   ├── application/       # Use cases and workflows
│   │   ├── command/       # Commands (Create, Update)
│   │   ├── query/         # Queries (Read operations)
│   │   └── services/      # Application services
│   ├── infrastructure/    # Technical implementations
│   │   ├── db/            # Database connections
│   │   ├── cache/         # Redis caching
│   │   ├── messaging/     # Kafka messaging
│   │   └── repository/    # Repository implementations
│   └── interface/         # External interfaces
│       ├── api/rest/      # REST endpoints
│       └── grpc/          # gRPC services
├── proto/                 # Protocol buffer definitions
└── deployments/           # Docker compositions
```

## Layer Principles

### Domain Layer

The core of the application containing pure business logic:

- ✅ **Independent**: No dependencies on other layers
- ✅ **Business Rules**: Implements all business logic and invariants
- ✅ **Entity Validation**: Validates business rules on creation/updates
- ✅ **Default Values**: Sets entity defaults (UUIDs, timestamps)
- ❌ **No Infrastructure**: Never accesses databases or external services

### Application Layer

Orchestrates domain operations and use cases:

- ✅ **Coordination**: Glue between domain and infrastructure
- ✅ **Use Cases**: Implements application-specific workflows
- ✅ **Transaction Management**: Handles cross-aggregate transactions
- ❌ **No Business Logic**: Delegates all business decisions to domain

### Infrastructure Layer

Handles technical concerns and external dependencies:

- ✅ **Repository Implementation**: Concrete data access implementations
- ✅ **External Services**: Database, cache, messaging integrations
- ✅ **Data Mapping**: Translates between domain and persistence models
- ✅ **Read After Write**: Always verify successful persistence
- ❌ **No Business Logic**: Pure technical implementation

### Interface Layer

Exposes application functionality to external consumers:

- ✅ **API Endpoints**: REST and gRPC service implementations
- ✅ **Input Validation**: Request format and basic validation
- ✅ **Response Mapping**: Converts domain results to API formats
- ❌ **No Domain Logic**: Thin layer that delegates to application services

## Best Practices

### Repository Patterns

**Read Operations**

- **Rule**: Return domain entities, not validated entities
- **Why**: Historical data compatibility - validations evolve over time

**Method Naming**

- **Rule**: `Find()` can return nil, `Get()` must return value or error
- **Why**: Clear contract expectations for callers

**After Write**

- **Rule**: Always read entity after persistence in the
- **Why**: Ensures data integrity and consistency

### Data Management

**Default Values**

- **Implementation**: Set in domain layer, never in database
- **Benefits**: Single source of truth, database independence

**Soft Deletion**

- **Implementation**: Use `deleted_at` timestamp column
- **Benefits**: Data recovery capability, audit trails

**Validation Timing**

- **Implementation**: Only on write operations (create/update)
- **Benefits**: Performance and backward compatibility

> 💡 **Pro Tip**: Start by exploring the `domain/entities` to understand the business model, then follow the data flow through application services to infrastructure implementations.

## Development

### Mock Generation

This project uses [Uber's mock](https://github.com/uber-go/mock) for generating type-safe mocks from interfaces:

**Generate all mocks:**

```bash
make mocks
```

**Generate mocks manually:**

```bash
go generate ./internal/application/interfaces/...
```

**Adding mock generation to new interfaces:**

Add the `//go:generate` directive to your interface file:

```go
package interfaces

//go:generate mockgen -source=your_service.go -destination=../../mocks/mock_your_service.go -package=mocks

type YourService interface {
    DoSomething() error
}
```

**Using generated mocks in tests:**

```go
func TestYourFunction(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockService := mocks.NewMockYourService(ctrl)
    mockService.EXPECT().DoSomething().Return(nil).Times(1)

    // Your test logic here
}
```

### Available Commands

- `make mocks` - Generate all mocks
- `make test` - Run all tests
- `make build` - Build the application
- `make proto` - Generate protobuf files
