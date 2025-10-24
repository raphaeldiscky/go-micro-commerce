// Package mapper provides functions for mapping checkout session entities to DTOs and vice versa.
package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
)

// MapToCheckoutSessionResponse converts domain entity to DTO response.
func MapToCheckoutSessionResponse(session *entity.CheckoutSession) *dto.CheckoutSessionResponse {
	return &dto.CheckoutSessionResponse{
		ID:             session.ID,
		IdempotencyKey: session.IdempotencyKey,
		CustomerID:     session.CustomerID,
		AddressID:      session.AddressID,
		CarrierID:      session.CarrierID,
		Status:         session.Status,
		PaymentGateway: session.PaymentGateway,
		PaymentMethod:  session.PaymentMethod,
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

// MapCheckoutSessionStatusToEventType maps checkout session status to Kafka event type.
func MapCheckoutSessionStatusToEventType(status constant.CheckoutSessionStatus) string {
	switch status {
	case constant.CheckoutSessionStatusPending:
		return kafka.CheckoutSessionCreatedEventType
	case constant.CheckoutSessionStatusOrderPlaced:
		return kafka.CheckoutSessionOrderPlacedEventType
	case constant.CheckoutSessionStatusCanceled:
		return kafka.CheckoutSessionCanceledEventType
	default:
		return "unknown"
	}
}

// MapCheckoutSessionItemsToPayload maps checkout session items to their payload representation.
func MapCheckoutSessionItemsToPayload(
	items []entity.CheckoutSessionItem,
) []kafkaevent.CheckoutItemPayload {
	payloadItems := make([]kafkaevent.CheckoutItemPayload, len(items))

	for i := range items {
		item := &items[i]
		payloadItems[i] = kafkaevent.CheckoutItemPayload{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}

	return payloadItems
}
