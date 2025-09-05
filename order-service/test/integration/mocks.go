package integration

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/saga"
)

// MockProductGRPCClient is a mock implementation of ProductClientInterface for testing.
type MockProductGRPCClient struct{}

// NewMockProductGRPCClient creates a new mock product gRPC client.
func NewMockProductGRPCClient() client.ProductClientInterface {
	return &MockProductGRPCClient{}
}

// GetProducts returns mock product data for testing.
func (m *MockProductGRPCClient) GetProducts(
	_ context.Context,
	productIDs []uuid.UUID,
) ([]entity.Product, error) {
	products := make([]entity.Product, 0, len(productIDs))

	for _, id := range productIDs {
		// Create mock product entity
		product := entity.Product{
			ID:        id,
			Name:      "Test Product",
			UnitPrice: decimal.NewFromFloat(99.99),
			Quantity:  100,
		}
		products = append(products, product)
	}

	return products, nil
}

// ReserveProducts simulates successful product reservation.
func (m *MockProductGRPCClient) ReserveProducts(
	_ context.Context,
	_ uuid.UUID,
	_ []dto.ProductReservationItem,
) ([]entity.Product, error) {
	return []entity.Product{}, nil
}

// ReleaseProducts simulates successful product release.
func (m *MockProductGRPCClient) ReleaseProducts(
	_ context.Context,
	_ []dto.ProductReservationItem,
) error {
	return nil
}

// ConfirmProductsDeduction simulates successful product deduction confirmation.
func (m *MockProductGRPCClient) ConfirmProductsDeduction(
	_ context.Context,
	_ []dto.ProductReservationItem,
) ([]entity.Product, error) {
	return []entity.Product{}, nil
}

// RestoreProducts simulates successful product restoration.
func (m *MockProductGRPCClient) RestoreProducts(
	_ context.Context,
	_ []dto.ProductRestorationItem,
) ([]entity.Product, error) {
	return []entity.Product{}, nil
}

// HealthCheck simulates successful health check.
func (m *MockProductGRPCClient) HealthCheck(_ context.Context) error {
	return nil
}

// Close simulates successful client close.
func (m *MockProductGRPCClient) Close() error {
	return nil
}

// MockKafkaProducer is a mock implementation of KafkaProducerInterface for testing.
type MockKafkaProducer struct{}

// Send simulates successful message send.
func (m *MockKafkaProducer) Send(_ context.Context, _ event.BaseEvent) error {
	// Do nothing - just simulate successful send for testing
	return nil
}

// Topic returns the topic for the mock producer.
func (m *MockKafkaProducer) Topic() string {
	return "test-topic"
}

// MockSagaOrchestrator is a mock implementation of saga.Orchestrator for testing.
type MockSagaOrchestrator struct{}

// NewMockSagaManager creates a new mock saga orchestrator.
func NewMockSagaManager() saga.Orchestrator {
	return saga.Orchestrator{} // Return empty struct since it's a concrete type
}

// NewMockSagaOrchestrator creates a new mock saga orchestrator.
func NewMockSagaOrchestrator() saga.Orchestrator {
	return saga.Orchestrator{} // Return empty struct
}

// MockTemporalClient is a mock implementation of TemporalClient for testing.
type MockTemporalClient struct{}

// NewMockTemporalClient creates a new mock temporal client.
func NewMockTemporalClient() *client.TemporalClient {
	return &client.TemporalClient{}
}

// MockOutboxPublisher is a mock implementation for outbox publishing.
type MockOutboxPublisher struct{}

// NewMockOutboxPublisher creates a new mock outbox publisher.
func NewMockOutboxPublisher() kafka.ProducerInterface {
	return &MockKafkaProducer{}
}
