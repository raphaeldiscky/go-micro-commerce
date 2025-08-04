// Package log provides logger-related types for the product service.
package log

import "github.com/raphaeldiscky/go-micro-template/pkg/logger"

// LoggerInterface defines the contract for a logger in the product service.
type LoggerInterface interface {
	logger.Logger
}
