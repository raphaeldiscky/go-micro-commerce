package mapper

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
)

// MapToGraphQLCheckoutSessionFromDTO maps dto.CheckoutSessionResponse to graph.CheckoutSession.
func MapToGraphQLCheckoutSessionFromDTO(
	session *dto.CheckoutSessionResponse,
) *graph.CheckoutSession {
	items := make([]*graph.CheckoutSessionItem, len(session.Items))
	for i := range session.Items {
		items[i] = MapToGraphQLCheckoutSessionItemFromDTO(&session.Items[i])
	}

	return &graph.CheckoutSession{
		ID:             session.ID,
		IdempotencyKey: session.IdempotencyKey,
		CustomerID:     session.CustomerID,
		AddressID:      session.AddressID,
		CarrierID:      session.CarrierID,
		Status:         session.Status,
		PaymentGateway: session.PaymentGateway,
		PaymentMethod:  session.PaymentMethod,
		Currency:       session.Currency,
		Items:          items,
		CreatedAt:      session.CreatedAt,
		UpdatedAt:      session.UpdatedAt,
	}
}

// MapToGraphQLCheckoutSessionItemFromDTO maps dto.CheckoutSessionItemResponse to graph.CheckoutSessionItem.
func MapToGraphQLCheckoutSessionItemFromDTO(
	item *dto.CheckoutSessionItemResponse,
) *graph.CheckoutSessionItem {
	return &graph.CheckoutSessionItem{
		ID:        item.ID,
		ProductID: item.ProductID,
		Quantity:  int(item.Quantity),
		UnitPrice: item.UnitPrice,
	}
}

// MapToCreateCheckoutSessionRequest maps graph.CreateCheckoutSessionInput to dto.CreateCheckoutSessionRequest.
func MapToCreateCheckoutSessionRequest(
	input graph.CreateCheckoutSessionInput,
	customerID uuid.UUID,
) *dto.CreateCheckoutSessionRequest {
	return &dto.CreateCheckoutSessionRequest{
		CustomerID:     customerID,
		IdempotencyKey: input.IdempotencyKey,
		CartID:         input.CartID,
	}
}

// MapToPlaceOrderRequest maps graph.PlaceOrderInput to dto.PlaceOrderRequest.
func MapToPlaceOrderRequest(
	input graph.PlaceOrderInput,
	customerID uuid.UUID,
) *dto.PlaceOrderRequest {
	return &dto.PlaceOrderRequest{
		CustomerID:     customerID,
		IdempotencyKey: input.IdempotencyKey,
	}
}
