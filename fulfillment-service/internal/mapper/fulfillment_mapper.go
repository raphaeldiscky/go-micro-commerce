// Package mapper provides functions for mapping entity.Fulfillment to dto.FulfillmentResponse.
package mapper

import (
	"github.com/shopspring/decimal"

	pb "github.com/raphaeldiscky/go-micro-commerce/proto/fulfillment/v1"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/entity"
)

// MapToFulfillmentResponse converts domain entity to DTO response.
func MapToFulfillmentResponse(fulfillment *entity.Fulfillment) *dto.FulfillmentResponse {
	return &dto.FulfillmentResponse{
		ID:                  fulfillment.ID,
		OrderID:             fulfillment.OrderID,
		Status:              fulfillment.Status,
		TrackingNumber:      fulfillment.TrackingNumber,
		CarrierID:           fulfillment.CarrierID,
		ShippingLabelURL:    fulfillment.ShippingLabelURL,
		ShippingCost:        fulfillment.ShippingCost,
		WeightKG:            fulfillment.WeightKG,
		Dimensions:          fulfillment.Dimensions,
		EstimatedDeliveryAt: fulfillment.EstimatedDeliveryAt,
		ActualDeliveryAt:    fulfillment.ActualDeliveryAt,
		CreatedAt:           fulfillment.CreatedAt,
		UpdatedAt:           fulfillment.UpdatedAt,
	}
}

// MapToCalculateShippingRateRequest maps a protobuf shipping request to a domain request.
func MapToCalculateShippingRateRequest(
	req *pb.GetShippingCostRequest,
) *dto.CalculateShippingRateRequest {
	return &dto.CalculateShippingRateRequest{
		CarrierID: constant.CarrierID(req.GetShipping().GetCarrierId()),
		Dimensions: entity.Dimensions{
			Width:  decimal.NewFromFloat(req.GetShipping().GetDimensions().GetWidth()),
			Height: decimal.NewFromFloat(req.GetShipping().GetDimensions().GetHeight()),
			Length: decimal.NewFromFloat(req.GetShipping().GetDimensions().GetLength()),
			Unit:   req.GetShipping().GetDimensions().GetUnit(),
		},
		WeightKG: decimal.NewFromFloat(req.GetShipping().GetWeightKg()),
		Currency: req.GetCurrency(),
		FromAddress: entity.FromAddress{
			City:       req.GetShipping().GetFromAddress().GetCity(),
			State:      req.GetShipping().GetFromAddress().GetState(),
			PostalCode: req.GetShipping().GetFromAddress().GetPostalCode(),
			Country:    req.GetShipping().GetFromAddress().GetCountry(),
		},
		ToAddress: entity.ToAddress{
			City:       req.GetShipping().GetToAddress().GetCity(),
			State:      req.GetShipping().GetToAddress().GetState(),
			PostalCode: req.GetShipping().GetToAddress().GetPostalCode(),
			Country:    req.GetShipping().GetToAddress().GetCountry(),
		},
	}
}
