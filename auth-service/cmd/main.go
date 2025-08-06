// Package main implements the entry point for the auth service.
package main

import (
	"log"
	"strconv"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger (0 = Info level)
	appLogger := logger.NewZapLogger(0)

	// Create server
	srv, err := server.NewHTTPServer(cfg, appLogger)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	portStr := strconv.Itoa(cfg.HTTPServer.Port)
	if err := srv.Start(portStr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
