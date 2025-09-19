package constant

import "time"

const (
	// GRPCMaxAttempts is the maximum number of attempts for gRPC requests.
	GRPCMaxAttempts = 3
	// GRPCInitialBackoff is the initial backoff for gRPC requests.
	GRPCInitialBackoff = 100 * time.Millisecond
	// GRPCBackoffMultiplier is the backoff multiplier for gRPC requests.
	GRPCBackoffMultiplier = 1.5
	// GRPCMaxBackoff is the maximum backoff for gRPC requests.
	GRPCMaxBackoff = 5 * time.Second
	// GRPCKeepaliveTime is the keepalive time for gRPC connections.
	GRPCKeepaliveTime = 30 * time.Second
	// GRPCKeepaliveTimeout is the keepalive timeout for gRPC connections.
	GRPCKeepaliveTimeout = 5 * time.Second
)
