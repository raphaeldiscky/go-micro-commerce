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
	WsServerReadTimeout                        = 10 * time.Second
	WsServerWriteTimeout                       = 10 * time.Second
	WsServerWriteWait                          = 10 * time.Second
	WsServerIdleTimeout                        = 120 * time.Second
	WsServerReadHeaderTimeout                  = 5 * time.Second
	WsServerMaxHeaderBytes                     = 1 << 20  // 1 MB
	WsServerHSTSMaxAge                         = 31536000 // 1 year
	WsServerRateLimiter          rate.Limit    = 1000     // 1000 requests per second
	WsServerSendBufferSize       int           = 1024
	WsServerMaxMessageSize       int           = 1024
	WsServerPongWait             time.Duration = 10 * time.Second
	WsServerPingPeriod           time.Duration = 60 * time.Second
	WsServerReadBufferSize       int           = 1024
	WsServerWriteBufferSize      int           = 1024
	WsCleaupTicker                             = 30 * time.Second
	WsShutdownTimeout                          = 30 * time.Second
)
