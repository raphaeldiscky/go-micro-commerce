package constant

import "time"

const (
	// GRPCGracePeriod is the grace period for gRPC server shutdown.
	GRPCGracePeriod = 10 * time.Second
	// GRPCPort is the port for the gRPC server.
	GRPCPort = 50055
)
