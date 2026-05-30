package cmd

import (
	"fmt"

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

			// Construct the asynq runner before assembling the slice so its
			// dependencies are resolved in the same order as the dedicated
			// roles.
			asynq, err := newAsynqRunner(app.cfg, app.logger, app.providers)
			if err != nil {
				return fmt.Errorf("create asynq runner: %w", err)
			}

			runners := []Runner{
				newHTTPRunner(app.ctx, app.cfg, app.logger, app.telemetry, app.providers),
				newGRPCRunner(app.cfg, app.logger, app.providers),
				newOutboxPublisherRunner(app.ctx, app.cfg, app.logger, app.providers),
				newKafkaConsumerRunner(app.cfg, app.logger, app.providers),
				asynq,
			}

			deregisterHTTP := registerConsulHTTP(app.cfg, app.logger)
			defer deregisterHTTP()

			deregisterGRPC := registerConsulGRPC(app.cfg, app.logger)
			defer deregisterGRPC()

			return newManager(app.cfg, app.logger).run(app.ctx, runners...)
		},
	}
}
