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
		CourierID:           fulfillment.CourierID,
		ShippingLabelURL:    fulfillment.ShippingLabelURL,
		ShippingCost:        fulfillment.ShippingCost,
		Package:             fulfillment.Package,
		Destination:         fulfillment.Destination,
		Origin:              fulfillment.Origin,
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
	courier := req.GetCourier()

	weightKG, err := decimal.NewFromString(req.GetPackage().GetWeightKg())
	if err != nil {
		return nil
	}

	length, err := decimal.NewFromString(req.GetPackage().GetLength())
	if err != nil {
		return nil
	}

	width, err := decimal.NewFromString(req.GetPackage().GetWidth())
	if err != nil {
		return nil
	}

	height, err := decimal.NewFromString(req.GetPackage().GetHeight())
	if err != nil {
		return nil
	}

	return &dto.CalculateShippingRateRequest{
		CourierID: constant.CourierID(courier.GetCourierId()),
		Destination: entity.Destination{
			City:       req.GetDestination().GetCity(),
			State:      req.GetDestination().GetState(),
			PostalCode: req.GetDestination().GetPostalCode(),
			Country:    req.GetDestination().GetCountry(),
		},
		Origin: entity.Origin{
			City:       req.GetOrigin().GetCity(),
			State:      req.GetOrigin().GetState(),
			PostalCode: req.GetOrigin().GetPostalCode(),
			Country:    req.GetOrigin().GetCountry(),
		},
		Package: entity.Package{
			WeightKG: weightKG,
			Length:   length,
			Width:    width,
			Height:   height,
			Unit:     req.GetPackage().GetUnit(),
		},
		Currency: req.GetCurrency(),
	}
}
