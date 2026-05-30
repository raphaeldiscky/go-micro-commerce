package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the order-service root command with one subcommand per
// deployable role plus an "all" command that runs every role in one process.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "order-service",
		Short: "Order service binary",
	}

	root.AddCommand(
		newServeCmd(),
		newKafkaConsumerCmd(),
		newOutboxCmd(),
		newSchedulerCmd(),
		newTemporalCmd(),
		newAsynqCmd(),
		newAllCmd(),
	)

	return root
}
