package constant

import (
	"time"

	"golang.org/x/time/rate"
)

const (
	// HTTPServerPort is the port for the HTTP server.
	HTTPServerPort = 8086
	// HTTPServerGracePeriod is the grace period for the HTTP server.
	HTTPServerGracePeriod = 10 * time.Second
	// HTTPServerRequestTimeoutPeriod is the request timeout period for the HTTP server.
	HTTPServerRequestTimeoutPeriod = 30 * time.Second
	// HTTPServerReadTimeout is the read timeout for the HTTP server.
	HTTPServerReadTimeout = 30 * time.Second
	// HTTPServerWriteTimeout is the write timeout for the HTTP server.
	HTTPServerWriteTimeout = 30 * time.Second
	// HTTPServerIdleTimeout is the idle timeout for the HTTP server.
	HTTPServerIdleTimeout = 120 * time.Second
	// HTTPServerReadHeaderTimeout is the read header timeout for the HTTP server.
	HTTPServerReadHeaderTimeout = 10 * time.Second
	// HTTPServerMaxHeaderBytes is the maximum header bytes for the HTTP server.
	HTTPServerMaxHeaderBytes = 1048576 // 1MB
	// HTTPServerHSTSMaxAge is the maximum age for the HTTP server.
	HTTPServerHSTSMaxAge = 3600
	// HTTPServerRateLimiter is the rate limiter for the HTTP server.
	HTTPServerRateLimiter rate.Limit = 1000
)
