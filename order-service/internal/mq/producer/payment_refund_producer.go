package producer

import (
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/shopspring/decimal"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// PaymentRefundEvent is the envelope for payment refund events.
type PaymentRefundEvent struct {
	Metadata event.Metadata             `json:"metadata"`
	Payload  event.PaymentRefundPayload `json:"payload"`
}

// NewPaymentRefundEvent creates a new PaymentRefundEvent.
func NewPaymentRefundEvent(
	orderID uuid.UUID,
	customerID uuid.UUID,
	totalPrice decimal.Decimal,
	currency string,
) *PaymentRefundEvent {
	return &PaymentRefundEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.PaymentRefundedEventType,
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.OrderServiceName,
		},
		Payload: event.PaymentRefundPayload{
			OrderID:    orderID,
			CustomerID: customerID,
			Amount:     totalPrice,
			Currency:   currency,
			Reason:     "order_canceled",
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// GetPayload returns the data associated with the PaymentRefundEvent.
func (e *PaymentRefundEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the PaymentRefundEvent.
func (e *PaymentRefundEvent) GetMetadata() event.Metadata {
	return e.Metadata
}
