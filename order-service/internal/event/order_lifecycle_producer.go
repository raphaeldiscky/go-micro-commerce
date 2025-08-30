package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// OrderItemPayload holds the data for each item in the order.
type OrderItemPayload struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// OrderLifecyclePayload holds the data for the Order Lifecycle event.
type OrderLifecyclePayload struct {
	OrderID    uuid.UUID            `json:"order_id"`
	UserID     uuid.UUID            `json:"user_id"`
	Status     constant.OrderStatus `json:"status"`
	TotalPrice decimal.Decimal      `json:"total_price"`
	Items      []OrderItemPayload   `json:"items"`
}

// OrderPaymentRequestPayload holds the data for payment request events.
type OrderPaymentRequestPayload struct {
	OrderID       uuid.UUID              `json:"order_id"`
	CustomerID    uuid.UUID              `json:"customer_id"`
	TotalPrice    decimal.Decimal        `json:"total_price"`
	Currency      string                 `json:"currency"`
	PaymentMethod constant.PaymentMethod `json:"payment_method"`
}

// OrderLifecycleEvent is the envelope for all Order events.
type OrderLifecycleEvent struct {
	Metadata mq.KafkaMetadata      `json:"metadata"`
	Payload  OrderLifecyclePayload `json:"payload"`
}

// OrderPaymentRequestEvent is the envelope for payment request events.
type OrderPaymentRequestEvent struct {
	Metadata mq.KafkaMetadata           `json:"metadata"`
	Payload  OrderPaymentRequestPayload `json:"payload"`
}

// GetPayload returns the data associated with the OrderPaymentRequestEvent.
func (e *OrderPaymentRequestEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the OrderPaymentRequestEvent.
func (e *OrderPaymentRequestEvent) GetMetadata() mq.KafkaMetadata {
	return e.Metadata
}

// GetPayload returns the data associated with the OrderLifecycleEvent.
func (e *OrderLifecycleEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the OrderLifecycleEvent.
func (e *OrderLifecycleEvent) GetMetadata() mq.KafkaMetadata { // Use the correct type from mq package
	return e.Metadata
}

// OrderLifecycleProducer is responsible for producing Order Lifecycle events.
type OrderLifecycleProducer struct {
	Producer *mq.KafkaAsyncProducer
	topic    string
}

// NewOrderLifecycleEvent creates a new OrderLifecycleEvent.
func NewOrderLifecycleEvent(
	orderID uuid.UUID,
	newStatus constant.OrderStatus,
	userID uuid.UUID,
	totalPrice decimal.Decimal,
	items []entity.OrderItem,
) *OrderLifecycleEvent {
	return &OrderLifecycleEvent{
		Metadata: mq.KafkaMetadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   mapStatusToEventType(newStatus),
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceOrderService,
		},
		Payload: OrderLifecyclePayload{
			OrderID:    orderID,
			UserID:     userID,
			Status:     newStatus,
			TotalPrice: totalPrice,
			Items:      mapOrderItemsToPayload(items),
		},
	}
}

// NewOrderPaymentRequestEvent creates a new OrderPaymentRequestEvent.
func NewOrderPaymentRequestEvent(
	orderID uuid.UUID,
	customerID uuid.UUID,
	totalPrice decimal.Decimal,
	currency string,
	paymentMethod constant.PaymentMethod,
) *OrderPaymentRequestEvent {
	return &OrderPaymentRequestEvent{
		Metadata: mq.KafkaMetadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeOrderPaymentRequested,
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceOrderService,
		},
		Payload: OrderPaymentRequestPayload{
			OrderID:       orderID,
			CustomerID:    customerID,
			TotalPrice:    totalPrice,
			Currency:      currency,
			PaymentMethod: paymentMethod,
		},
	}
}

// NewOrderLifecycleProducer creates a new instance of OrderLifecycleProducer.
func NewOrderLifecycleProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducerInterface {
	return &OrderLifecycleProducer{
		Producer: producer,
		topic:    constant.TopicOrderLifecycle,
	}
}

// Send implements the KafkaProducer interface.
func (p *OrderLifecycleProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *OrderLifecycleProducer) Topic() string {
	return p.topic
}
