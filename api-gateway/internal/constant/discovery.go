package constant

import "time"

const (
	// ServiceDiscoveryTimeout is the timeout for service discovery.
	ServiceDiscoveryTimeout = 10 * time.Second
	// ServiceDiscoveryConsulRefreshInterval is the interval for refreshing Consul service discovery.
	ServiceDiscoveryConsulRefreshInterval = 10 * time.Second
)
