// Package mapper provides functions for mapping entity.Fulfillment to dto.FulfillmentResponse.
package mapper

import (
	"fmt"
	"strings"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/entity"
)

// MapStringToCarrierType converts a string to a CarrierType.
func MapStringToCarrierType(s string) (constant.CarrierType, error) {
	carriers := []constant.CarrierType{
		constant.CarrierTypeJNE,
		constant.CarrierTypeJT,
		constant.CarrierTypePOS,
		constant.CarrierTypeTiki,
		constant.CarrierTypeSiCepat,
		constant.CarrierTypeAnterAja,
		constant.CarrierTypeDHL,
		constant.CarrierTypeFedEx,
	}

	for _, c := range carriers {
		if strings.EqualFold(s, string(c)) {
			return c, nil
		}
	}

	return "", fmt.Errorf("invalid carrier type: %s", s)
}

// MapStringToFulfillmentStatus converts a string to a FulfillmentStatus.
func MapStringToFulfillmentStatus(s string) (constant.FulfillmentStatus, error) {
	statuses := []constant.FulfillmentStatus{
		constant.FulfillmentStatusPending,
		constant.FulfillmentStatusProcessing,
		constant.FulfillmentStatusShipped,
		constant.FulfillmentStatusInTransit,
		constant.FulfillmentStatusDelivered,
		constant.FulfillmentStatusCanceled,
		constant.FulfillmentStatusReturned,
	}

	for _, status := range statuses {
		if strings.EqualFold(s, string(status)) {
			return status, nil
		}
	}

	return "", fmt.Errorf("invalid fulfillment status: %s", s)
}

// MapToFulfillmentResponse converts domain entity to DTO response.
func MapToFulfillmentResponse(fulfillment *entity.Fulfillment) *dto.FulfillmentResponse {
	return &dto.FulfillmentResponse{
		ID:                  fulfillment.ID,
		OrderID:             fulfillment.OrderID,
		Status:              fulfillment.Status,
		TrackingNumber:      fulfillment.TrackingNumber,
		Carrier:             fulfillment.Carrier,
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

// MapStatusToEventType maps fulfillment status to Kafka event type.
func MapStatusToEventType(status constant.FulfillmentStatus) string {
	switch status {
	case constant.FulfillmentStatusPending:
		return kafka.FulfillmentCreatedEventType
	case constant.FulfillmentStatusProcessing:
		return kafka.FulfillmentProcessingEventType
	case constant.FulfillmentStatusShipped:
		return kafka.FulfillmentShippedEventType
	case constant.FulfillmentStatusInTransit:
		return kafka.FulfillmentInTransitEventType
	case constant.FulfillmentStatusDelivered:
		return kafka.FulfillmentDeliveredEventType
	case constant.FulfillmentStatusCanceled:
		return kafka.FulfillmentCanceledEventType
	case constant.FulfillmentStatusReturned:
		return kafka.FulfillmentReturnedEventType
	default:
		return kafka.FulfillmentUpdatedEventType
	}
}
