package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the auth-service root command with one subcommand per
// deployable role plus an "all" command that runs every role in one process.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "auth-service",
		Short: "Auth service binary",
	}

	root.AddCommand(
		newServeCmd(),
		newAllCmd(),
	)

	return root
}
