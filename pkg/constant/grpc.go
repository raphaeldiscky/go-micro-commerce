package constant

const (
	// GRPCHealthUnknown represents the unknown health status.
	GRPCHealthUnknown = "UNKNOWN"
	// GRPCHealthServing represents the serving health status.
	GRPCHealthServing = "SERVING"
	// GRPCHealthNotServing represents the not serving health status.
	GRPCHealthNotServing = "NOT_SERVING"
)

const (
	// GRPCServiceNameProduct represents the product service name.
	GRPCServiceNameProduct = "product-service-grpc"
	// GRPCServiceNameOrder represents the order service name.
	GRPCServiceNameOrder = "order-service-grpc"
	// GRPCServiceNamePayment represents the payment service name.
	GRPCServiceNamePayment = "payment-service-grpc"
	// GRPCServiceNameNotification represents the notification service name.
	GRPCServiceNameNotification = "notification-service-grpc"
	// GRPCServiceNameAuth represents the auth service name.
	GRPCServiceNameAuth = "auth-service-grpc"
)
