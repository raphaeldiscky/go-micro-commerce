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
	"github.com/raphaeldiscky/go-micro-commerce/proto/cart/v1/cartv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/types/known/timestamppb"

	connectauth "github.com/raphaeldiscky/go-micro-commerce/pkg/connect"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/cart/v1"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
)

// GRPCServer is the Connect-RPC server for cart service.
type GRPCServer struct {
	cfg             *config.Config
	checkoutService service.CheckoutSessionService
	logger          logger.Logger
	httpServer      *http.Server
}

// NewGRPCServer creates a new Connect-RPC server for cart service.
func NewGRPCServer(
	checkoutService service.CheckoutSessionService,
	appLogger logger.Logger,
	cfg *config.Config,
) *GRPCServer {
	return &GRPCServer{cfg: cfg, checkoutService: checkoutService, logger: appLogger}
}

// GetCheckoutSession retrieves a checkout session by ID via gRPC.
func (s *GRPCServer) GetCheckoutSession(
	ctx context.Context,
	req *connect.Request[pb.GetCheckoutSessionRequest],
) (*connect.Response[pb.GetCheckoutSessionResponse], error) {
	// Parse checkout session ID from request
	sessionID, err := uuid.Parse(req.Msg.GetCheckoutSessionId())
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("invalid checkout session ID: %w", err),
		)
	}

	// Call the existing service method
	session, err := s.checkoutService.GetCheckoutSession(ctx, sessionID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Map items from DTO to proto
	items := make([]*pb.CheckoutSessionItem, len(session.Items))
	for i := range session.Items {
		item := &session.Items[i]
		items[i] = &pb.CheckoutSessionItem{
			Id:          item.ID.String(),
			ProductId:   item.ProductID.String(),
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice.String(),
		}
	}

	// Map DTO response to protobuf response
	resp := &pb.GetCheckoutSessionResponse{
		CheckoutSessionId: session.ID.String(),
		IdempotencyKey:    session.IdempotencyKey.String(),
		CustomerId:        session.CustomerID.String(),
		CartId:            session.CartID.String(),
		Status:            mapper.MapCheckoutSessionStatusToProto(session.Status),
		Destination: &pb.Destination{
			City:        session.Destination.City,
			State:       session.Destination.State,
			PostalCode:  session.Destination.PostalCode,
			CountryCode: session.Destination.CountryCode,
		},
		Origin: &pb.Origin{
			City:        session.Origin.City,
			State:       session.Origin.State,
			PostalCode:  session.Origin.PostalCode,
			CountryCode: session.Origin.CountryCode,
		},
		Courier: &pb.Courier{
			CourierId: session.Courier.CourierID,
		},
		Package: &pb.Package{
			WeightKg: session.Package.WeightKG.String(),
			Length:   session.Package.Length.String(),
			Width:    session.Package.Width.String(),
			Height:   session.Package.Height.String(),
			Unit:     session.Package.Unit,
		},
		Currency:  session.Currency,
		CreatedAt: timestamppb.New(session.CreatedAt),
		UpdatedAt: timestamppb.New(session.UpdatedAt),
		ExpiresAt: timestamppb.New(
			session.CreatedAt.Add(constant.CheckoutSessionExpirationTime),
		),
		Items:          items,
		PaymentGateway: session.PaymentGateway,
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
	path, handler := cartv1connect.NewCartServiceHandler(
		s,
		connect.WithInterceptors(authInterceptor.ServiceToServiceAuth()),
	)

	// Create HTTP mux and register the handler
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	// Add gRPC reflection support for Connect-RPC
	reflector := grpcreflect.NewStaticReflector(
		cartv1connect.CartServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Add gRPC health check
	checker := grpchealth.NewStaticChecker(
		cartv1connect.CartServiceName,
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
