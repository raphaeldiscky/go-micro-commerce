// Package main is the entry point for the product-service binary. It delegates
// to the cobra command tree in internal/cmd, where each subcommand maps to a
// deployable role (serve, grpc) and "all" runs every role in a single process.
package main

import (
	"os"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
