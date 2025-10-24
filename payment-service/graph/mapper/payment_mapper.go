// Package mapper provides mapping functions between service DTOs and GraphQL models.
package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

// MapToGraphQLPayment converts a service PaymentResponse DTO to a GraphQL Payment model.
func MapToGraphQLPayment(payment *dto.PaymentResponse) *graph.Payment {
	if payment == nil {
		return nil
	}

	return &graph.Payment{
		ID:             payment.ID,
		OrderID:        payment.OrderID,
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		Status:         payment.Status,
		PaymentMethod:  payment.PaymentMethod,
		PaymentGateway: payment.PaymentGateway,
		ClientSecret:   payment.ClientSecret, // Stripe client secret for Payment Element
		ExpiresAt:      payment.ExpiresAt,    // 24-hour payment window expiry
		CreatedAt:      payment.CreatedAt,
		UpdatedAt:      payment.UpdatedAt,
		CompletedAt:    payment.CompletedAt,
		FailedAt:       payment.FailedAt,
	}
}
