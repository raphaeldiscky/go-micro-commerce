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

			// HTTP first so the order service is constructed before the
			// asynq worker, then the async roles.
			asynq, err := newAsynqRunner(app.cfg, app.logger, app.providers)
			if err != nil {
				return fmt.Errorf("create asynq runner: %w", err)
			}

			runners := []Runner{
				newHTTPRunner(app.ctx, app.cfg, app.logger, app.providers),
				newKafkaConsumerRunner(app.cfg, app.logger, app.providers),
				newOutboxPublisherRunner(app.ctx, app.cfg, app.logger, app.providers),
				newJobSchedulerRunner(app.logger, app.providers),
				newTemporalRunner(app.logger, app.providers),
				asynq,
			}

			deregister := registerConsulHTTP(app.cfg, app.logger)
			defer deregister()

			return newManager(app.cfg, app.logger).run(app.ctx, runners...)
		},
	}
}
