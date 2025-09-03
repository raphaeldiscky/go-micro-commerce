package constant

const (
	// OrderLifecycleTopicNumPartitions is the number of partitions for the order lifecycle .
	OrderLifecycleTopicNumPartitions = 3
	// OrderLifecycleTopicReplicationFactor is the replication factor for the order lifecycle .
	OrderLifecycleTopicReplicationFactor = 1
	// PaymentGatewayRequestTopicNumPartitions is the number of partitions for the payment request .
	PaymentGatewayRequestTopicNumPartitions = 3
	// PaymentGatewayRequestTopicReplicationFactor is the replication factor for the payment request .
	PaymentGatewayRequestTopicReplicationFactor = 1
	// OrderDLQTopicNumPartitions is the number of partitions for the order DLQ .
	OrderDLQTopicNumPartitions = 1
	// OrderDLQTopicReplicationFactor is the replication factor for the order DLQ .
	OrderDLQTopicReplicationFactor = 1
	// PaymentDLQTopicNumPartitions is the number of partitions for the payment DLQ .
	PaymentDLQTopicNumPartitions = 1
	// PaymentDLQTopicReplicationFactor is the replication factor for the payment DLQ .
	PaymentDLQTopicReplicationFactor = 1
)
