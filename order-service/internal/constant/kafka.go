package constant

const (
	// OrderLifecycleTopicNumPartitions is the number of partitions for the order lifecycle .
	OrderLifecycleTopicNumPartitions = 3
	// OrderLifecycleTopicReplicationFactor is the replication factor for the order lifecycle .
	OrderLifecycleTopicReplicationFactor = 1
	// PaymentRequestTopicNumPartitions is the number of partitions for the payment request .
	PaymentRequestTopicNumPartitions = 3
	// PaymentRequestTopicReplicationFactor is the replication factor for the payment request .
	PaymentRequestTopicReplicationFactor = 1
	// OrderDLQTopicNumPartitions is the number of partitions for the order DLQ .
	OrderDLQTopicNumPartitions = 1
	// OrderDLQTopicReplicationFactor is the replication factor for the order DLQ .
	OrderDLQTopicReplicationFactor = 1
	// PaymentDLQTopicNumPartitions is the number of partitions for the payment DLQ .
	PaymentDLQTopicNumPartitions = 1
	// PaymentDLQTopicReplicationFactor is the replication factor for the payment DLQ .
	PaymentDLQTopicReplicationFactor = 1
	// FulfillmentRequestTopicNumPartitions is the number of partitions for the fulfillment request .
	FulfillmentRequestTopicNumPartitions = 3
	// FulfillmentRequestTopicReplicationFactor is the replication factor for the fulfillment request .
	FulfillmentRequestTopicReplicationFactor = 1
	// FulfillmentDLQTopicNumPartitions is the number of partitions for the fulfillment DLQ .
	FulfillmentDLQTopicNumPartitions = 1
	// FulfillmentDLQTopicReplicationFactor is the replication factor for the fulfillment DLQ .
	FulfillmentDLQTopicReplicationFactor = 1
)
