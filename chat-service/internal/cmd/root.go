package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the chat-service root command with one subcommand per
// deployable role plus an "all" command that runs every role in one process.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "chat-service",
		Short: "Chat service binary",
	}

	root.AddCommand(
		newAPICmd(),
		newWebSocketCmd(),
		newAllCmd(),
	)

	return root
}
