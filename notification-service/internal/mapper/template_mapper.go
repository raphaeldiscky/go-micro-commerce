// Package mapper provides functions for mapping.
package mapper

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
)

// MapToOrderConfirmedTemplateData converts event data to template data.
func MapToOrderConfirmedTemplateData(
	customerName, orderID, orderDate, customerEmail string,
	items []dto.OrderItemTemplateData,
	currency string,
	subtotal, shippingCost, totalTax, totalDiscount, totalPrice decimal.Decimal,
	trackingNumber *string,
	estimatedDelivery *time.Time,
) *dto.OrderConfirmedTemplateData {
	data := &dto.OrderConfirmedTemplateData{
		CustomerName:   customerName,
		OrderID:        orderID,
		OrderDate:      orderDate,
		CustomerEmail:  customerEmail,
		Items:          items,
		Currency:       currency,
		Subtotal:       currency + " " + subtotal.String(),
		ShippingCost:   currency + " " + shippingCost.String(),
		TotalTax:       currency + " " + totalTax.String(),
		TotalDiscount:  currency + " " + totalDiscount.String(),
		TotalPrice:     currency + " " + totalPrice.String(),
		TrackingNumber: trackingNumber,
	}

	// Handle optional estimated delivery date
	if estimatedDelivery != nil {
		data.EstimatedDelivery = estimatedDelivery.Format("January 2, 2006")
	}

	return data
}

// MapToOrderDeliveredTemplateData converts event data to template data.
func MapToOrderDeliveredTemplateData(
	customerName, orderID, orderDate, customerEmail string,
	items []dto.OrderItemTemplateData,
	currency string,
	subtotal, shippingCost, totalTax, totalDiscount, totalPrice decimal.Decimal,
	trackingNumber *string,
	estimatedDelivery *time.Time,
	actualDeliveryAt time.Time,
) *dto.OrderDeliveredTemplateData {
	data := &dto.OrderDeliveredTemplateData{
		CustomerName:     customerName,
		OrderID:          orderID,
		OrderDate:        orderDate,
		CustomerEmail:    customerEmail,
		Items:            items,
		Currency:         currency,
		Subtotal:         currency + " " + subtotal.String(),
		ShippingCost:     currency + " " + shippingCost.String(),
		TotalTax:         currency + " " + totalTax.String(),
		TotalDiscount:    currency + " " + totalDiscount.String(),
		TotalPrice:       currency + " " + totalPrice.String(),
		TrackingNumber:   trackingNumber,
		ActualDeliveryAt: actualDeliveryAt.Format("January 2, 2006 at 3:04pm (MST)"),
	}

	// Handle optional estimated delivery date
	if estimatedDelivery != nil {
		data.EstimatedDelivery = estimatedDelivery.Format("January 2, 2006")
	}

	return data
}

// MapToOrderPaymentRequiredTemplateData converts event data to template data.
func MapToOrderPaymentRequiredTemplateData(
	customerName, orderID, orderDate, customerEmail string,
	items []dto.OrderItemTemplateData,
	currency string,
	subtotal, shippingCost, totalTax, totalDiscount, totalPrice decimal.Decimal,
	paymentDeadline time.Time,
	paymentURL *string,
) *dto.OrderPaymentRequiredTemplateData {
	data := &dto.OrderPaymentRequiredTemplateData{
		CustomerName:    customerName,
		OrderID:         orderID,
		OrderDate:       orderDate,
		CustomerEmail:   customerEmail,
		Items:           items,
		Currency:        currency,
		Subtotal:        currency + " " + subtotal.String(),
		ShippingCost:    currency + " " + shippingCost.String(),
		TotalTax:        currency + " " + totalTax.String(),
		TotalDiscount:   currency + " " + totalDiscount.String(),
		TotalPrice:      currency + " " + totalPrice.String(),
		PaymentDeadline: paymentDeadline.Format("January 2, 2006 at 3:04pm (MST)"),
	}

	if paymentURL != nil {
		data.PaymentURL = *paymentURL
	}

	return data
}
