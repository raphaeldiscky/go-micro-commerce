package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the cart-service root command with one subcommand per
// deployable role plus an "all" command that runs every role in one process.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "cart-service",
		Short: "Cart service binary",
	}

	root.AddCommand(
		newAPICmd(),
		newGRPCCmd(),
		newWorkerCmd(),
		newAllCmd(),
	)

	return root
}
