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
		Courier: &graph.Courier{
			CourierID: session.Courier.CourierID,
		},
		Destination: &graph.Destination{
			City:        session.Destination.City,
			State:       session.Destination.State,
			PostalCode:  session.Destination.PostalCode,
			CountryCode: session.Destination.CountryCode,
		},
		Origin: &graph.Origin{
			City:        session.Origin.City,
			State:       session.Origin.State,
			PostalCode:  session.Origin.PostalCode,
			CountryCode: session.Origin.CountryCode,
		},
		Package: &graph.Package{
			WeightKg: session.Package.WeightKG,
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
		ID:          item.ID,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		Quantity:    int(item.Quantity),
		UnitPrice:   item.UnitPrice,
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

// MapToUpdateCheckoutSessionRequest maps graph.UpdateCheckoutSessionInput to dto.UpdateCheckoutSessionRequest.
func MapToUpdateCheckoutSessionRequest(
	input graph.UpdateCheckoutSessionInput,
	customerID uuid.UUID,
) *dto.UpdateCheckoutSessionRequest {
	req := &dto.UpdateCheckoutSessionRequest{
		CustomerID:     customerID,
		PaymentGateway: input.PaymentGateway,
	}

	if input.Courier != nil {
		req.Courier = &dto.Courier{
			CourierID: input.Courier.CourierID,
		}
	}

	if input.Destination != nil {
		req.Destination = &dto.Destination{
			City:        input.Destination.City,
			State:       input.Destination.State,
			PostalCode:  input.Destination.PostalCode,
			CountryCode: input.Destination.CountryCode,
		}
	}

	if input.Origin != nil {
		req.Origin = &dto.Origin{
			City:        input.Origin.City,
			State:       input.Origin.State,
			PostalCode:  input.Origin.PostalCode,
			CountryCode: input.Origin.CountryCode,
		}
	}

	if input.Package != nil {
		req.Package = &dto.Package{
			WeightKG: input.Package.WeightKg,
			Width:    input.Package.Width,
			Height:   input.Package.Height,
			Length:   input.Package.Length,
			Unit:     input.Package.Unit,
		}
	}

	return req
}

// MapToPlaceOrderRequest maps graph.PlaceOrderInput to dto.PlaceOrderRequest.
func MapToPlaceOrderRequest(
	input graph.PlaceOrderInput,
	customerID uuid.UUID,
	sessionID uuid.UUID,
) *dto.PlaceOrderRequest {
	return &dto.PlaceOrderRequest{
		CustomerID:        customerID,
		CheckoutSessionID: sessionID,
		IdempotencyKey:    input.IdempotencyKey,
	}
}

// MapToGraphQLPlaceOrderResponseFromDTO maps dto.PlaceOrderResponse to graph.PlaceOrderResponse.
func MapToGraphQLPlaceOrderResponseFromDTO(
	resp *dto.PlaceOrderResponse,
) *graph.PlaceOrderResponse {
	checkoutSession := MapToGraphQLCheckoutSessionFromDTO(&resp.CheckoutSession)

	var redirectURL *string
	if resp.RedirectURL != "" {
		redirectURL = &resp.RedirectURL
	}

	return &graph.PlaceOrderResponse{
		CheckoutSession: checkoutSession,
		TransactionID:   resp.TransactionID,
		Amount:          resp.Amount,
		Currency:        resp.Currency,
		Status:          resp.Status,
		RedirectURL:     redirectURL,
		GatewayMetadata:  resp.GatewayMetadata,
	}
}
