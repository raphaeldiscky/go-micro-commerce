# Go DDD Microservices Marketplace

A marketplace application built with **Domain-Driven Design (DDD)** principles and **microservices architecture** using Go.

## 🏗️ Architecture Overview

The application has been transformed from a monolithic architecture to a microservices architecture with the following components:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │ Product Service │    │ Seller Service  │
│     :8000       │    │     :8080       │    │     :8081       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │     Kafka       │
                    │     :9092       │
                    └─────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Product DB     │    │   Seller DB     │    │     Redis       │
│    :5432        │    │    :5433        │    │    :6379        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🛠️ Services

### 1. **API Gateway** (Port 8000)

- **Purpose**: Single entry point for all client requests
- **Features**:
  - Request routing to appropriate services
  - Load balancing and service discovery
  - Request/response logging and monitoring
  - CORS handling
  - Aggregated endpoints for marketplace statistics

### 2. **Product Service** (Port 8080)

- **Purpose**: Manages product catalog and inventory
- **Features**:
  - CRUD operations for products
  - Product categorization and search
  - Integration with seller service via events
  - PostgreSQL database (marketplace_products)
  - Kafka event publishing

### 3. **Seller Service** (Port 8081)

- **Purpose**: Manages seller profiles and onboarding
- **Features**:
  - Seller registration and profile management
  - Email uniqueness validation
  - Seller activation/deactivation
  - PostgreSQL database (marketplace_sellers)
  - Kafka event publishing

## 🔧 Technology Stack

- **Language**: Go 1.23
- **Framework**: Echo (HTTP framework)
- **Database**: PostgreSQL with pgx driver
- **Message Broker**: Apache Kafka
- **Caching**: Redis
- **Containerization**: Docker & Docker Compose
- **Architecture**: Clean Architecture + DDD

## 🚀 Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local development)

### Running the Microservices

1. **Clone the repository**:

```bash
git clone <repository-url>
cd go-ddd
```

2. **Start all services**:

```bash
docker-compose -f docker-compose.microservices.yml up --build
```

3. **Verify services are running**:

```bash
# Check API Gateway
curl http://localhost:8000/health

# Check Product Service
curl http://localhost:8080/health

# Check Seller Service
curl http://localhost:8081/health
```

### Service URLs

- **API Gateway**: http://localhost:8000
- **Product Service**: http://localhost:8080
- **Seller Service**: http://localhost:8081
- **Kafka UI**: http://localhost:9090
- **Product Database**: localhost:5432
- **Seller Database**: localhost:5433
- **Redis**: localhost:6379

## 📡 API Endpoints

### Via API Gateway (Recommended)

#### Products

- `POST /api/v1/products` - Create product
- `GET /api/v1/products` - List products
- `GET /api/v1/products/{id}` - Get product
- `PUT /api/v1/products/{id}` - Update product
- `DELETE /api/v1/products/{id}` - Delete product

#### Sellers

- `POST /api/v1/sellers` - Create seller
- `GET /api/v1/sellers` - List sellers
- `GET /api/v1/sellers/{id}` - Get seller
- `GET /api/v1/sellers/email/{email}` - Get seller by email
- `PUT /api/v1/sellers/{id}` - Update seller
- `PATCH /api/v1/sellers/{id}/status` - Update seller status
- `DELETE /api/v1/sellers/{id}` - Delete seller

#### Marketplace Aggregation

- `GET /api/v1/marketplace/stats` - Get marketplace statistics
- `GET /api/v1/marketplace/sellers/{seller_id}/products` - Get products by seller

### Direct Service Access

You can also access services directly:

- **Product Service**: http://localhost:8080/api/v1/products
- **Seller Service**: http://localhost:8081/api/v1/sellers

## 🔄 Event-Driven Communication

Services communicate via Kafka events:

### Product Events (Topic: `product-events`)

- `ProductCreated` - When a product is created
- `ProductUpdated` - When a product is updated
- `ProductDeleted` - When a product is deleted

### Seller Events (Topic: `seller-events`)

- `SellerCreated` - When a seller is created
- `SellerUpdated` - When a seller is updated
- `SellerActivated` - When a seller is activated
- `SellerDeactivated` - When a seller is deactivated
- `SellerDeleted` - When a seller is deleted

## 📊 Monitoring & Observability

### Kafka UI

Access Kafka UI at http://localhost:9090 to monitor:

- Topics and partitions
- Consumer groups
- Message flow
- Broker health

### Logs

View service logs using Docker Compose:

```bash
# All services
docker-compose -f docker-compose.microservices.yml logs -f

# Specific service
docker-compose -f docker-compose.microservices.yml logs -f product-service
```

### Health Checks

Each service exposes a `/health` endpoint:

```bash
curl http://localhost:8000/health  # API Gateway
curl http://localhost:8080/health  # Product Service
curl http://localhost:8081/health  # Seller Service
```

## 🧪 Testing

### Integration Testing

```bash
# Create a seller
curl -X POST http://localhost:8000/api/v1/sellers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "1234567890",
    "address": "123 Main St, City, State"
  }'

# Create a product
curl -X POST http://localhost:8000/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15",
    "price": 999.99,
    "seller_id": "seller-id-from-above"
  }'

# Get marketplace stats
curl http://localhost:8000/api/v1/marketplace/stats
```

## 🏁 Development

### Running Individual Services

#### Product Service

```bash
cd services/product-service
go run ./cmd/main.go
```

#### Seller Service

```bash
cd services/seller-service
go run ./cmd/main.go
```

#### API Gateway

```bash
go run ./cmd/api-gateway/main.go
```

### Environment Variables

Each service can be configured via environment variables. See individual service READMEs for details:

- [Product Service README](./services/product-service/README.md)
- [Seller Service README](./services/seller-service/README.md)

## 🔧 Configuration

### Docker Compose Environment Variables

The main configuration is in `docker-compose.microservices.yml`. Key environment variables:

- `PRODUCT_SERVICE_URL` - Product service URL for API Gateway
- `SELLER_SERVICE_URL` - Seller service URL for API Gateway
- `KAFKA_BROKERS` - Kafka broker addresses
- `DB_*` - Database connection settings

### Production Considerations

For production deployment, consider:

1. **Service Discovery**: Replace hardcoded URLs with service discovery
2. **Load Balancing**: Add multiple instances behind load balancers
3. **Security**: Add authentication, authorization, and TLS
4. **Monitoring**: Add Prometheus metrics and distributed tracing
5. **Database Migration**: Implement proper database migration strategies
6. **Circuit Breakers**: Add resilience patterns for service communication

## 📁 Project Structure

```
go-ddd/
├── cmd/
│   └── api-gateway/              # API Gateway entry point
├── services/
│   ├── product-service/          # Product microservice
│   └── seller-service/           # Seller microservice
├── internal/
│   └── api-gateway/              # API Gateway implementation
├── deployments/
│   └── docker/                   # Docker configurations
├── docker-compose.microservices.yml  # Full microservices setup
└── README.md                     # This file
```

## 🎯 Benefits Achieved

✅ **Service Independence**: Each service can be developed, deployed, and scaled independently  
✅ **Technology Flexibility**: Services can use different technologies as needed  
✅ **Fault Isolation**: Issues in one service don't affect others  
✅ **Team Autonomy**: Different teams can own different services  
✅ **Scalability**: Independent scaling based on service-specific needs  
✅ **Event-Driven Architecture**: Loose coupling through asynchronous events

## 🔮 Future Enhancements

- [ ] Add authentication service (JWT/OAuth2)
- [ ] Implement API rate limiting
- [ ] Add distributed tracing (Jaeger/Zipkin)
- [ ] Implement CQRS pattern for read/write separation
- [ ] Add event sourcing for audit trails
- [ ] Implement saga pattern for distributed transactions
- [ ] Add service mesh (Istio/Linkerd)
- [ ] Implement blue-green deployments

## 🤝 Contributing

1. Follow Go best practices and conventions
2. Maintain clean architecture principles
3. Write tests for new functionality
4. Update documentation as needed
5. Ensure backward compatibility in APIs
