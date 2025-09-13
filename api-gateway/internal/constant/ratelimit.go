package constant

import "time"

const (
	// RateLimitEnabled determines if rate limiting is enabled.
	RateLimitEnabled = true
	// RateLimitRequests is the maximum number of requests allowed in the rate limiting window.
	RateLimitRequests = 100
	// RateLimitWindow is the window of time in which the rate limit is applied.
	RateLimitWindow = 1 * time.Minute
	// RateLimitBurstLimit is the maximum number of requests allowed in a burst.
	RateLimitBurstLimit = 10
)
