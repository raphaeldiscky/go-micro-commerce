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
		ShippingCost:   session.ShippingCost,
		TotalAmount:    session.TotalAmount,
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

// MapCourierToPayload maps entity.Courier to kafkaevent.Courier.
func MapCourierToPayload(courier entity.Courier) kafkaevent.Courier {
	return kafkaevent.Courier{
		CourierID: courier.CourierID,
	}
}

// MapDestinationToPayload maps entity.Destination to kafkaevent.Destination.
func MapDestinationToPayload(destination entity.Destination) kafkaevent.Destination {
	return kafkaevent.Destination{
		City:        destination.City,
		State:       destination.State,
		PostalCode:  destination.PostalCode,
		CountryCode: destination.CountryCode,
	}
}

// MapOriginToPayload maps entity.Origin to kafkaevent.Origin.
func MapOriginToPayload(origin entity.Origin) kafkaevent.Origin {
	return kafkaevent.Origin{
		City:        origin.City,
		State:       origin.State,
		PostalCode:  origin.PostalCode,
		CountryCode: origin.CountryCode,
	}
}

// MapPackageToPayload maps entity.Package to kafkaevent.Package.
func MapPackageToPayload(packageData entity.Package) kafkaevent.Package {
	return kafkaevent.Package{
		WeightKG: packageData.WeightKG,
		Width:    packageData.Width,
		Height:   packageData.Height,
		Length:   packageData.Length,
		Unit:     packageData.Unit,
	}
}
