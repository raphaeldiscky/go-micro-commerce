package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/proto/product/v1/productv1connect"

	pkgconnect "github.com/raphaeldiscky/go-micro-commerce/pkg/connect"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/product/v1"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mapper"
)

// ProductClient defines methods available for fetching products.
type ProductClient interface {
	GetProducts(ctx context.Context, ids []uuid.UUID) ([]entity.Product, error)
	ReserveProducts(
		ctx context.Context,
		idempotencyKey uuid.UUID,
		items []dto.ProductReservationItem,
	) (reservedProducts []entity.Product, err error)
	ReleaseProducts(
		ctx context.Context,
		items []dto.ProductRestorationItem,
	) error
	ConfirmProductsDeduction(
		ctx context.Context,
		items []dto.ProductRestorationItem,
	) (products []entity.Product, err error)
	RestoreProducts(
		ctx context.Context,
		items []dto.ProductRestorationItem,
	) ([]entity.Product, error)
	HealthCheck(ctx context.Context) error
}

// productClient is a Connect-RPC client for interacting with the product service.
type productClient struct {
	client productv1connect.ProductServiceClient
}

// NewProductClient creates a new productClient instance with Connect-RPC.
func NewProductClient(
	cfg *config.Config,
) (ProductClient, error) {
	// Create HTTP client for Connect-RPC
	httpClient := &http.Client{
		Timeout: constant.ProductClientTimeout,
	}

	// Use static configuration for now
	baseURL := "http://" + net.JoinHostPort(
		cfg.Client.ProductGRPCHost,
		strconv.Itoa(cfg.Client.ProductGRPCPort),
	)

	// Create Connect-RPC client
	client := productv1connect.NewProductServiceClient(httpClient, baseURL)

	return &productClient{
		client: client,
	}, nil
}

// GetProducts fetches product data by IDs.
func (pc *productClient) GetProducts(
	ctx context.Context,
	ids []uuid.UUID,
) ([]entity.Product, error) {
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	ctx, cancel := context.WithTimeout(ctx, constant.ProductClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.GetProductsRequest{Ids: stringIDs})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := pc.client.GetProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call GetProducts: %w", err)
	}

	products := make([]entity.Product, len(resp.Msg.GetProducts()))

	for i, p := range resp.Msg.GetProducts() {
		product, rowErr := mapper.MapProtoToProduct(p)
		if rowErr != nil {
			return nil, rowErr
		}

		products[i] = product
	}

	return products, nil
}

// ReserveProducts reserves stock for products with optimistic locking.
func (pc *productClient) ReserveProducts(
	ctx context.Context,
	idempotencyKey uuid.UUID,
	items []dto.ProductReservationItem,
) ([]entity.Product, error) {
	// Convert to protobuf format
	pbItems := make([]*pb.ProductQuantityWithVersion, len(items))
	for i, item := range items {
		pbItems[i] = &pb.ProductQuantityWithVersion{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
			Version:   item.ExpectedVersion,
		}
	}

	ctx, cancel := context.WithTimeout(ctx, constant.ProductClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.ReserveProductsRequest{
		IdempotencyKey: idempotencyKey.String(),
		Items:          pbItems,
	})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := pc.client.ReserveProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ReserveProducts: %w", err)
	}

	if !resp.Msg.GetSuccess() {
		return nil, fmt.Errorf(
			"reservation failed from product-service: %s",
			resp.Msg.GetErrorMessage(),
		)
	}

	// Convert response to entities
	products := make([]entity.Product, len(resp.Msg.GetReservedProducts()))

	for i, p := range resp.Msg.GetReservedProducts() {
		product, rowErr := mapper.MapProtoToProduct(p)
		if rowErr != nil {
			return nil, rowErr
		}

		products[i] = product
	}

	return products, nil
}

// ReleaseProducts releases reserved stock for products.
func (pc *productClient) ReleaseProducts(
	ctx context.Context,
	items []dto.ProductRestorationItem,
) error {
	// Convert to protobuf format
	pbItems := make([]*pb.ProductQuantity, len(items))
	for i, item := range items {
		pbItems[i] = &pb.ProductQuantity{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
		}
	}

	ctx, cancel := context.WithTimeout(ctx, constant.ProductClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.ReleaseProductsRequest{
		Items: pbItems,
	})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := pc.client.ReleaseProducts(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to call ReleaseProducts: %w", err)
	}

	if !resp.Msg.GetSuccess() {
		return fmt.Errorf("products release failed: %s", resp.Msg.GetErrorMessage())
	}

	return nil
}

// ConfirmProductsDeduction confirms the reserved stock and removes reserved quantity.
func (pc *productClient) ConfirmProductsDeduction(
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

	ctx, cancel := context.WithTimeout(ctx, constant.ProductClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.ConfirmProductsDeductionRequest{
		Items: pbItems,
	})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := pc.client.ConfirmProductsDeduction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ConfirmProductsDeduction: %w", err)
	}

	if !resp.Msg.GetSuccess() {
		return nil, fmt.Errorf(
			"stocks deduction confirmation failed: %s",
			resp.Msg.GetErrorMessage(),
		)
	}

	products := make([]entity.Product, len(resp.Msg.GetUpdatedProducts()))

	for i, p := range resp.Msg.GetUpdatedProducts() {
		product, rowErr := mapper.MapProtoToProduct(p)
		if rowErr != nil {
			return nil, rowErr
		}

		products[i] = product
	}

	return products, nil
}

// RestoreProducts restores stock in case of compensation.
func (pc *productClient) RestoreProducts(
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

	ctx, cancel := context.WithTimeout(ctx, constant.ProductClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.RestoreProductsRequest{
		Items:  pbItems,
		Reason: "order_compensation", // Standard reason for saga compensation
	})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := pc.client.RestoreProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call RestoreProducts: %w", err)
	}

	if !resp.Msg.GetSuccess() {
		return nil, fmt.Errorf("products restoration failed: %s", resp.Msg.GetErrorMessage())
	}

	// Convert response to entities
	products := make([]entity.Product, len(resp.Msg.GetRestoredProducts()))

	for i, p := range resp.Msg.GetRestoredProducts() {
		product, rowErr := mapper.MapProtoToProduct(p)
		if rowErr != nil {
			return nil, rowErr
		}

		products[i] = product
	}

	return products, nil
}

// HealthCheck verifies the connection to product-service.
func (pc *productClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, constant.ProductClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.HealthRequest{})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := pc.client.Health(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Msg.GetStatus() != pb.HealthStatus_HEALTH_STATUS_SERVING {
		return fmt.Errorf("service unhealthy: %s", resp.Msg.GetStatus())
	}

	return nil
}
