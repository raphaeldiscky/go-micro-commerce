// Package mapper provides functions for mapping checkout session entities to DTOs and vice versa.
package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
)

// MapToCheckoutSessionResponse converts domain entity to DTO response.
func MapToCheckoutSessionResponse(session *entity.CheckoutSession) *dto.CheckoutSessionResponse {
	return &dto.CheckoutSessionResponse{
		ID:             session.ID,
		IdempotencyKey: session.IdempotencyKey,
		CustomerID:     session.CustomerID,
		Courier: dto.Courier{
			CourierID: session.Courier.CourierID,
		},
		Destination: dto.Destination{
			City:        session.Destination.City,
			State:       session.Destination.State,
			PostalCode:  session.Destination.PostalCode,
			CountryCode: session.Destination.CountryCode,
		},
		Origin: dto.Origin{
			City:        session.Origin.City,
			State:       session.Origin.State,
			PostalCode:  session.Origin.PostalCode,
			CountryCode: session.Origin.CountryCode,
		},
		Package: dto.Package{
			WeightKG: session.Package.WeightKG,
			Width:    session.Package.Width,
			Height:   session.Package.Height,
			Length:   session.Package.Length,
			Unit:     session.Package.Unit,
		},
		Status:         session.Status,
		PaymentGateway: session.PaymentGateway,
		Currency:       session.Currency,
		Items:          MapToCheckoutSessionItemResponses(session.Items),
		CreatedAt:      session.CreatedAt,
		UpdatedAt:      session.UpdatedAt,
	}
}

// MapToCheckoutSessionItemResponses converts domain entities to DTO responses.
func MapToCheckoutSessionItemResponses(
	items []entity.CheckoutSessionItem,
) []dto.CheckoutSessionItemResponse {
	var responses []dto.CheckoutSessionItemResponse

	for i := range items {
		item := &items[i]
		responses = append(responses, dto.CheckoutSessionItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}

	return responses
}
