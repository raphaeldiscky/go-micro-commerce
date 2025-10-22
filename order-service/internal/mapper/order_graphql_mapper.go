package mapper

import (
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
		ID:             order.ID,
		IdempotencyKey: order.IdempotencyKey,
		CustomerID:     order.CustomerID,
		Status:         order.Status,
		Currency:       order.Currency,
		PaymentGateway: order.PaymentGateway,
		PaymentMethod:  order.PaymentMethod,
		ShippingCost:   order.ShippingCost,
		Subtotal:       order.Subtotal,
		TotalPrice:     order.TotalPrice,
		TotalTax:       order.TotalTax,
		TotalDiscount:  order.TotalDiscount,
		Items:          items,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

// MapToGraphQLOrderItemFromDTO maps dto.OrderItemResponse to graph.OrderItem.
func MapToGraphQLOrderItemFromDTO(item *dto.OrderItemResponse) *graph.OrderItem {
	return &graph.OrderItem{
		ID:            item.ID,
		OrderID:       item.OrderID,
		ProductID:     item.ProductID,
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
		Shipping: dto.Shipping{
			CarrierID: input.Shipping.CarrierID,
			FromAddress: dto.FromAddress{
				City:       input.Shipping.FromAddress.City,
				State:      input.Shipping.FromAddress.State,
				PostalCode: input.Shipping.FromAddress.PostalCode,
				Country:    input.Shipping.FromAddress.Country,
			},
			ToAddress: dto.ToAddress{
				City:       input.Shipping.ToAddress.City,
				State:      input.Shipping.ToAddress.State,
				PostalCode: input.Shipping.ToAddress.PostalCode,
				Country:    input.Shipping.ToAddress.Country,
			},
			WeightKG: input.Shipping.WeightKg,
			Dimensions: dto.Dimensions{
				Length: input.Shipping.Dimensions.Length,
				Height: input.Shipping.Dimensions.Height,
				Width:  input.Shipping.Dimensions.Width,
				Unit:   input.Shipping.Dimensions.Unit,
			},
		},
		PaymentMethod:  input.PaymentMethod,
		PaymentGateway: input.PaymentGateway,
		Currency:       input.Currency,
	}, nil
}
