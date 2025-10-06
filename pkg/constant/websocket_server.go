package constant

import (
	"time"

	"golang.org/x/time/rate"
)

// WebSocket server constants.
const (
	WsServerPort                               = 8081
	WsServerGracePeriod                        = 30 * time.Second
	WsServerRequestTimeoutPeriod               = 30 * time.Second
	WsServerReadTimeout                        = 60 * time.Second // Increased
	WsServerWriteTimeout                       = 60 * time.Second // Increased
	WsServerWriteWait                          = 10 * time.Second
	WsServerIdleTimeout                        = 120 * time.Second
	WsServerReadHeaderTimeout                  = 5 * time.Second
	WsServerMaxHeaderBytes                     = 1 << 20          // 1 MB
	WsServerHSTSMaxAge                         = 31536000         // 1 year
	WsServerRateLimiter          rate.Limit    = 1000             // 1000 requests per second
	WsServerSendBufferSize       int           = 4096             // Increased
	WsServerMaxMessageSize       int64         = 512 * 1024       // 512 KB
	WsServerPongWait             time.Duration = 60 * time.Second // Increased to 60s
	WsServerPingPeriod           time.Duration = 50 * time.Second // Less than pong wait
	WsServerReadBufferSize       int           = 4096             // Increased
	WsServerWriteBufferSize      int           = 4096             // Increased
	WsCleanupTicker                            = 30 * time.Second
	WsShutdownTimeout                          = 30 * time.Second
)
