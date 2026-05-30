package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the notification-service root command with one subcommand
// per deployable role plus an "all" command that runs every role in one
// process.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "notification-service",
		Short: "Notification service binary",
	}

	root.AddCommand(
		newAPICmd(),
		newSSECmd(),
		newWorkerCmd(),
		newAllCmd(),
	)

	return root
}
