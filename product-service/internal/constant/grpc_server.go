package constant

import "time"

const (
	// GRPCGracePeriod is the grace period for Connect-RPC server shutdown.
	GRPCGracePeriod = 10 * time.Second
	// GRPCPort is the port for the Connect-RPC server.
	GRPCPort = 50052
	// GRPCReadHeaderTimeout is the timeout for reading HTTP request headers.
	GRPCReadHeaderTimeout = 30 * time.Second
	// GRPCDefaultLimit is the default limit for pagination in gRPC requests.
	GRPCDefaultLimit = 10
)
