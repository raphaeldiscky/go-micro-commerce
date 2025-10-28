package mapper

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
)

// MapToGraphQLOrderFromDTO maps dto.OrderResponse to graph.Order.
func MapToGraphQLOrderFromDTO(order *dto.OrderResponse) *graph.Order {
	items := make([]*graph.OrderItem, len(order.Items))
	for i := range order.Items {
		items[i] = MapToGraphQLOrderItemFromDTO(&order.Items[i])
	}

	return &graph.Order{
		ID:                order.ID,
		IdempotencyKey:    order.IdempotencyKey,
		CheckoutSessionID: order.CheckoutSessionID,
		CustomerID:        order.CustomerID,
		Status:            order.Status,
		Currency:          order.Currency,
		PaymentGateway:    order.PaymentGateway,
		Courier: &graph.Courier{
			CourierID: order.Courier.CourierID,
		},
		Package: &graph.Package{
			WeightKg: order.Package.WeightKG,
			Length:   order.Package.Length,
			Height:   order.Package.Height,
			Width:    order.Package.Width,
			Unit:     order.Package.Unit,
		},
		Origin: &graph.Origin{
			City:        order.Origin.City,
			State:       order.Origin.State,
			PostalCode:  order.Origin.PostalCode,
			CountryCode: order.Origin.CountryCode,
		},
		Destination: &graph.Destination{
			City:        order.Destination.City,
			State:       order.Destination.State,
			PostalCode:  order.Destination.PostalCode,
			CountryCode: order.Destination.CountryCode,
		},
		ShippingCost:  order.ShippingCost,
		Subtotal:      order.Subtotal,
		TotalPrice:    order.TotalPrice,
		TotalTax:      order.TotalTax,
		TotalDiscount: order.TotalDiscount,
		Items:         items,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}
}

// MapToGraphQLOrderItemFromDTO maps dto.OrderItemResponse to graph.OrderItem.
func MapToGraphQLOrderItemFromDTO(item *dto.OrderItemResponse) *graph.OrderItem {
	return &graph.OrderItem{
		ID:            item.ID,
		OrderID:       item.OrderID,
		ProductID:     item.ProductID,
		ProductName:   item.ProductName,
		Quantity:      int(item.Quantity),
		UnitPrice:     item.UnitPrice,
		TotalPrice:    item.TotalPrice,
		TaxRate:       item.TaxRate,
		TotalTax:      item.TotalTax,
		TotalDiscount: item.TotalDiscount,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

// MapToGraphQLOrderConnection maps order list to GraphQL connection.
func MapToGraphQLOrderConnection(
	orders []dto.OrderResponse,
	nextCursor string,
	hasNextPage bool,
) *graph.OrderConnection {
	edges := make([]*graph.OrderEdge, len(orders))

	for i := range orders {
		// Generate cursor for each edge
		orderID := orders[i].ID.String()
		timestamp := orders[i].CreatedAt.Unix()

		cursorData := fmt.Sprintf(`{"id":"%s","timestamp":%d}`, orderID, timestamp)

		edges[i] = &graph.OrderEdge{
			Node:   MapToGraphQLOrderFromDTO(&orders[i]),
			Cursor: cursorData,
		}
	}

	var endCursor *string

	if nextCursor != "" {
		endCursor = &nextCursor
	}

	return &graph.OrderConnection{
		Edges: edges,
		PageInfo: &graph.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: false,
			StartCursor:     nil,
			EndCursor:       endCursor,
		},
	}
}

// MapToCreateOrderRequest maps graph.CreateOrderInput to dto.CreateOrderRequest.
func MapToCreateOrderRequest(
	input graph.CreateOrderInput,
	customerID uuid.UUID,
	customerEmail string,
) (*dto.CreateOrderRequest, error) {
	items := make([]dto.CreateOrderItemRequest, len(input.Items))

	for i, item := range input.Items {
		items[i] = dto.CreateOrderItemRequest{
			ProductID: item.ProductID,
			Quantity:  int64(item.Quantity),
		}
	}

	return &dto.CreateOrderRequest{
		CustomerID:     customerID,
		CustomerEmail:  customerEmail,
		IdempotencyKey: input.IdempotencyKey,
		Items:          items,
		PaymentGateway: input.PaymentGateway,
		Currency:       input.Currency,
	}, nil
}

// MapToGraphQLPaymentMetadata maps dto.PaymentMetadata to graph.PaymentMetadata.
func MapToGraphQLPaymentMetadata(metadata dto.PaymentMetadata) (*graph.PaymentMetadata, error) {
	// Marshal gateway metadata map to JSON string
	gatewayMetadataBytes, err := json.Marshal(metadata.GatewayMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gateway metadata: %w", err)
	}

	return &graph.PaymentMetadata{
		PaymentID:            metadata.PaymentID,
		PaymentGateway:       metadata.PaymentGateway,
		GatewayTransactionID: metadata.GatewayTransactionID,
		GatewayMetadata:      string(gatewayMetadataBytes),
		Amount:               metadata.Amount,
		Currency:             metadata.Currency,
	}, nil
}

// MapToGraphQLPlaceOrderPayload maps dto.PlaceOrderResponse to graph.PlaceOrderPayload.
func MapToGraphQLPlaceOrderPayload(
	response *dto.PlaceOrderResponse,
) (*graph.PlaceOrderPayload, error) {
	paymentMetadata, err := MapToGraphQLPaymentMetadata(response.PaymentMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to map payment metadata: %w", err)
	}

	return &graph.PlaceOrderPayload{
		Order:           MapToGraphQLOrderFromDTO(response.Order),
		PaymentMetadata: paymentMetadata,
	}, nil
}
