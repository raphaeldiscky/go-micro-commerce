package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// NotificationRequestEvent is the envelope for notification request events.
type NotificationRequestEvent struct {
	Metadata event.Metadata                   `json:"metadata"`
	Payload  event.NotificationRequestPayload `json:"payload"`
}

// NotificationRequestProducer is responsible for producing Notification Request events.
type NotificationRequestProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewNotificationRequestEvent creates a new order notification event.
func NewNotificationRequestEvent(
	order *entity.Order,
	products []entity.Product,
	customerEmail, customerName string,
	trackingNumber *string,
	templateID pkgconstant.TemplateIDType,
	subject string,
) *NotificationRequestEvent {
	// Prepare order items data for email template
	items := make([]event.OrderItemData, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		product := &products[i]
		items[i] = event.OrderItemData{
			ProductName: product.Name,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
		}
	}

	// Create order confirmation data
	orderData := event.OrderConfirmedData{
		OrderID:        order.ID,
		CustomerName:   customerName,
		CustomerEmail:  customerEmail,
		Items:          items,
		Subtotal:       order.Subtotal,
		ShippingCost:   order.ShippingCost,
		TotalTax:       order.TotalTax,
		TotalDiscount:  order.TotalDiscount,
		TotalPrice:     order.TotalPrice,
		Currency:       order.Currency,
		TrackingNumber: trackingNumber,
		OrderDate:      order.CreatedAt,
	}

	// Convert to map for template data
	templateData := map[string]any{
		"order":           orderData,
		"customer_name":   customerName,
		"order_id":        order.ID.String(),
		"total_price":     order.TotalPrice.String(),
		"currency":        order.Currency,
		"tracking_number": trackingNumber,
	}

	// Add template-specific data based on template ID
	switch templateID {
	case pkgconstant.TemplateOrderPaymentRequired:
		paymentDeadline := time.Now().UTC().Add(1 * time.Hour) // 1 hour deadline
		templateData["payment_deadline"] = paymentDeadline.Format(time.RFC3339)
		templateData["payment_url"] = nil // No payment URL provided
	case pkgconstant.TemplateOrderConfirmed,
		pkgconstant.TemplateOrderShipped,
		pkgconstant.TemplateOrderCanceled,
		pkgconstant.TemplateOrderDelivered:
		// No additional data needed
	}

	payload := event.NotificationRequestPayload{
		ID:               uuid.New(),
		RecipientEmail:   customerEmail,
		RecipientName:    customerName,
		NotificationType: event.NotificationTypeEmail,
		TemplateID:       templateID,
		Subject:          subject,
		Priority:         event.NotificationPriorityNormal,
		Data:             templateData,
		CreatedAt:        time.Now().UTC(),
	}

	return &NotificationRequestEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.NotificationRequestedEventType,
			AggregateID: order.ID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.OrderServiceName,
		},
		Payload: payload,
	}
}

// GetPayload returns the data associated with the NotificationRequestEvent.
func (e *NotificationRequestEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the NotificationRequestEvent.
func (e *NotificationRequestEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// NewNotificationRequestProducer creates a new instance of NotificationRequestProducer.
func NewNotificationRequestProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &NotificationRequestProducer{
		Producer: producer,
		topic:    kafka.NotificationRequestTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *NotificationRequestProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *NotificationRequestProducer) Topic() string {
	return p.topic
}
