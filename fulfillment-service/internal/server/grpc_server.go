package server

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/proto/fulfillment/v1/fulfillmentv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	connectauth "github.com/raphaeldiscky/go-micro-commerce/pkg/connect"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/fulfillment/v1"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/service"
)

// GRPCServer is the Connect-RPC server for fulfillment service.
type GRPCServer struct {
	cfg                *config.Config
	fulfillmentService service.FulfillmentService
	logger             logger.Logger
	httpServer         *http.Server
}

// NewGRPCServer creates a new Connect-RPC server for fulfillment service.
func NewGRPCServer(
	fulfillmentService service.FulfillmentService,
	appLogger logger.Logger,
	cfg *config.Config,
) *GRPCServer {
	return &GRPCServer{cfg: cfg, fulfillmentService: fulfillmentService, logger: appLogger}
}

// GetShippingCost gets the shipping cost for the order via Connect-RPC.
func (s *GRPCServer) GetShippingCost(
	ctx context.Context,
	req *connect.Request[pb.GetShippingCostRequest],
) (*connect.Response[pb.GetShippingCostResponse], error) {
	calculateShippingReq := mapper.MapToCalculateShippingRateRequest(req.Msg)

	res, err := s.fulfillmentService.CalculateShippingRate(ctx, calculateShippingReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &pb.GetShippingCostResponse{
		Success:      true,
		ShippingCost: res.ShippingCost.InexactFloat64(),
		Currency:     calculateShippingReq.Currency,
	}

	return connect.NewResponse(resp), nil
}

// Health returns the health status of the fulfillment service.
func (s *GRPCServer) Health(
	_ context.Context,
	_ *connect.Request[pb.HealthRequest],
) (*connect.Response[pb.HealthResponse], error) {
	resp := &pb.HealthResponse{
		Status: pb.HealthStatus_HEALTH_STATUS_SERVING,
	}

	return connect.NewResponse(resp), nil
}

// Start runs the Connect-RPC server.
func (s *GRPCServer) Start(_ context.Context) error {
	address := fmt.Sprintf("%s:%d", s.cfg.GRPCServer.Host, s.cfg.GRPCServer.Port)

	// Create authentication interceptor
	authInterceptor := connectauth.NewAuthInterceptor()

	// Create Connect-RPC handler with auth interceptor
	path, handler := fulfillmentv1connect.NewFulfillmentServiceHandler(
		s,
		connect.WithInterceptors(authInterceptor.ServiceToServiceAuth()),
	)

	// Create HTTP mux and register the handler
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	// Add gRPC reflection support for Connect-RPC
	reflector := grpcreflect.NewStaticReflector(
		fulfillmentv1connect.FulfillmentServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Add gRPC health check
	checker := grpchealth.NewStaticChecker(
		fulfillmentv1connect.FulfillmentServiceName,
	)
	mux.Handle(grpchealth.NewHandler(checker))

	// Add simple health endpoint for Consul health checks
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{"status":"healthy"}`))
		if err != nil {
			s.logger.Errorf("Failed to write health check response: %v", err)
		}
	})

	// Create HTTP server with h2c support for gRPC compatibility
	s.httpServer = &http.Server{
		Addr:              address,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: s.cfg.GRPCServer.ReadHeaderTimeout,
	}

	s.logger.Infof("Connect-RPC server listening on %s", address)

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the Connect-RPC server.
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Attempting to shut down the Connect-RPC server...")

	if s.httpServer == nil {
		s.logger.Info("Connect-RPC server was not started, nothing to shut down")
		return nil
	}

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		s.logger.Warn("Connect-RPC server shutdown error: %v", err)
		return err
	}

	s.logger.Info("Connect-RPC server shut down gracefully")

	return nil
}
