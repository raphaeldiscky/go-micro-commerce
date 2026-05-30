package cmd

import (
	"github.com/spf13/cobra"
)

// newAllCmd runs every role in a single process. This is the
// single-deployment entry point; each role can later be split into its own
// deployment by running its dedicated subcommand instead.
func newAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Run all roles in a single process",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := bootstrap(cmd.Context())
			if err != nil {
				return err
			}
			defer app.stop()

			runners := []Runner{
				newHTTPRunner(app.ctx, app.cfg, app.logger, app.providers),
				newWebSocketRunner(app.cfg, app.logger, app.providers),
			}

			deregisterHTTP := registerConsulHTTP(app.cfg, app.logger)
			defer deregisterHTTP()

			deregisterWebSocket := registerConsulWebSocket(app.cfg, app.logger)
			defer deregisterWebSocket()

			return newManager(app.cfg, app.logger).run(app.ctx, runners...)
		},
	}
}
