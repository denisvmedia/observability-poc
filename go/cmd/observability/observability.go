package main

import (
	"github.com/spf13/cobra"

	"github.com/denisvmedia/observability-poc/cmd/observability/run"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "observability",
		Short: "Observability POC server",
	}
	cmd.AddCommand(run.New())

	return cmd
}

func execute() error {
	return newRootCmd().Execute()
}
