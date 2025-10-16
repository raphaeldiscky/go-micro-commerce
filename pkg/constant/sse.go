package constant

import "time"

const (
	// SSEHeartbeatTicker is the heartbeat interval for SSE connections.
	SSEHeartbeatTicker = 2 * time.Minute
	// SSECleanupTicker is the cleanup interval for inactive SSE connections.
	SSECleanupTicker = 2 * time.Minute
	// SSEBroadcastBufferSize is the buffer size for SSE broadcast messages.
	SSEBroadcastBufferSize = 256
	// SSEMessageBufferSize is the buffer size for SSE messages.
	SSEMessageBufferSize = 256
)
