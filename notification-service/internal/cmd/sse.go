package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/consul"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/server"
)

// sseRunner wraps the SSE server as a Runner.
type sseRunner struct {
	server *server.SSEServer
}

// newSSERunner creates a new SSE runner.
func newSSERunner(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *sseRunner {
	return &sseRunner{
		server: server.NewSSEServer(cfg, appLogger, providers),
	}
}

// Name returns the name of the runner.
func (r *sseRunner) Name() string {
	return "SSE Server"
}

// Start starts the SSE server.
func (r *sseRunner) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if err := r.server.Start(); err != nil {
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

// Shutdown gracefully shuts down the SSE server.
func (r *sseRunner) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}

// newSSECmd runs the SSE server role.
func newSSECmd() *cobra.Command {
	return roleCmd("sse", "Run the SSE server", func(app *appContext) ([]Runner, func()) {
		runner := newSSERunner(app.cfg, app.logger, app.providers)

		return []Runner{runner}, registerConsulSSE(app.cfg, app.logger)
	})
}

// registerConsulSSE registers the SSE service with Consul and returns a
// deregister cleanup func. It is a no-op when Consul is disabled.
func registerConsulSSE(cfg *config.Config, appLogger logger.Logger) func() {
	if !cfg.Consul.Enabled {
		appLogger.Infof("Consul service discovery is disabled")

		return func() {}
	}

	consulClient, err := consul.NewServiceRegistration(cfg.Consul.Address, appLogger)
	if err != nil {
		return func() {}
	}

	if err = consulClient.RegisterSSE(
		cfg.App.Name+"-sse",
		cfg.SSEServer.Host,
		cfg.SSEServer.Port,
	); err != nil {
		return func() {}
	}

	return func() {
		if deregErr := consulClient.Deregister(); deregErr != nil {
			appLogger.Errorf("Failed to deregister from Consul: %v", deregErr)
		}
	}
}
