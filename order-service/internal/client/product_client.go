// Package client provides a gRPC client for interacting with the product service.
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/grpc"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/emptypb"

	pkgConfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/product"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// ProductReservationItem represents a product reservation request.
type ProductReservationItem struct {
	ProductID       uuid.UUID
	Quantity        int64
	ExpectedVersion int64
}

// ProductRestorationItem represents a product restoration request.
type ProductRestorationItem struct {
	ProductID uuid.UUID
	Quantity  int64
}

// ProductClientInterface defines methods available for fetching products.
type ProductClientInterface interface {
	GetProducts(ctx context.Context, ids []uuid.UUID) ([]entity.Product, error)
	ReserveProducts(
		ctx context.Context,
		idempotencyKey uuid.UUID,
		items []ProductReservationItem,
	) ([]entity.Product, error)
	ReleaseProducts(
		ctx context.Context,
		reservationID uuid.UUID,
		items []ProductReservationItem,
	) error
	DeductProducts(
		ctx context.Context,
		reservationID uuid.UUID,
		items []ProductReservationItem,
	) error
	RestoreProducts(
		ctx context.Context,
		items []ProductRestorationItem,
	) ([]entity.Product, error)
	HealthCheck(ctx context.Context) error
	Close() error
}

// ProductClient is a gRPC client for interacting with the product service.
type ProductClient struct {
	grpcClient *grpc.Client
	client     pb.ProductServiceClient
}

// NewProductClient creates a new ProductClient instance with gRPC connection.
func NewProductClient(
	clientCfg *config.ClientConfig,
	consulCfg *config.ConsulConfig,
) (ProductClientInterface, error) {
	// Create gRPC client configuration
	grpcConfig := pkgConfig.DefaultGRPCClientConfig("product-service-grpc")

	// Configure based on existing client config
	grpcConfig.UseServiceDiscovery = clientCfg.UseServiceDiscovery
	grpcConfig.ConsulEnabled = consulCfg.Enabled
	grpcConfig.ConsulAddress = consulCfg.Address
	grpcConfig.SetStaticAddress(clientCfg.ProductGRPCHost, clientCfg.ProductGRPCPort)

	gClient, err := grpc.NewGRPCClient(grpcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create the product service client
	client := pb.NewProductServiceClient(gClient.GetConnection())

	return &ProductClient{
		grpcClient: gClient,
		client:     client,
	}, nil
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

	resp, err := pc.client.GetProducts(ctx, &pb.GetProductsRequest{Ids: stringIDs})
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
			ID:               uid,
			Name:             p.Name,
			Price:            decimal.NewFromFloat(p.Price), // safely convert double → decimal
			Quantity:         p.Quantity,
			Version:          p.Version,
			ReservedQuantity: p.ReservedQuantity,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		}
	}

	return products, nil
}

// ReserveProducts reserves stock for products with optimistic locking.
func (pc *ProductClient) ReserveProducts(
	ctx context.Context,
	idempotencyKey uuid.UUID,
	items []ProductReservationItem,
) ([]entity.Product, error) {
	// Convert to protobuf format
	pbItems := make([]*pb.ProductQuantity, len(items))
	for i, item := range items {
		pbItems[i] = &pb.ProductQuantity{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
			Version:   item.ExpectedVersion,
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := pc.client.ReserveProducts(ctx, &pb.ReserveProductsRequest{
		IdempotencyKey: idempotencyKey.String(),
		Items:          pbItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call ReserveProducts: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("reservation failed: %s", resp.ErrorMessage)
	}

	// Convert response to entities
	products := make([]entity.Product, len(resp.ReservedProducts))

	for i, p := range resp.ReservedProducts {
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
			ID:               uid,
			Name:             p.Name,
			Price:            decimal.NewFromFloat(p.Price),
			Quantity:         p.Quantity,
			Version:          p.Version,
			ReservedQuantity: p.ReservedQuantity,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		}
	}

	return products, nil
}

// ReleaseProducts releases reserved stock for products.
func (pc *ProductClient) ReleaseProducts(
	ctx context.Context,
	reservationID uuid.UUID,
	items []ProductReservationItem,
) error {
	// Convert to protobuf format
	pbItems := make([]*pb.ProductQuantity, len(items))
	for i, item := range items {
		pbItems[i] = &pb.ProductQuantity{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := pc.client.ReleaseProducts(ctx, &pb.ReleaseProductsRequest{
		ReservationId: reservationID.String(),
		Items:         pbItems,
	})
	if err != nil {
		return fmt.Errorf("failed to call ReleaseProducts: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("products release failed: %s", resp.ErrorMessage)
	}

	return nil
}

// DeductProducts confirms the stock deduction for a reservation.
func (pc *ProductClient) DeductProducts(
	ctx context.Context,
	reservationID uuid.UUID,
	items []ProductReservationItem,
) error {
	// Convert to protobuf format
	pbItems := make([]*pb.ProductQuantity, len(items))
	for i, item := range items {
		pbItems[i] = &pb.ProductQuantity{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := pc.client.DeductProducts(ctx, &pb.DeductProductsRequest{
		ReservationId: reservationID.String(),
		Items:         pbItems,
	})
	if err != nil {
		return fmt.Errorf("failed to call DeductProducts: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("stocks deduction confirmation failed: %s", resp.ErrorMessage)
	}

	return nil
}

// RestoreProducts restores stock in case of compensation.
func (pc *ProductClient) RestoreProducts(
	ctx context.Context,
	items []ProductRestorationItem,
) ([]entity.Product, error) {
	// Convert to protobuf format
	pbItems := make([]*pb.ProductQuantity, len(items))
	for i, item := range items {
		pbItems[i] = &pb.ProductQuantity{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := pc.client.RestoreProducts(ctx, &pb.RestoreProductsRequest{
		Items:  pbItems,
		Reason: "order_compensation", // Standard reason for saga compensation
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call RestoreProducts: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("products restoration failed: %s", resp.ErrorMessage)
	}

	// Convert response to entities
	products := make([]entity.Product, len(resp.RestoredProducts))

	for i, p := range resp.RestoredProducts {
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
			ID:               uid,
			Name:             p.Name,
			Price:            decimal.NewFromFloat(p.Price),
			Quantity:         p.Quantity,
			Version:          p.Version,
			ReservedQuantity: p.ReservedQuantity,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		}
	}

	return products, nil
}

// HealthCheck verifies the connection to product-service.
func (pc *ProductClient) HealthCheck(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := pc.client.Health(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != constant.GRPCHealthServing {
		return fmt.Errorf("service unhealthy: %s", resp.Status)
	}

	return nil
}

// Close closes the gRPC connection.
func (pc *ProductClient) Close() error {
	return pc.grpcClient.Close()
}
