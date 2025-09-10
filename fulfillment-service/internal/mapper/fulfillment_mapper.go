// Package mapper provides functions for mapping entity.Fulfillment to dto.FulfillmentResponse.
package mapper

import (
	"github.com/shopspring/decimal"

	pb "github.com/raphaeldiscky/go-micro-commerce/proto/fulfillment"

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
		CarrierID: constant.CarrierID(req.Shipping.CarrierId),
		Dimensions: entity.Dimensions{
			Width:  decimal.NewFromFloat(req.Shipping.Dimensions.Width),
			Height: decimal.NewFromFloat(req.Shipping.Dimensions.Height),
			Length: decimal.NewFromFloat(req.Shipping.Dimensions.Length),
			Unit:   req.Shipping.Dimensions.Unit,
		},
		WeightKG: decimal.NewFromFloat(req.Shipping.WeightKg),
		Currency: req.Currency,
		FromAddress: entity.FromAddress{
			City:       req.Shipping.FromAddress.City,
			State:      req.Shipping.FromAddress.State,
			PostalCode: req.Shipping.FromAddress.PostalCode,
			Country:    req.Shipping.FromAddress.Country,
		},
		ToAddress: entity.ToAddress{
			City:       req.Shipping.ToAddress.City,
			State:      req.Shipping.ToAddress.State,
			PostalCode: req.Shipping.ToAddress.PostalCode,
			Country:    req.Shipping.ToAddress.Country,
		},
	}
}
