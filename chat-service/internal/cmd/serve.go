package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/consul"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/server"
)

// httpRunner wraps the HTTP server as a Runner.
type httpRunner struct {
	server *server.HTTPServer
}

// newHTTPRunner creates a new HTTP runner.
func newHTTPRunner(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *httpRunner {
	return &httpRunner{
		server: server.NewHTTPServer(ctx, cfg, appLogger, providers),
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
		runner := newHTTPRunner(app.ctx, app.cfg, app.logger, app.providers)

		return []Runner{runner}, registerConsulHTTP(app.cfg, app.logger)
	})
}

// registerConsulHTTP registers the HTTP service with Consul and returns a
// deregister cleanup func. It is a no-op when Consul is disabled.
func registerConsulHTTP(cfg *config.Config, appLogger logger.Logger) func() {
	if !cfg.Consul.Enabled {
		appLogger.Infof("Consul service discovery is disabled")

		return func() {}
	}

	consulClient, err := consul.NewServiceRegistration(cfg.Consul.Address, appLogger)
	if err != nil {
		return func() {}
	}

	if err = consulClient.RegisterHTTP(
		cfg.App.Name,
		cfg.HTTPServer.Host,
		cfg.HTTPServer.Port,
	); err != nil {
		return func() {}
	}

	return func() {
		if deregErr := consulClient.Deregister(); deregErr != nil {
			appLogger.Errorf("Failed to deregister from Consul: %v", deregErr)
		}
	}
}
