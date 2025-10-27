// Package server provides the Connect-RPC server for the payment service.
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
	"github.com/raphaeldiscky/go-micro-commerce/proto/payment/v1/paymentv1connect"
	"github.com/shopspring/decimal"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	connectauth "github.com/raphaeldiscky/go-micro-commerce/pkg/connect"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/payment/v1"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// GRPCServer is the Connect-RPC server for payment service.
type GRPCServer struct {
	cfg            *config.Config
	paymentService service.PaymentService
	logger         logger.Logger
	httpServer     *http.Server
}

// NewGRPCServer creates a new Connect-RPC server for payment service.
func NewGRPCServer(
	paymentService service.PaymentService,
	appLogger logger.Logger,
	cfg *config.Config,
) *GRPCServer {
	return &GRPCServer{
		cfg:            cfg,
		paymentService: paymentService,
		logger:         appLogger,
	}
}

// CreatePaymentIntent creates a PaymentIntent via Connect-RPC.
//
//nolint:gocyclo,cyclop,funlen // ignore complexity
func (s *GRPCServer) CreatePaymentIntent(
	ctx context.Context,
	req *connect.Request[pb.CreatePaymentIntentRequest],
) (*connect.Response[pb.CreatePaymentIntentResponse], error) {
	s.logger.Infof(
		"Creating PaymentIntent via Connect-RPC - OrderID: %s, Amount: %s %s, Customer: %s",
		req.Msg.GetOrderId(),
		req.Msg.GetAmount(),
		req.Msg.GetCurrency(),
		req.Msg.GetCustomerId(),
	)

	// Parse UUIDs from request
	orderID, err := uuid.Parse(req.Msg.GetOrderId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	customerID, err := uuid.Parse(req.Msg.GetCustomerId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	checkoutSessionID, err := uuid.Parse(req.Msg.GetCheckoutSessionId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	idempotencyKey, err := uuid.Parse(req.Msg.GetIdempotencyKey())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Parse amount
	amount, err := decimal.NewFromString(req.Msg.GetAmount())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Parse payment gateway
	paymentGateway := constant.PaymentGateway(req.Msg.GetPaymentGateway())
	// Validate payment gateway
	validGateways := []constant.PaymentGateway{
		constant.PaymentGatewayStripe,
		constant.PaymentGatewayMock,
	}
	isValid := false

	for _, gateway := range validGateways {
		if paymentGateway == gateway {
			isValid = true
			break
		}
	}

	if !isValid {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("invalid payment gateway: %s", req.Msg.GetPaymentGateway()))
	}

	// Convert items
	items := make([]dto.PaymentItemDTO, len(req.Msg.GetItems()))
	for i, item := range req.Msg.GetItems() {
		productID, errParse := uuid.Parse(item.GetProductId())
		if errParse != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errParse)
		}

		unitPrice, errn := decimal.NewFromString(item.GetUnitPrice())
		if errn != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errn)
		}

		items[i] = dto.PaymentItemDTO{
			ProductID:   productID,
			ProductName: item.GetProductName(),
			Quantity:    item.GetQuantity(),
			UnitPrice:   unitPrice,
			Currency:    item.GetCurrency(),
		}
	}

	// Build service request
	serviceReq := dto.CreatePaymentIntentRequest{
		OrderID:           orderID,
		CustomerID:        customerID,
		CustomerEmail:     req.Msg.GetCustomerEmail(),
		Amount:            amount,
		Currency:          req.Msg.GetCurrency(),
		PaymentGateway:    paymentGateway,
		IdempotencyKey:    idempotencyKey,
		Items:             items,
		CheckoutSessionID: checkoutSessionID,
	}

	// Call payment service
	serviceResp, err := s.paymentService.CreatePaymentIntent(ctx, serviceReq)
	if err != nil {
		s.logger.Errorf("Failed to create PaymentIntent: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert gateway metadata to protobuf struct
	var gatewayMetadata *structpb.Struct

	if serviceResp.GatewayMetadata != nil {
		// Convert map[string]string to map[string]any
		metadata := make(map[string]any)
		for k, v := range serviceResp.GatewayMetadata {
			metadata[k] = v
		}

		gatewayMetadata, err = structpb.NewStruct(metadata)
		if err != nil {
			s.logger.Errorf("Failed to convert gateway metadata: %v", err)
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// Build response
	resp := &pb.CreatePaymentIntentResponse{
		PaymentIntentId: serviceResp.PaymentIntentID,
		ClientSecret:    serviceResp.ClientSecret,
		PaymentGateway:  serviceResp.PaymentGateway,
		Status:          serviceResp.Status,
		Amount:          serviceResp.Amount,
		Currency:        serviceResp.Currency,
		OrderId:         serviceResp.OrderID,
		GatewayMetadata: gatewayMetadata,
	}

	if serviceResp.ExpiresAt != nil {
		resp.ExpiresAt = timestamppb.New(*serviceResp.ExpiresAt)
	}

	return connect.NewResponse(resp), nil
}

// Start runs the Connect-RPC server.
func (s *GRPCServer) Start(_ context.Context) error {
	address := fmt.Sprintf("%s:%d", s.cfg.HTTPServer.Host, s.cfg.HTTPServer.GRPCPort)

	// Create authentication interceptor
	authInterceptor := connectauth.NewAuthInterceptor()

	// Create Connect-RPC handler with auth interceptor
	path, handler := paymentv1connect.NewPaymentServiceHandler(
		s,
		connect.WithInterceptors(authInterceptor.ServiceToServiceAuth()),
	)

	// Create HTTP mux and register the handler
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	// Add gRPC reflection support for Connect-RPC
	reflector := grpcreflect.NewStaticReflector(
		paymentv1connect.PaymentServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Add gRPC health check
	checker := grpchealth.NewStaticChecker(
		paymentv1connect.PaymentServiceName,
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
		ReadHeaderTimeout: s.cfg.HTTPServer.ReadHeaderTimeout,
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
