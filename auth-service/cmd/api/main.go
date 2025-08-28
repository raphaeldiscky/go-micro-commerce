// Package main implements the API for the product service.
package main

import (
	"log"

	"github.com/raphaeldiscky/go-micro-template/pkg/consul"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/auth-service/cmd/worker"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.Logger.Level)

	consulCleanup := setupConsulRegistration(cfg)
	defer consulCleanup()

	worker.Start(cfg, appLogger)
}

// setupConsulRegistration handles Consul service registration and returns a cleanup function.
func setupConsulRegistration(cfg *config.Config) func() {
	if !cfg.Consul.Enabled {
		log.Println("Consul service discovery is disabled")

		return func() {}
	}

	consulClient, err := consul.NewServiceRegistration(cfg.Consul.Address)
	if err != nil {
		log.Printf("Failed to create Consul client: %v", err)

		return func() {}
	}

	if err := consulClient.RegisterHTTP(cfg.Consul.ServiceName, cfg.Consul.ServiceHost, cfg.HTTPServer.Port); err != nil {
		log.Printf("Failed to register with Consul: %v", err)

		return func() {}
	}

	log.Printf("Service registered with Consul: %s at %s:%d",
		cfg.Consul.ServiceName, cfg.Consul.ServiceHost, cfg.HTTPServer.Port)

	return func() {
		if err := consulClient.Deregister(); err != nil {
			log.Printf("Failed to deregister from Consul: %v", err)
		} else {
			log.Println("Service deregistered from Consul")
		}
	}
}
