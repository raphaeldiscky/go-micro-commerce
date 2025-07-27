# Go DDD Marketplace Makefile

# Variables
APP_NAME := marketplace
GO_VERSION := 1.23
PROTOC_VERSION := 3.21.12
DOCKER_COMPOSE := docker-compose
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
PROTO_FILES := $(shell find proto -name '*.proto')

# Default target
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
.PHONY: deps
deps: ## Install Go dependencies
	go mod tidy
	go mod vendor

.PHONY: proto
proto: ## Generate protobuf files
	@echo "Generating protobuf files..."
	@chmod +x generate_proto.sh
	@./generate_proto.sh

.PHONY: mocks
mocks: ## Generate mocks using uber-go/mock
	@echo "Generating mocks..."
	go generate ./internal/application/interfaces/...
	@echo "Mocks generated successfully!"

.PHONY: build
build: proto ## Build the application
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) cmd/marketplace/main.go

.PHONY: run
run: ## Run the application
	@echo "Running $(APP_NAME)..."
	go run cmd/marketplace/main.go

.PHONY: dev
dev: infrastructure proto ## Start development environment
	@echo "Starting development server..."
	air -c .air.toml || go run cmd/marketplace/main.go

# Testing
.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-race
test-race: ## Run tests with race detection
	go test -race -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: benchmark
benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...

# Code quality
.PHONY: lint
lint: ## Run linters
	golangci-lint run

.PHONY: fmt
fmt: ## Format Go code
	gofmt -s -w $(GO_FILES)
	goimports -w $(GO_FILES)

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: check
check: fmt vet lint test ## Run all code quality checks

# Infrastructure
.PHONY: infrastructure
infrastructure: ## Start infrastructure services (PostgreSQL, Redis, Kafka)
	@echo "Starting infrastructure services..."
	$(DOCKER_COMPOSE) up -d postgres redis kafka
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Infrastructure is ready!"

.PHONY: infrastructure-all
infrastructure-all: ## Start all infrastructure services including UI
	@echo "Starting all infrastructure services..."
	$(DOCKER_COMPOSE) up -d
	@echo "Services started!"
	@echo "PostgreSQL: localhost:9920"
	@echo "Redis: localhost:6379"
	@echo "Kafka: localhost:9092"
	@echo "Kafka UI: http://localhost:8090"

.PHONY: infrastructure-stop
infrastructure-stop: ## Stop infrastructure services
	@echo "Stopping infrastructure services..."
	$(DOCKER_COMPOSE) down

.PHONY: infrastructure-logs
infrastructure-logs: ## Show infrastructure logs
	$(DOCKER_COMPOSE) logs -f

.PHONY: infrastructure-reset
infrastructure-reset: ## Reset infrastructure (delete all data)
	@echo "Resetting infrastructure..."
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d postgres redis kafka
	@sleep 10

# Database
.PHONY: migrate-build
migrate-build: ## Build migration CLI tool
	@echo "Building migration tool..."
	go build -o bin/migrate cmd/migrate/main.go

.PHONY: migrate-up
migrate-up: migrate-build ## Run database migrations up
	@echo "Running database migrations..."
	./bin/migrate -database-url="$(DATABASE_URL)" -action=up

.PHONY: migrate-down
migrate-down: migrate-build ## Rollback database migrations
	@echo "Rolling back database migrations..."
	./bin/migrate -database-url="$(DATABASE_URL)" -action=down -steps=$(STEPS)

.PHONY: migrate-version
migrate-version: migrate-build ## Show current migration version
	@echo "Getting migration version..."
	./bin/migrate -database-url="$(DATABASE_URL)" -action=version

.PHONY: migrate-create
migrate-create: ## Create new migration files (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then echo "Please provide NAME: make migrate-create NAME=migration_name"; exit 1; fi
	@echo "Creating migration files for: $(NAME)"
	@timestamp=$$(date +%s); \
	seq=$$(printf "%06d" $$timestamp); \
	touch migrations/$${seq}_$(NAME).up.sql; \
	touch migrations/$${seq}_$(NAME).down.sql; \
	echo "Created migrations/$${seq}_$(NAME).up.sql"; \
	echo "Created migrations/$${seq}_$(NAME).down.sql"

.PHONY: db-migrate
db-migrate: migrate-up ## Alias for migrate-up

.PHONY: db-seed
db-seed: ## Seed database with test data
	@echo "Seeding database..."
	go run scripts/seed.go

# gRPC
.PHONY: grpc-health
grpc-health: ## Check gRPC server health
	grpcurl -plaintext localhost:8090 list

.PHONY: grpc-list-products
grpc-list-products: ## List products via gRPC
	grpcurl -plaintext localhost:8090 marketplace.v1.ProductService/ListProducts

.PHONY: grpc-list-sellers
grpc-list-sellers: ## List sellers via gRPC
	grpcurl -plaintext localhost:8090 marketplace.v1.SellerService/ListSellers

# REST API
.PHONY: api-health
api-health: ## Check REST API health
	curl -f http://localhost:8080/health || echo "API not responding"

.PHONY: api-list-products
api-list-products: ## List products via REST API
	curl -s http://localhost:8080/api/v1/products | jq .

.PHONY: api-list-sellers
api-list-sellers: ## List sellers via REST API
	curl -s http://localhost:8080/api/v1/sellers | jq .

# Deployment
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(APP_NAME):latest .

.PHONY: docker-run
docker-run: docker-build ## Run application in Docker
	docker run -p 8080:8080 -p 8090:8090 $(APP_NAME):latest

# Cleanup
.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf vendor/
	rm -f coverage.out coverage.html
	go clean -cache
	go clean -modcache

.PHONY: clean-all
clean-all: clean infrastructure-stop ## Clean everything including infrastructure
	docker system prune -f
	$(DOCKER_COMPOSE) down -v --remove-orphans

# Install tools
.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/cosmtrek/air@latest

# Documentation
.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	godoc -http=:6060
	@echo "Documentation server started at http://localhost:6060"

# Quick start
.PHONY: quick-start
quick-start: install-tools infrastructure proto build ## Quick start for new developers
	@echo ""
	@echo "🎉 Setup complete!"
	@echo ""
	@echo "To start the application:"
	@echo "  make run"
	@echo ""
	@echo "To run tests:"
	@echo "  make test"
	@echo ""
	@echo "To see all available commands:"
	@echo "  make help"

.DEFAULT_GOAL := help
