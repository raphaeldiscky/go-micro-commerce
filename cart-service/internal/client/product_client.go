// Package client provides a client for interacting with the product service.
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

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mapper"
)

// ProductClient defines methods available for grpc products.
type ProductClient interface {
	ValidateProducts(ctx context.Context, items []entity.CheckoutSessionItem) error
	GetProducts(ctx context.Context, ids []uuid.UUID) ([]entity.Product, error)
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

	req := connect.NewRequest(&pb.BatchGetProductsByIDsRequest{Ids: stringIDs})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := pc.client.BatchGetProductsByIDs(ctx, req)
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

// ValidateProducts validates products before place order using checkout session items snapshot.
func (pc *productClient) ValidateProducts(
	ctx context.Context,
	items []entity.CheckoutSessionItem,
) error {
	if len(items) == 0 {
		return nil
	}

	// Build validation request from checkout session items
	protoProducts := make([]*pb.ProductForValidation, len(items))
	for i, item := range items {
		protoProducts[i] = &pb.ProductForValidation{
			Id:       item.ProductID.String(),
			Price:    item.UnitPrice.InexactFloat64(),
			Quantity: item.Quantity,
		}
	}

	ctx, cancel := context.WithTimeout(ctx, constant.ProductClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.ValidateProductsRequest{
		Products: protoProducts,
	})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := pc.client.ValidateProducts(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to validate products: %w", err)
	}

	if !resp.Msg.GetSuccess() {
		return fmt.Errorf("product validation failed: %s", resp.Msg.GetMessage())
	}

	return nil
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
