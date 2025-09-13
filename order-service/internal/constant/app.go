package constant

import "time"

const (
	// AppLoggerLevel is the log level for the application.
	AppLoggerLevel = 1
	// AppTimeoutShutdown is the timeout for the application shutdown.
	AppTimeoutShutdown = 10 * time.Second
)
