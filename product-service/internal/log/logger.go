// Package log provides logger-related types for the product service.
package log

import "github.com/raphaeldiscky/go-micro-template/pkg/logger"

// LoggerInterface defines the contract for a logger in the product service.
type LoggerInterface interface {
	logger.Logger
}

// Logger is the package-level logger instance for the product service.
var Logger logger.Logger

// SetLogger sets the logger for the product service.
func SetLogger(logger logger.Logger) {
	Logger = logger
}
