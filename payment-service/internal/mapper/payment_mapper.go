// Package mapper provides functions for mapping entity.Payment to dto.PaymentResponse.
package mapper

import (
	"fmt"
	"strings"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
)

// MapStringToPaymentMethod converts a string to a PaymentMethod.
func MapStringToPaymentMethod(s string) (constant.PaymentMethod, error) {
	constants := []constant.PaymentMethod{
		constant.PaymentMethodCreditCard,
		constant.PaymentMethodDebitCard,
		constant.PaymentMethodPayPal,
		constant.PaymentMethodBankTransfer,
	}

	for _, c := range constants {
		if strings.EqualFold(s, string(c)) {
			return c, nil
		}
	}

	return "", fmt.Errorf("invalid payment method: %s", s)
}

// MapToPaymentResponse converts domain entity to DTO response.
func MapToPaymentResponse(payment *entity.Payment) *dto.PaymentResponse {
	return &dto.PaymentResponse{
		ID:                 payment.ID,
		OrderID:            payment.OrderID,
		Amount:             payment.Amount,
		Currency:           payment.Currency,
		Status:             payment.Status,
		PaymentMethod:      payment.PaymentMethod,
		PaymentGateway:     payment.PaymentGateway,
		GatewayReferenceID: payment.GatewayReferenceID,
		CreatedAt:          payment.CreatedAt,
		UpdatedAt:          payment.UpdatedAt,
		CompletedAt:        payment.CompletedAt,
		FailedAt:           payment.FailedAt,
	}
}
