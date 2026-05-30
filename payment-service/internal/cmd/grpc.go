package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/consul"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/server"
)

// grpcRunner wraps the gRPC server as a Runner.
type grpcRunner struct {
	server *server.GRPCServer
}

// newGRPCRunner creates a new gRPC runner.
func newGRPCRunner(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *grpcRunner {
	return &grpcRunner{
		server: server.NewGRPCServer(providers.PaymentService, appLogger, cfg),
	}
}

// Name returns the name of the runner.
func (r *grpcRunner) Name() string {
	return "gRPC Server"
}

// Start starts the gRPC server.
func (r *grpcRunner) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if err := r.server.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil // Context canceled, normal shutdown
	case err := <-errChan:
		return err // Server error
	}
}

// Shutdown gracefully shuts down the gRPC server.
func (r *grpcRunner) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}

// newGRPCCmd runs the gRPC server role.
func newGRPCCmd() *cobra.Command {
	return roleCmd("grpc", "Run the gRPC server", func(app *appContext) ([]Runner, func()) {
		runner := newGRPCRunner(app.cfg, app.logger, app.providers)

		return []Runner{runner}, registerConsulGRPC(app.cfg, app.logger)
	})
}

// registerConsulGRPC registers the Connect-RPC service with Consul and returns
// a deregister cleanup func. It is a no-op when Consul is disabled.
func registerConsulGRPC(cfg *config.Config, appLogger logger.Logger) func() {
	if !cfg.Consul.Enabled {
		appLogger.Infof("Consul service discovery is disabled")

		return func() {}
	}

	consulClient, err := consul.NewServiceRegistration(cfg.Consul.Address, appLogger)
	if err != nil {
		return func() {}
	}

	if err = consulClient.RegisterConnectRPC(
		cfg.GRPCServer.ServiceName,
		cfg.GRPCServer.Host,
		cfg.GRPCServer.Port,
	); err != nil {
		appLogger.Errorf("Failed to register Connect-RPC service with Consul: %v", err)

		return func() {}
	}

	return func() {
		if deregErr := consulClient.Deregister(); deregErr != nil {
			appLogger.Errorf("Failed to deregister from Consul: %v", deregErr)
		}
	}
}
