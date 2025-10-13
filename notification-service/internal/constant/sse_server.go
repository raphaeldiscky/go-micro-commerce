package constant

import "time"

// SSE server constants.
const (
	// SSEServerPort is the port for the SSE server.
	SSEServerPort = 8087
	// SSEServerTimeout is the timeout for Server-Sent Events (SSE) connections.
	SSEServerTimeout = 300 * time.Second // 5 minutes for streaming connections
	// SSEServerRateLimiter is the rate limiter for the SSE server.
	SSEServerRateLimiter = 100
)
