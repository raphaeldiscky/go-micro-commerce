// Package mapper provides functions for mapping.
package mapper

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
)

// MapToOrderConfirmationTemplateData converts event data to template data.
func MapToOrderConfirmationTemplateData(
	customerName, orderNumber, orderID, orderDate, customerEmail string,
	items []dto.OrderItemTemplateData,
	currency string,
	subtotal, shippingCost, totalTax, totalDiscount, totalPrice decimal.Decimal,
	trackingNumber string,
	estimatedDelivery *time.Time,
) *dto.OrderConfirmationTemplateData {
	data := &dto.OrderConfirmationTemplateData{
		CustomerName:  customerName,
		OrderNumber:   orderNumber,
		OrderID:       orderID,
		OrderDate:     orderDate,
		CustomerEmail: customerEmail,
		Items:         items,
		Currency:      currency,
		Subtotal:      currency + " " + subtotal.String(),
		ShippingCost:  currency + " " + shippingCost.String(),
		TotalTax:      currency + " " + totalTax.String(),
		TotalDiscount: currency + " " + totalDiscount.String(),
		TotalPrice:    currency + " " + totalPrice.String(),
	}

	// Add shipping info if available
	if trackingNumber != "" {
		shipping := &dto.ShippingInfoTemplateData{
			TrackingNumber: trackingNumber,
		}
		if estimatedDelivery != nil {
			shipping.EstimatedDelivery = estimatedDelivery.Format("January 2, 2006")
		}

		data.Shipping = shipping
	}

	return data
}
