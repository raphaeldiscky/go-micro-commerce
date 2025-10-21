package mapper

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// MapToGraphQLOrder maps entity.Order to graph.Order.
func MapToGraphQLOrder(order *entity.Order) *graph.Order {
	items := make([]*graph.OrderItem, len(order.Items))
	for i := range order.Items {
		items[i] = MapToGraphQLOrderItem(&order.Items[i])
	}

	return &graph.Order{
		ID:             order.ID.String(),
		IdempotencyKey: order.IdempotencyKey.String(),
		CustomerID:     order.CustomerID.String(),
		Status:         order.Status,
		Currency:       order.Currency,
		PaymentGateway: order.PaymentGateway,
		PaymentMethod:  order.PaymentMethod,
		ShippingCost:   order.ShippingCost.String(),
		Subtotal:       order.Subtotal.String(),
		TotalPrice:     order.TotalPrice.String(),
		TotalTax:       order.TotalTax.String(),
		TotalDiscount:  order.TotalDiscount.String(),
		Items:          items,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

// MapToGraphQLOrderFromDTO maps dto.OrderResponse to graph.Order.
func MapToGraphQLOrderFromDTO(order *dto.OrderResponse) *graph.Order {
	items := make([]*graph.OrderItem, len(order.Items))
	for i := range order.Items {
		items[i] = MapToGraphQLOrderItemFromDTO(&order.Items[i])
	}

	return &graph.Order{
		ID:             order.ID.String(),
		CustomerID:     order.CustomerID.String(),
		Status:         order.Status,
		Currency:       order.Currency,
		PaymentGateway: order.PaymentGateway,
		PaymentMethod:  order.PaymentMethod,
		ShippingCost:   order.ShippingCost.String(),
		Subtotal:       order.Subtotal.String(),
		TotalPrice:     order.TotalPrice.String(),
		TotalTax:       order.TotalTax.String(),
		TotalDiscount:  order.TotalDiscount.String(),
		Items:          items,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

// MapToGraphQLOrderItem maps entity.OrderItem to graph.OrderItem.
func MapToGraphQLOrderItem(item *entity.OrderItem) *graph.OrderItem {
	return &graph.OrderItem{
		ID:            item.ID.String(),
		OrderID:       item.OrderID.String(),
		ProductID:     item.ProductID.String(),
		Quantity:      int(item.Quantity),
		UnitPrice:     item.UnitPrice.String(),
		TotalPrice:    item.TotalPrice.String(),
		TaxRate:       item.TaxRate.String(),
		TotalTax:      item.TotalTax.String(),
		TotalDiscount: item.TotalDiscount.String(),
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

// MapToGraphQLOrderItemFromDTO maps dto.OrderItemResponse to graph.OrderItem.
func MapToGraphQLOrderItemFromDTO(item *dto.OrderItemResponse) *graph.OrderItem {
	return &graph.OrderItem{
		ID:            item.ID.String(),
		ProductID:     item.ProductID.String(),
		Quantity:      int(item.Quantity),
		UnitPrice:     item.UnitPrice.String(),
		TotalPrice:    item.TotalPrice.String(),
		TaxRate:       item.TaxRate.String(),
		TotalTax:      item.TotalTax.String(),
		TotalDiscount: item.TotalDiscount.String(),
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
	idempotencyKey, err := uuid.Parse(input.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("invalid idempotency key: %w", err)
	}

	items := make([]dto.CreateOrderItemRequest, len(input.Items))

	for i, item := range input.Items {
		productID, parseErr := uuid.Parse(item.ProductID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid product ID at index %d: %w", i, parseErr)
		}

		items[i] = dto.CreateOrderItemRequest{
			ProductID: productID,
			Quantity:  int64(item.Quantity),
		}
	}

	// Parse shipping dimensions
	length, err := decimal.NewFromString(input.Shipping.Dimensions.Length)
	if err != nil {
		return nil, fmt.Errorf("invalid shipping dimension length: %w", err)
	}

	height, err := decimal.NewFromString(input.Shipping.Dimensions.Height)
	if err != nil {
		return nil, fmt.Errorf("invalid shipping dimension height: %w", err)
	}

	width, err := decimal.NewFromString(input.Shipping.Dimensions.Width)
	if err != nil {
		return nil, fmt.Errorf("invalid shipping dimension width: %w", err)
	}

	weightKG, err := decimal.NewFromString(input.Shipping.WeightKg)
	if err != nil {
		return nil, fmt.Errorf("invalid shipping weight: %w", err)
	}

	return &dto.CreateOrderRequest{
		CustomerID:     customerID,
		CustomerEmail:  customerEmail,
		IdempotencyKey: idempotencyKey,
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
			WeightKG: weightKG,
			Dimensions: dto.Dimensions{
				Length: length,
				Height: height,
				Width:  width,
				Unit:   input.Shipping.Dimensions.Unit,
			},
		},
		PaymentMethod:  input.PaymentMethod,
		PaymentGateway: input.PaymentGateway,
		Currency:       input.Currency,
	}, nil
}
