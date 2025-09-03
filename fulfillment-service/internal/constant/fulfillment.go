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

// CarrierType represents the different carriers available for shipping.
type CarrierType string

const (
	// CarrierTypeJNE represents the JNE carrier.
	CarrierTypeJNE CarrierType = "jne"
	// CarrierTypeJT represents the J&T carrier.
	CarrierTypeJT CarrierType = "j&t"
	// CarrierTypePOS represents the POS Indonesia carrier.
	CarrierTypePOS CarrierType = "pos"
	// CarrierTypeTiki represents the Tiki carrier.
	CarrierTypeTiki CarrierType = "tiki"
	// CarrierTypeSiCepat represents the SiCepat carrier.
	CarrierTypeSiCepat CarrierType = "sicepat"
	// CarrierTypeAnterAja represents the Anteraja carrier.
	CarrierTypeAnterAja CarrierType = "anteraja"
	// CarrierTypeDHL represents the DHL carrier.
	CarrierTypeDHL CarrierType = "dhl"
	// CarrierTypeFedEx represents the FedEx carrier.
	CarrierTypeFedEx CarrierType = "fedex"
)

// Kafka Topic Configuration Constants.
const (
	// FulfillmentLifecycleTopicNumPartitions defines the number of partitions for fulfillment lifecycle topic.
	FulfillmentLifecycleTopicNumPartitions = 3
	// FulfillmentLifecycleTopicReplicationFactor defines the replication factor for fulfillment lifecycle topic.
	FulfillmentLifecycleTopicReplicationFactor = 1
	// FulfillmentDLQTopicNumPartitions defines the number of partitions for fulfillment DLQ topic.
	FulfillmentDLQTopicNumPartitions = 3
	// FulfillmentDLQTopicReplicationFactor defines the replication factor for fulfillment DLQ topic.
	FulfillmentDLQTopicReplicationFactor = 1
)
