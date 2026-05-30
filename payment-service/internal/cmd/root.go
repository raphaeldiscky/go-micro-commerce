package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the payment-service root command with one subcommand per
// deployable role plus an "all" command that runs every role in one process.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "payment-service",
		Short: "Payment service binary",
	}

	root.AddCommand(
		newServeCmd(),
		newGRPCCmd(),
		newKafkaConsumerCmd(),
		newOutboxCmd(),
		newSchedulerCmd(),
		newAllCmd(),
	)

	return root
}
