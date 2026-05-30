package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the search-service root command with one subcommand per
// deployable role plus an "all" command that runs every role in one process.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "search-service",
		Short: "Search service binary",
	}

	root.AddCommand(
		newServeCmd(),
		newKafkaConsumerCmd(),
		newInboxCmd(),
		newAllCmd(),
	)

	return root
}
