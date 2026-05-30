package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/consul"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/server"
)

// httpRunner wraps the HTTP server as a Runner.
type httpRunner struct {
	server *server.HTTPServer
}

// newHTTPRunner creates a new HTTP runner.
func newHTTPRunner(
	cfg *config.Config,
	appLogger logger.Logger,
	tel *telemetry.Telemetry,
	providers *provider.Providers,
	gw *gateway.Gateway,
) *httpRunner {
	return &httpRunner{
		server: server.NewHTTPServer(cfg, appLogger, tel, providers, gw),
	}
}

// Name returns the name of the runner.
func (r *httpRunner) Name() string {
	return "HTTP Server"
}

// Start starts the HTTP server.
func (r *httpRunner) Start(ctx context.Context) error {
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

// Shutdown gracefully shuts down the HTTP server.
func (r *httpRunner) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}

// newServeCmd runs the HTTP API role.
func newServeCmd() *cobra.Command {
	return roleCmd("serve", "Run the HTTP API server", func(app *appContext) ([]Runner, func()) {
		runner := newHTTPRunner(app.cfg, app.logger, app.telemetry, app.providers, app.gateway)

		return []Runner{runner}, registerConsulHTTP(app.cfg, app.logger)
	})
}

// registerConsulHTTP registers the HTTP service with Consul and returns a
// deregister cleanup func. It is a no-op when Consul service discovery is
// disabled.
func registerConsulHTTP(cfg *config.Config, appLogger logger.Logger) func() {
	if cfg.ServiceDiscovery.Type != consulDiscoveryName {
		appLogger.Info("Consul service registration is disabled")

		return func() {}
	}

	consulClient, err := consul.NewServiceRegistration(
		cfg.ServiceDiscovery.ConsulAddress,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("Failed to create Consul client: %v", err)

		return func() {}
	}

	if err = consulClient.RegisterHTTP(
		cfg.App.Name,
		cfg.HTTPServer.Host,
		cfg.HTTPServer.Port,
	); err != nil {
		appLogger.Errorf("Failed to register with Consul: %v", err)

		return func() {}
	}

	appLogger.Infof("Successfully registered %s with Consul", cfg.App.Name)

	return func() {
		if deregErr := consulClient.Deregister(); deregErr != nil {
			appLogger.Errorf("Failed to deregister from Consul: %v", deregErr)
		}
	}
}
