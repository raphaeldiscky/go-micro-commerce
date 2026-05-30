package cmd

import (
	"github.com/spf13/cobra"
)

// newWorkerCmd runs all async background workers (Kafka consumer, inbox
// processor) in a single process. It binds no request/response endpoint and
// registers nothing with Consul.
func newWorkerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "worker",
		Short: "Run the async workers (kafka consumer, inbox processor)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := bootstrap(cmd.Context())
			if err != nil {
				return err
			}
			defer app.stop()

			runners := []Runner{
				newKafkaConsumerRunner(app.cfg, app.logger, app.providers),
				newInboxProcessorRunner(app.cfg, app.logger, app.providers),
			}

			return newManager(app.cfg, app.logger).run(app.ctx, runners...)
		},
	}
}
