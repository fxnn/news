// Package version provides build-time version information, injected via ldflags.
package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set at build time via -ldflags.
var (
	BuildTimestamp string
	BuildBranch    string
)

// NewCommand returns a cobra command that prints build version information.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print build version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), String())
		},
	}
}

func String() string {
	ts := BuildTimestamp
	if ts == "" {
		ts = "unknown"
	}
	br := BuildBranch
	if br == "" {
		br = "unknown"
	}
	return fmt.Sprintf("built %s from %s", ts, br)
}
