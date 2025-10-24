// Package mapper provides functions for mapping entity.Payment to dto.PaymentResponse.
package mapper

import (
	"fmt"
	"strings"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
)

// MapStringToPaymentMethod converts a string to a PaymentMethod.
func MapStringToPaymentMethod(s string) (constant.PaymentMethod, error) {
	constants := []constant.PaymentMethod{
		constant.PaymentMethodCard,
	}

	for _, c := range constants {
		if strings.EqualFold(s, string(c)) {
			return c, nil
		}
	}

	return "", fmt.Errorf("invalid payment method: %s", s)
}

// MapStringToPaymentGateway converts a string to a PaymentGateway.
func MapStringToPaymentGateway(s string) (constant.PaymentGateway, error) {
	gateways := []constant.PaymentGateway{
		constant.PaymentGatewayStripe,
		constant.PaymentGatewayMock,
	}

	for _, g := range gateways {
		if strings.EqualFold(s, string(g)) {
			return g, nil
		}
	}

	return "", fmt.Errorf("invalid payment gateway: %s", s)
}

// MapToPaymentResponse converts domain entity to DTO response.
// Extracts gateway-specific fields from metadata for backward compatibility.
func MapToPaymentResponse(payment *entity.Payment) *dto.PaymentResponse {
	response := &dto.PaymentResponse{
		ID:                 payment.ID,
		OrderID:            payment.OrderID,
		Amount:             payment.Amount,
		Currency:           payment.Currency,
		Status:             payment.Status,
		PaymentGateway:     payment.PaymentGateway,
		GatewayReferenceID: payment.GatewayTransactionID,
		ExpiresAt:          payment.ExpiresAt,
		CreatedAt:          payment.CreatedAt,
		UpdatedAt:          payment.UpdatedAt,
		CompletedAt:        payment.CompletedAt,
		FailedAt:           payment.FailedAt,
	}

	// Extract gateway-specific metadata based on payment gateway type
	switch payment.PaymentGateway {
	case constant.PaymentGatewayStripe:
		// Extract Stripe metadata
		stripeMetadata, err := payment.GetStripeMetadata()
		if err == nil {
			response.PaymentMethodID = stripeMetadata.PaymentMethodID
			response.StripeCustomerID = stripeMetadata.CustomerID
			response.ClientSecret = stripeMetadata.ClientSecret
		}
	case constant.PaymentGatewayMock:
	// Add other gateways here as needed
	// case constant.PaymentGatewayMidtrans:
	// 	midtransMetadata, err := payment.GetMidtransMetadata()
	//  ...
	default:
		// For unknown gateways, try to extract common fields from raw metadata
		if payment.GatewayMetadata != nil {
			if clientSecret, ok := payment.GatewayMetadata["client_secret"].(string); ok {
				response.ClientSecret = &clientSecret
			}
		}
	}

	return response
}

// MapStatusToEventType maps payment status to Kafka event type.
func MapStatusToEventType(status constant.PaymentStatus) string {
	switch status {
	case constant.PaymentStatusPending:
		return kafka.PaymentCreatedEventType
	case constant.PaymentStatusProcessing:
		return kafka.PaymentProcessingEventType
	case constant.PaymentStatusCompleted:
		return kafka.PaymentCompletedEventType
	case constant.PaymentStatusFailed:
		return kafka.PaymentFailedEventType
	case constant.PaymentStatusRefunded:
		return kafka.PaymentRefundedEventType
	case constant.PaymentStatusTimeout:
		return kafka.PaymentTimeoutEventType
	default:
		return "unknown"
	}
}
