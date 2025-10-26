package constant

// FulfillmentStatus represents the status of a fulfillment.
type FulfillmentStatus string

const (
	// FulfillmentStatusPending indicates that the fulfillment is pending.
	FulfillmentStatusPending FulfillmentStatus = "pending"
	// FulfillmentStatusProcessing indicates that the fulfillment is being processed.
	FulfillmentStatusProcessing FulfillmentStatus = "processing"
	// FulfillmentStatusShipped indicates that the fulfillment has been shipped.
	FulfillmentStatusShipped FulfillmentStatus = "shipped"
	// FulfillmentStatusInTransit indicates that the fulfillment is in transit.
	FulfillmentStatusInTransit FulfillmentStatus = "in_transit"
	// FulfillmentStatusDelivered indicates that the fulfillment has been delivered.
	FulfillmentStatusDelivered FulfillmentStatus = "delivered"
	// FulfillmentStatusCanceled indicates that the fulfillment has been canceled.
	FulfillmentStatusCanceled FulfillmentStatus = "canceled"
	// FulfillmentStatusReturned indicates that the fulfillment has been returned.
	FulfillmentStatusReturned FulfillmentStatus = "returned"
)

// CourierID represents the different curiers available for shipping.
type CourierID string

const (
	// CourierJNE represents the JNE Courier.
	CourierJNE CourierID = "jne"
	// CourierJT represents the J&T Courier.
	CourierJT CourierID = "jt"
	// CourierPOS represents the POS Indonesia Courier.
	CourierPOS CourierID = "pos"
	// CourierSiCepat represents the SiCepat Courier.
	CourierSiCepat CourierID = "sicepat"
)

const (
	// MockEstimatedDeliveryDays is a mock estimated delivery days.
	MockEstimatedDeliveryDays = 7
)
