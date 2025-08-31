package server

import (
	"context"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	grpcAuth "github.com/raphaeldiscky/go-micro-commerce/pkg/grpc"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/product"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/service"
)

// GRPCServer is the gRPC server for product service.
type GRPCServer struct {
	pb.UnimplementedProductServiceServer
	cfg            *config.Config
	productService service.ProductServiceInterface
	logger         logger.Logger
	grpcServer     *grpc.Server
}

// NewGRPCServer creates a new gRPC server for product service.
func NewGRPCServer(
	productService service.ProductServiceInterface,
	appLogger logger.Logger,
	cfg *config.Config,
) *GRPCServer {
	return &GRPCServer{cfg: cfg, productService: productService, logger: appLogger}
}

// GetProducts retrieves products by IDs via gRPC.
func (s *GRPCServer) GetProducts(
	ctx context.Context,
	req *pb.GetProductsRequest,
) (*pb.GetProductsResponse, error) {
	ids := make([]uuid.UUID, len(req.Ids))

	for i, idStr := range req.Ids {
		uid, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %w", err)
		}

		ids[i] = uid
	}

	products, err := s.productService.GetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetProductsResponse{}

	for i := range products {
		p := &products[i]
		resp.Products = append(resp.Products, &pb.Product{
			Id:               p.ID.String(),
			Name:             p.Name,
			Price:            p.Price.InexactFloat64(),
			Quantity:         p.Quantity,
			Version:          p.Version,
			ReservedQuantity: p.ReservedQuantity,
		})
	}

	return resp, nil
}

// ReserveProducts reserves stock for products with optimistic locking.
func (s *GRPCServer) ReserveProducts(
	ctx context.Context,
	req *pb.ReserveProductsRequest,
) (*pb.ReserveProductsResponse, error) {
	// Convert protobuf request to service DTO
	reserveReq := dto.ReserveProductsRequest{
		OrderID: req.OrderId,
		Items:   make([]dto.ProductReservationItem, len(req.Items)),
	}

	for i, item := range req.Items {
		productID, err := uuid.Parse(item.ProductId)
		if err != nil {
			return &pb.ReserveProductsResponse{
				Success:      false,
				ErrorMessage: fmt.Sprintf("invalid product ID: %s", item.ProductId),
			}, nil
		}

		reserveReq.Items[i] = dto.ProductReservationItem{
			ProductID:       productID,
			Quantity:        item.Quantity,
			ExpectedVersion: item.Version,
		}
	}

	// Call service method
	reservedProducts, err := s.productService.ReserveProducts(ctx, reserveReq)
	if err != nil {
		return &pb.ReserveProductsResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Convert service response to protobuf
	resp := &pb.ReserveProductsResponse{
		Success:          true,
		ReservedProducts: make([]*pb.Product, len(reservedProducts)),
	}

	for i := range reservedProducts {
		p := &reservedProducts[i]
		resp.ReservedProducts[i] = &pb.Product{
			Id:               p.ID.String(),
			Name:             p.Name,
			Price:            p.Price.InexactFloat64(),
			Quantity:         p.Quantity,
			Version:          p.Version,
			ReservedQuantity: p.ReservedQuantity,
		}
	}

	return resp, nil
}

// Health returns the health status of the product service.
func (s *GRPCServer) Health(_ context.Context, _ *emptypb.Empty) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: constant.GRPCHealthServing}, nil
}

// Start runs the gRPC server.
func (s *GRPCServer) Start() error {
	address := fmt.Sprintf("%s:%d", s.cfg.GRPCServer.Host, s.cfg.GRPCServer.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	// Create authentication interceptor
	authInterceptor := grpcAuth.NewAuthInterceptor()

	// Create gRPC server with authentication interceptor
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.ServiceToServiceAuth()),
	)
	pb.RegisterProductServiceServer(s.grpcServer, s)

	// Enable gRPC reflection for development
	reflection.Register(s.grpcServer)

	s.logger.Infof("gRPC server listening on %s", address)

	return s.grpcServer.Serve(lis)
}

// Shutdown gracefully shuts down the gRPC server.
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Attempting to shut down the gRPC server...")

	if s.grpcServer == nil {
		s.logger.Info("gRPC server was not started, nothing to shut down")

		return nil
	}

	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.logger.Warn("Graceful shutdown timed out, forcing stop...")
		s.grpcServer.Stop()

		return ctx.Err()
	case <-stopped:
		s.logger.Info("gRPC server shut down gracefully")

		return nil
	}
}
