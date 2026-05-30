package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the api-gateway root command with one subcommand per
// deployable role plus an "all" command that runs every role in one process.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "api-gateway",
		Short: "API gateway binary",
	}

	root.AddCommand(
		newServeCmd(),
		newAllCmd(),
	)

	return root
}
