// Package constant provides application-level constants.
package constant

import "time"

const (
	// AppTimeoutProxyRequest is the timeout for proxy requests to the application.
	AppTimeoutProxyRequest = 10 * time.Second
	// AppTimeoutSSEConnection is the timeout for Server-Sent Events (SSE) connections.
	AppTimeoutSSEConnection = 300 * time.Second // 5 minutes for streaming connections
	// AppTimeoutShutdown is the timeout for the application shutdown.
	AppTimeoutShutdown = 10 * time.Second
	// AppLoggerLevel is the log level for the application.
	AppLoggerLevel = 1
)
