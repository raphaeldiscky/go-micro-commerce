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

// CarrierID represents the different carriers available for shipping.
type CarrierID string

const (
	// CarrierJNE represents the JNE carrier.
	CarrierJNE CarrierID = "jne"
	// CarrierJT represents the J&T carrier.
	CarrierJT CarrierID = "jt"
	// CarrierPOS represents the POS Indonesia carrier.
	CarrierPOS CarrierID = "pos"
	// CarrierSiCepat represents the SiCepat carrier.
	CarrierSiCepat CarrierID = "sicepat"
)
