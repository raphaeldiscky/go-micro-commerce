package server

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/proto/product/v1/productv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	connectauth "github.com/raphaeldiscky/go-micro-commerce/pkg/connect"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/product/v1"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/service"
)

// GRPCServer is the Connect-RPC server for product service.
type GRPCServer struct {
	cfg            *config.Config
	productService service.ProductService
	logger         logger.Logger
	httpServer     *http.Server
}

// NewGRPCServer creates a new Connect-RPC server for product service.
func NewGRPCServer(
	productService service.ProductService,
	appLogger logger.Logger,
	cfg *config.Config,
) *GRPCServer {
	return &GRPCServer{cfg: cfg, productService: productService, logger: appLogger}
}

// BatchGetProductsByIDs retrieves products by IDs via Connect-RPC.
func (s *GRPCServer) BatchGetProductsByIDs(
	ctx context.Context,
	req *connect.Request[pb.BatchGetProductsByIDsRequest],
) (*connect.Response[pb.BatchGetProductsByIDsResponse], error) {
	ids := make([]uuid.UUID, len(req.Msg.GetIds()))

	for i, idStr := range req.Msg.GetIds() {
		uid, err := uuid.Parse(idStr)
		if err != nil {
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				fmt.Errorf("invalid product ID: %w", err),
			)
		}

		ids[i] = uid
	}

	products, err := s.productService.BatchGetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &pb.BatchGetProductsByIDsResponse{
		Products: mapper.MapDTOToProtobufProducts(products),
	}

	return connect.NewResponse(resp), nil
}

// ReserveProducts reserves stock for products with optimistic locking.
func (s *GRPCServer) ReserveProducts(
	ctx context.Context,
	req *connect.Request[pb.ReserveProductsRequest],
) (*connect.Response[pb.ReserveProductsResponse], error) {
	// Convert protobuf request to service DTO
	reserveReq := dto.ReserveProductsRequest{
		IdempotencyKey: req.Msg.GetIdempotencyKey(),
		Items:          make([]dto.ProductReservationItem, len(req.Msg.GetItems())),
	}

	for i, item := range req.Msg.GetItems() {
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil {
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				fmt.Errorf("invalid product ID: %s", item.GetProductId()),
			)
		}

		reserveReq.Items[i] = dto.ProductReservationItem{
			ProductID:       productID,
			Quantity:        item.GetQuantity(),
			ExpectedVersion: item.GetVersion(),
		}
	}

	// Call service method
	reservedProducts, err := s.productService.ReserveProducts(ctx, reserveReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service response to protobuf
	resp := &pb.ReserveProductsResponse{
		Success:          true,
		ReservedProducts: mapper.MapDTOToProtobufProducts(reservedProducts),
	}

	return connect.NewResponse(resp), nil
}

// ConfirmProductsDeduction confirms stock deduction for reserved products via Connect-RPC.
func (s *GRPCServer) ConfirmProductsDeduction(
	ctx context.Context,
	req *connect.Request[pb.ConfirmProductsDeductionRequest],
) (*connect.Response[pb.ConfirmProductsDeductionResponse], error) {
	// Convert protobuf request to service DTO
	deductReq := dto.ConfirmProductsDeductionRequest{
		Items: make([]dto.ProductRestorationItem, len(req.Msg.GetItems())),
	}

	for i, item := range req.Msg.GetItems() {
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil {
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				fmt.Errorf("invalid product ID: %s", item.GetProductId()),
			)
		}

		deductReq.Items[i] = dto.ProductRestorationItem{
			ProductID: productID,
			Quantity:  item.GetQuantity(),
		}
	}

	// Call service method
	updatedProducts, err := s.productService.ConfirmProductsDeduction(ctx, deductReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service response to protobuf
	resp := &pb.ConfirmProductsDeductionResponse{
		Success:         true,
		UpdatedProducts: mapper.MapDTOToProtobufProducts(updatedProducts),
	}

	return connect.NewResponse(resp), nil
}

// ReleaseProducts releases reserved products via Connect-RPC.
func (s *GRPCServer) ReleaseProducts(
	ctx context.Context,
	req *connect.Request[pb.ReleaseProductsRequest],
) (*connect.Response[pb.ReleaseProductsResponse], error) {
	// Convert protobuf request to service DTO
	releaseReq := dto.ReleaseProductsRequest{
		Items: make([]dto.ProductRestorationItem, len(req.Msg.GetItems())),
	}

	for i, item := range req.Msg.GetItems() {
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil {
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				fmt.Errorf("invalid product ID: %s", item.GetProductId()),
			)
		}

		releaseReq.Items[i] = dto.ProductRestorationItem{
			ProductID: productID,
			Quantity:  item.GetQuantity(),
		}
	}

	// Call service method
	err := s.productService.ReleaseProducts(ctx, releaseReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.ReleaseProductsResponse{Success: true}), nil
}

// RestoreProducts restores products via Connect-RPC.
func (s *GRPCServer) RestoreProducts(
	ctx context.Context,
	req *connect.Request[pb.RestoreProductsRequest],
) (*connect.Response[pb.RestoreProductsResponse], error) {
	// Convert protobuf request to service DTO
	restoreReq := dto.RestoreProductsRequest{
		Items:  make([]dto.ProductRestorationItem, len(req.Msg.GetItems())),
		Reason: req.Msg.GetReason(),
	}

	for i, item := range req.Msg.GetItems() {
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil {
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				fmt.Errorf("invalid product ID: %s", item.GetProductId()),
			)
		}

		restoreReq.Items[i] = dto.ProductRestorationItem{
			ProductID: productID,
			Quantity:  item.GetQuantity(),
		}
	}

	// Call service method
	restoredProducts, err := s.productService.RestoreProducts(ctx, restoreReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service response to protobuf
	resp := &pb.RestoreProductsResponse{
		Success:          true,
		RestoredProducts: mapper.MapDTOToProtobufProducts(restoredProducts),
	}

	return connect.NewResponse(resp), nil
}

// Health returns the health status of the product service.
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
	path, handler := productv1connect.NewProductServiceHandler(
		s,
		connect.WithInterceptors(authInterceptor.ServiceToServiceAuth()),
	)

	// Create HTTP mux and register the handler
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	// Add gRPC reflection support for Connect-RPC
	reflector := grpcreflect.NewStaticReflector(
		productv1connect.ProductServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Add gRPC health check
	checker := grpchealth.NewStaticChecker(
		productv1connect.ProductServiceName,
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
