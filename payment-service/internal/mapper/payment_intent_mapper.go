// Package mapper provides functions for mapping payment intent DTOs to proto messages.
package mapper

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/raphaeldiscky/go-micro-commerce/proto/payment/v1"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

// MapCreatePaymentIntentRequestFromProto converts proto request to DTO request.
func MapCreatePaymentIntentRequestFromProto(
	req *pb.CreatePaymentIntentRequest,
) (*dto.CreatePaymentIntentRequestDTO, error) {
	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, err
	}

	customerID, err := uuid.Parse(req.GetCustomerId())
	if err != nil {
		return nil, err
	}

	// Map payment gateway enum
	var paymentGateway constant.PaymentGateway

	switch req.GetPaymentGateway() {
	case pb.PaymentGateway_PAYMENT_GATEWAY_STRIPE:
		paymentGateway = constant.PaymentGatewayStripe
	case pb.PaymentGateway_PAYMENT_GATEWAY_XENDIT:
		paymentGateway = constant.PaymentGatewayXendit
	case pb.PaymentGateway_PAYMENT_GATEWAY_UNSPECIFIED:
		paymentGateway = constant.PaymentGatewayStripe
	default:
		paymentGateway = constant.PaymentGatewayStripe // Default to Stripe
	}

	return &dto.CreatePaymentIntentRequestDTO{
		OrderID:        orderID,
		Amount:         decimal.NewFromFloat(req.GetAmount()),
		Currency:       req.GetCurrency(),
		PaymentGateway: paymentGateway,
		CustomerEmail:  req.GetCustomerEmail(),
		CustomerID:     customerID,
	}, nil
}

// MapCreatePaymentIntentResponseToProto converts DTO response to proto response.
func MapCreatePaymentIntentResponseToProto(
	res *dto.CreatePaymentIntentResponseDTO,
) (*pb.CreatePaymentIntentResponse, error) {
	// Convert gateway_metadata map to protobuf Struct
	gatewayMetadataStruct, err := structpb.NewStruct(res.GatewayMetadata)
	if err != nil {
		return nil, err
	}

	// Map payment gateway enum
	var paymentGateway pb.PaymentGateway
	switch res.PaymentGateway {
	case constant.PaymentGatewayStripe:
		paymentGateway = pb.PaymentGateway_PAYMENT_GATEWAY_STRIPE
	case constant.PaymentGatewayXendit:
		paymentGateway = pb.PaymentGateway_PAYMENT_GATEWAY_XENDIT
	case constant.PaymentGatewayMock:
		paymentGateway = pb.PaymentGateway_PAYMENT_GATEWAY_UNSPECIFIED
	default:
		paymentGateway = pb.PaymentGateway_PAYMENT_GATEWAY_UNSPECIFIED
	}

	protoResponse := &pb.CreatePaymentIntentResponse{
		PaymentId:            res.PaymentID.String(),
		OrderId:              res.OrderID.String(),
		Amount:               res.Amount.InexactFloat64(),
		Currency:             res.Currency,
		PaymentGateway:       paymentGateway,
		GatewayTransactionId: res.GatewayTransactionID,
		GatewayMetadata:      gatewayMetadataStruct,
		CreatedAt:            timestamppb.New(res.CreatedAt),
		UpdatedAt:            timestamppb.New(res.UpdatedAt),
	}

	if res.ExpiresAt != nil {
		protoResponse.ExpiresAt = timestamppb.New(*res.ExpiresAt)
	}

	return protoResponse, nil
}
