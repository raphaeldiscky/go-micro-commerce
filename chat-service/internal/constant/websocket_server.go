package constant

import (
	"time"

	"golang.org/x/time/rate"
)

// WebSocket server constants.
const (
	WebsocketServerPort                            = 8081
	WebsocketServerGracePeriod                     = 30 * time.Second
	WebsocketServerRequestTimeoutPeriod            = 30 * time.Second
	WebsocketServerReadTimeout                     = 10 * time.Second
	WebsocketServerWriteTimeout                    = 10 * time.Second
	WebsocketServerIdleTimeout                     = 120 * time.Second
	WebsocketServerReadHeaderTimeout               = 5 * time.Second
	WebsocketServerMaxHeaderBytes                  = 1 << 20  // 1 MB
	WebsocketServerHSTSMaxAge                      = 31536000 // 1 year
	WebsocketServerRateLimiter          rate.Limit = 1000     // 1000 requests per second
)
