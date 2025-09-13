package constant

import "time"

const (
	// CircuitBreakerMaxRequests is the maximum number of requests allowed in the circuit breaker.
	CircuitBreakerMaxRequests = 3
	// CircuitBreakerInterval is the interval for checking the circuit breaker.
	CircuitBreakerInterval = 1 * time.Second
	// CircuitBreakerTimeout is the timeout for the circuit breaker.
	CircuitBreakerTimeout = 5 * time.Second
)
