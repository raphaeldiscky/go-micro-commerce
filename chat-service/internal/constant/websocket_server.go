package constant

import (
	"time"

	"golang.org/x/time/rate"
)

// WebSocket server constants.
const (
	WsServerPort                            = 8081
	WsServerGracePeriod                     = 30 * time.Second
	WsServerRequestTimeoutPeriod            = 30 * time.Second
	WsServerReadTimeout                     = 10 * time.Second
	WsServerWriteTimeout                    = 10 * time.Second
	WsServerIdleTimeout                     = 120 * time.Second
	WsServerReadHeaderTimeout               = 5 * time.Second
	WsServerMaxHeaderBytes                  = 1 << 20  // 1 MB
	WsServerHSTSMaxAge                      = 31536000 // 1 year
	WsServerRateLimiter          rate.Limit = 1000     // 1000 requests per second
)
