// Package client provides a gRPC client for interacting with the product service.
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/proto/product"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/entity"
)

// ProductClientInterface defines methods available for fetching products.
type ProductClientInterface interface {
	GetProducts(ctx context.Context, ids []uuid.UUID) ([]entity.Product, error)
	HealthCheck(ctx context.Context) error
	Close() error
}

// ProductClient is a gRPC client for interacting with the product service.
type ProductClient struct {
	conn   *grpc.ClientConn
	client product.ProductServiceClient
}

// NewProductClient creates a new gRPC client for product-service.
func NewProductClient(address string) (ProductClientInterface, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product-service: %w", err)
	}

	client := product.NewProductServiceClient(conn)

	return &ProductClient{conn: conn, client: client}, nil
}

// GetProducts fetches product data by IDs.
func (pc *ProductClient) GetProducts(
	ctx context.Context,
	ids []uuid.UUID,
) ([]entity.Product, error) {
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := pc.client.GetProducts(ctx, &product.GetProductsRequest{Ids: stringIDs})
	if err != nil {
		return nil, fmt.Errorf("failed to call GetProducts: %w", err)
	}

	products := make([]entity.Product, len(resp.Products))

	for i, p := range resp.Products {
		uid, err := uuid.Parse(p.Id)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID from product-service: %w", err)
		}

		// Convert protobuf Timestamp → time.Time safely
		var createdAt, updatedAt time.Time
		if p.CreatedAt != nil {
			createdAt = p.CreatedAt.AsTime()
		}

		if p.UpdatedAt != nil {
			updatedAt = p.UpdatedAt.AsTime()
		}

		products[i] = entity.Product{
			ID:        uid,
			Name:      p.Name,
			Price:     decimal.NewFromFloat(p.Price), // safely convert double → decimal
			Quantity:  int(p.Quantity),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
	}

	return products, nil
}

// Close closes the gRPC connection.
func (pc *ProductClient) Close() error {
	return pc.conn.Close()
}

// HealthCheck verifies the connection to product-service.
func (pc *ProductClient) HealthCheck(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// You might want to add a Health method to your proto service
	// For now, we'll just check if the connection is ready
	state := pc.conn.GetState()
	if state != connectivity.Ready {
		return fmt.Errorf("connection not ready, state: %s", state.String())
	}

	return nil
}
