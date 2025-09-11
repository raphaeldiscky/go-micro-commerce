package server

import (
	"context"
	"fmt"
	"net"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	grpcauth "github.com/raphaeldiscky/go-micro-commerce/pkg/grpc"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/fulfillment"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/service"
)

// GRPCServer is the gRPC server for fulfillment service.
type GRPCServer struct {
	pb.UnimplementedFulfillmentServiceServer
	cfg                *config.Config
	fulfillmentService service.FulfillmentServiceInterface
	logger             logger.Logger
	grpcServer         *grpc.Server
}

// NewGRPCServer creates a new gRPC server for fulfillment service.
func NewGRPCServer(
	fulfillmentService service.FulfillmentServiceInterface,
	appLogger logger.Logger,
	cfg *config.Config,
) *GRPCServer {
	return &GRPCServer{cfg: cfg, fulfillmentService: fulfillmentService, logger: appLogger}
}

// GetShippingCost gets the shipping cost for the order.
func (s *GRPCServer) GetShippingCost(
	ctx context.Context,
	req *pb.GetShippingCostRequest,
) (*pb.GetShippingCostResponse, error) {
	calculateShippingReq := mapper.MapToCalculateShippingRateRequest(req)

	res, err := s.fulfillmentService.CalculateShippingRate(ctx, calculateShippingReq)
	if err != nil {
		return nil, err
	}

	return &pb.GetShippingCostResponse{ShippingCost: res.ShippingCost.InexactFloat64()}, nil
}

// Health returns the health status of the product service.
func (s *GRPCServer) Health(_ context.Context, _ *emptypb.Empty) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: pkgconstant.GRPCHealthServing}, nil
}

// Start runs the gRPC server.
func (s *GRPCServer) Start(ctx context.Context) error {
	address := fmt.Sprintf("%s:%d", s.cfg.GRPCServer.Host, s.cfg.GRPCServer.Port)
	lc := &net.ListenConfig{}

	lis, err := lc.Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	// Create authentication interceptor
	authInterceptor := grpcauth.NewAuthInterceptor()

	// Create gRPC server with authentication interceptor
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.ServiceToServiceAuth()),
	)
	pb.RegisterFulfillmentServiceServer(s.grpcServer, s)

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
