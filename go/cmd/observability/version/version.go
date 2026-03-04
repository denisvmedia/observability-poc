// Package version provides the version sub-command.
package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/denisvmedia/observability-poc/internal/version"
)

// New returns the version sub-command.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("version: %s\ncommit:  %s\ndate:    %s\n", version.Version, version.Commit, version.Date)
		},
	}
}
