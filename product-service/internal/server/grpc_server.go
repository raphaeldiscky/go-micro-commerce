package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	grpcAuth "github.com/raphaeldiscky/go-micro-template/pkg/grpc"
	pb "github.com/raphaeldiscky/go-micro-template/proto/product"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/service"
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
	for _, p := range products {
		resp.Products = append(resp.Products, &pb.Product{
			Id:       p.ID.String(),
			Name:     p.Name,
			Price:    p.Price.InexactFloat64(),
			Quantity: int32(p.Quantity),
		})
	}

	return resp, nil
}

// Health returns the health status of the product service.
func (s *GRPCServer) Health(_ context.Context, _ *emptypb.Empty) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: constant.GRPCHealthServing}, nil
}

// StartGRPC runs the gRPC server.
func (s *GRPCServer) StartGRPC() error {
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
func (s *GRPCServer) Shutdown() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(s.cfg.GRPCServer.GracePeriod)*time.Second,
	)
	defer cancel()

	s.logger.Info("Attempting to shut down the gRPC server...")

	if s.grpcServer != nil {
		stopped := make(chan struct{})
		go func() {
			s.grpcServer.GracefulStop()
			close(stopped)
		}()

		select {
		case <-ctx.Done():
			s.logger.Warn("Graceful shutdown timed out, forcing stop...")
			s.grpcServer.Stop()
		case <-stopped:
			s.logger.Info("gRPC server shut down gracefully")
		}
	}
}
