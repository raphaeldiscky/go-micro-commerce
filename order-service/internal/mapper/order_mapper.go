// Package mapper provides functions for mapping domain entities to DTOs and vice versa.
package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"

	pb "github.com/raphaeldiscky/go-micro-commerce/proto/fulfillment/v1"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// MapToOrderResponse converts domain entity to DTO response.
func MapToOrderResponse(order *entity.Order) *dto.OrderResponse {
	return &dto.OrderResponse{
		ID:             order.ID,
		IdempotencyKey: order.IdempotencyKey,
		CustomerID:     order.CustomerID,
		Status:         order.Status,
		Currency:       order.Currency,
		PaymentGateway: order.PaymentGateway,
		ShippingCost:   order.ShippingCost,
		Subtotal:       order.Subtotal,
		TotalPrice:     order.TotalPrice,
		TotalTax:       order.TotalTax,
		TotalDiscount:  order.TotalDiscount,
		Items:          MapToOrderItemResponses(order.Items),
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

// MapToOrderItemResponses converts domain entities to DTO responses.
func MapToOrderItemResponses(items []entity.OrderItem) []dto.OrderItemResponse {
	var responses []dto.OrderItemResponse

	for i := range items {
		item := &items[i]
		responses = append(responses, dto.OrderItemResponse{
			ID:            item.ID,
			OrderID:       item.OrderID,
			ProductID:     item.ProductID,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			TaxRate:       item.TaxRate,
			TotalPrice:    item.TotalPrice,
			TotalTax:      item.TotalTax,
			TotalDiscount: item.TotalDiscount,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		})
	}

	return responses
}

// MapStatusToEventType maps order status to Kafka event type.
func MapStatusToEventType(status constant.OrderStatus) string {
	switch status {
	case constant.OrderStatusPending:
		return kafka.OrderCreatedEventType
	case constant.OrderStatusPaid:
		return kafka.OrderPaidEventType
	case constant.OrderStatusShipped:
		return kafka.OrderShippedEventType
	case constant.OrderStatusDelivered:
		return kafka.OrderDeliveredEventType
	case constant.OrderStatusCanceled:
		return kafka.OrderCanceledEventType
	case constant.OrderStatusPaymentPending:
		return kafka.OrderPaymentPendingEventType
	case constant.OrderStatusPaymentExpired:
		return kafka.OrderPaymentExpiredEventType
	case constant.OrderStatusProcessing:
		return kafka.OrderProcessingEventType
	case constant.OrderStatusFailed:
		return kafka.OrderFailedEventType
	case constant.OrderStatusCompleted:
		return kafka.OrderCompletedEventType
	default:
		return "unknown"
	}
}

// MapOrderItemsToPayload maps order items to their payload representation.
func MapOrderItemsToPayload(items []entity.OrderItem) []kafkaevent.OrderItemPayload {
	payloadItems := make([]kafkaevent.OrderItemPayload, len(items))

	for i := range items {
		item := &items[i]
		payloadItems[i] = kafkaevent.OrderItemPayload{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return payloadItems
}

// MapShippingDtoToEventPayload maps shipping details to their payload representation.
func MapShippingDtoToEventPayload(shipping *dto.Shipping) kafkaevent.Shipping {
	return kafkaevent.Shipping{
		CarrierID: shipping.CarrierID,
		FromAddress: kafkaevent.FromAddressPayload{
			City:       shipping.FromAddress.City,
			State:      shipping.FromAddress.State,
			PostalCode: shipping.FromAddress.PostalCode,
			Country:    shipping.FromAddress.Country,
		},
		ToAddress: kafkaevent.ToAddressPayload{
			City:       shipping.ToAddress.City,
			State:      shipping.ToAddress.State,
			PostalCode: shipping.ToAddress.PostalCode,
			Country:    shipping.ToAddress.Country,
		},
		WeightKG: shipping.WeightKG,
		Dimensions: kafkaevent.Dimensions{
			Width:  shipping.Dimensions.Width,
			Height: shipping.Dimensions.Height,
			Length: shipping.Dimensions.Length,
			Unit:   shipping.Dimensions.Unit,
		},
	}
}

// MapShippingDtoToProto maps shipping details to their protobuf representation.
func MapShippingDtoToProto(shipping *dto.Shipping) *pb.Shipping {
	return &pb.Shipping{
		CarrierId: shipping.CarrierID,
		FromAddress: &pb.FromAddress{
			City:       shipping.FromAddress.City,
			State:      shipping.FromAddress.State,
			PostalCode: shipping.FromAddress.PostalCode,
			Country:    shipping.FromAddress.Country,
		},
		ToAddress: &pb.ToAddress{
			City:       shipping.ToAddress.City,
			State:      shipping.ToAddress.State,
			PostalCode: shipping.ToAddress.PostalCode,
			Country:    shipping.ToAddress.Country,
		},
		WeightKg: shipping.WeightKG.InexactFloat64(),
		Dimensions: &pb.Dimensions{
			Width:  shipping.Dimensions.Width.InexactFloat64(),
			Height: shipping.Dimensions.Height.InexactFloat64(),
			Length: shipping.Dimensions.Length.InexactFloat64(),
			Unit:   shipping.Dimensions.Unit,
		},
	}
}
