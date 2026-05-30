package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newWorkerCmd runs all async background workers (Kafka consumer, outbox
// publisher, Temporal worker, asynq worker) in a single process. It binds no
// request/response endpoint and registers nothing with Consul.
func newWorkerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "worker",
		Short: "Run the async workers (kafka consumer, outbox, temporal, asynq)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := bootstrap(cmd.Context())
			if err != nil {
				return err
			}
			defer app.stop()

			asynq, err := newAsynqRunner(app.cfg, app.logger, app.providers)
			if err != nil {
				return fmt.Errorf("create asynq runner: %w", err)
			}

			runners := []Runner{
				newKafkaConsumerRunner(app.cfg, app.logger, app.providers),
				newOutboxPublisherRunner(app.ctx, app.cfg, app.logger, app.providers),
				newTemporalRunner(app.logger, app.providers),
				asynq,
			}

			return newManager(app.cfg, app.logger).run(app.ctx, runners...)
		},
	}
}
