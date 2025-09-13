package constant

import "time"

const (
	// AppTimeoutProxyRequest is the timeout for proxy requests to the application.
	AppTimeoutProxyRequest = 10 * time.Second
	// AppTimeoutShutdown is the timeout for the application shutdown.
	AppTimeoutShutdown = 10 * time.Second
	// AppLoggerLevel is the log level for the application.
	AppLoggerLevel = 1
)
