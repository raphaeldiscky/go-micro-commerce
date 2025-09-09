// Package client provides a gRPC client for interacting with the product service.
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pkgconfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"
	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/product"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mapper"
)

// ProductClientInterface defines methods available for fetching products.
type ProductClientInterface interface {
	GetProducts(ctx context.Context, ids []uuid.UUID) ([]entity.Product, error)
	ReserveProducts(
		ctx context.Context,
		idempotencyKey uuid.UUID,
		items []dto.ProductReservationItem,
	) (reservedProducts []entity.Product, err error)
	ReleaseProducts(
		ctx context.Context,
		items []dto.ProductReservationItem,
	) error
	ConfirmProductsDeduction(
		ctx context.Context,
		items []dto.ProductReservationItem,
	) (products []entity.Product, err error)
	RestoreProducts(
		ctx context.Context,
		items []dto.ProductRestorationItem,
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
	cfg *config.Config,
) (ProductClientInterface, error) {
	// Create gRPC client configuration
	grpcConfig := pkgconfig.DefaultGRPCClientConfig(pkgconstant.GRPCServiceNameProduct)

	// Configure based on existing client config
	grpcConfig.UseServiceDiscovery = cfg.Client.UseServiceDiscovery
	grpcConfig.ConsulEnabled = cfg.Consul.Enabled
	grpcConfig.ConsulAddress = cfg.Consul.Address
	grpcConfig.SetStaticAddress(cfg.Client.ProductGRPCHost, cfg.Client.ProductGRPCPort)

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
		product, err := mapper.MapProtoToProduct(p)
		if err != nil {
			return nil, err
		}

		products[i] = product
	}

	return products, nil
}

// ReserveProducts reserves stock for products with optimistic locking.
func (pc *ProductClient) ReserveProducts(
	ctx context.Context,
	idempotencyKey uuid.UUID,
	items []dto.ProductReservationItem,
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
		return nil, fmt.Errorf("failed to call ReserveProducts gRPC: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("reservation failed from product-service: %s", resp.ErrorMessage)
	}

	// Convert response to entities
	products := make([]entity.Product, len(resp.ReservedProducts))

	for i, p := range resp.ReservedProducts {
		product, err := mapper.MapProtoToProduct(p)
		if err != nil {
			return nil, err
		}

		products[i] = product
	}

	return products, nil
}

// ReleaseProducts releases reserved stock for products.
func (pc *ProductClient) ReleaseProducts(
	ctx context.Context,
	items []dto.ProductReservationItem,
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
		Items: pbItems,
	})
	if err != nil {
		return fmt.Errorf("failed to call ReleaseProducts: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("products release failed: %s", resp.ErrorMessage)
	}

	return nil
}

// ConfirmProductsDeduction confirms the reserved stock and removes reserved quantity.
func (pc *ProductClient) ConfirmProductsDeduction(
	ctx context.Context,
	items []dto.ProductReservationItem,
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

	resp, err := pc.client.ConfirmProductsDeduction(ctx, &pb.ConfirmProductsDeductionRequest{
		Items: pbItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call ConfirmProductsDeduction: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("stocks deduction confirmation failed: %s", resp.ErrorMessage)
	}

	products := make([]entity.Product, len(resp.UpdatedProducts))

	for i, p := range resp.UpdatedProducts {
		product, err := mapper.MapProtoToProduct(p)
		if err != nil {
			return nil, err
		}

		products[i] = product
	}

	return products, nil
}

// RestoreProducts restores stock in case of compensation.
func (pc *ProductClient) RestoreProducts(
	ctx context.Context,
	items []dto.ProductRestorationItem,
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
		product, err := mapper.MapProtoToProduct(p)
		if err != nil {
			return nil, err
		}

		products[i] = product
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

	if resp.Status != pkgconstant.GRPCHealthServing {
		return fmt.Errorf("service unhealthy: %s", resp.Status)
	}

	return nil
}

// Close closes the gRPC connection.
func (pc *ProductClient) Close() error {
	return pc.grpcClient.Close()
}
