// Package client provides a gRPC client for interacting with the product service.
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/emptypb"

	grpcAuth "github.com/raphaeldiscky/go-micro-commerce/pkg/grpc"
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
	conn         *grpc.ClientConn
	client       pb.ProductServiceClient
	consulClient *api.Client
	serviceName  string
	clientCfg    *config.ClientConfig
	consulCfg    *config.ConsulConfig
}

// NewProductClient creates a new ProductClient instance with gRPC connection.
func NewProductClient(
	clientCfg *config.ClientConfig,
	consulCfg *config.ConsulConfig,
) (ProductClientInterface, error) {
	var conn *grpc.ClientConn

	var err error

	var consulClient *api.Client

	if shouldUseServiceDiscovery(clientCfg, consulCfg) {
		conn, consulClient, err = createConsulConnection(consulCfg)
		if err != nil {
			return nil, err
		}
	} else {
		conn, err = createStaticConnection(clientCfg)
	}

	if err != nil {
		return nil, err
	}

	client := pb.NewProductServiceClient(conn)

	return &ProductClient{
		conn:         conn,
		client:       client,
		consulClient: consulClient,
		serviceName:  "product-service-grpc",
		clientCfg:    clientCfg,
		consulCfg:    consulCfg,
	}, nil
}

// shouldUseServiceDiscovery checks if service discovery should be used.
func shouldUseServiceDiscovery(
	clientCfg *config.ClientConfig,
	consulCfg *config.ConsulConfig,
) bool {
	return clientCfg.UseServiceDiscovery && consulCfg.Enabled
}

// createConsulConnection creates a gRPC connection using Consul service discovery.
func createConsulConnection(consulCfg *config.ConsulConfig) (*grpc.ClientConn, *api.Client, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulCfg.Address

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	address, err := getServiceAddress(consulClient, "product-service-grpc")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get product-service address from consul: %w", err)
	}

	conn, err := createGRPCConnection(address)
	if err != nil {
		return nil, nil, err
	}

	return conn, consulClient, nil
}

// createStaticConnection creates a gRPC connection using static configuration.
func createStaticConnection(clientCfg *config.ClientConfig) (*grpc.ClientConn, error) {
	address := fmt.Sprintf("%s:%d", clientCfg.ProductGRPCHost, clientCfg.ProductGRPCPort)

	return createGRPCConnection(address)
}

// createGRPCConnection creates a gRPC connection with common options and resilience features.
func createGRPCConnection(address string) (*grpc.ClientConn, error) {
	clientAuth := grpcAuth.NewClientAuthInterceptor()

	// Configure keepalive parameters for automatic reconnection
	kacp := keepalive.ClientParameters{
		Time:                30 * time.Second, // Send keepalive pings every 30 seconds (less aggressive)
		Timeout:             5 * time.Second,  // Wait 5 seconds for ping ack before considering the connection dead
		PermitWithoutStream: false,            // Only send keepalive pings when there are active streams
	}

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(clientAuth.ForwardUserAuth()),
		grpc.WithKeepaliveParams(kacp),
		// Enable automatic reconnection with reasonable retry policy
		grpc.WithDefaultServiceConfig(`{
			"methodConfig": [{
				"name": [{"service": "product.ProductService"}],
				"retryPolicy": {
					"MaxAttempts": 3,
					"InitialBackoff": "0.1s",
					"MaxBackoff": "1s", 
					"BackoffMultiplier": 2.0,
					"RetryableStatusCodes": [ "UNAVAILABLE", "DEADLINE_EXCEEDED" ]
				}
			}],
			"loadBalancingPolicy": "round_robin"
		}`),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product-service: %w", err)
	}

	return conn, nil
}

// getServiceAddress retrieves the service address from Consul.
func getServiceAddress(consulClient *api.Client, serviceName string) (string, error) {
	services, _, err := consulClient.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("failed to query consul for service %s: %w", serviceName, err)
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances found for service %s", serviceName)
	}

	// Use the first healthy instance
	service := services[0].Service

	return fmt.Sprintf("%s:%d", service.Address, service.Port), nil
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
	return pc.conn.Close()
}
