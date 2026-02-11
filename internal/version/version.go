// Package version provides build-time version information, injected via ldflags.
package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set at build time via -ldflags.
var (
	Version        string
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
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), String())
		},
	}
}

// String returns a formatted version string including build metadata.
func String() string {
	ts := BuildTimestamp
	if ts == "" {
		ts = "unknown"
	}
	br := BuildBranch
	if br == "" {
		br = "unknown"
	}

	// Only include version prefix if version is set (e.g., from a release tag).
	// Local/PR builds without a version tag will show just "built {timestamp} from {branch}".
	if Version != "" {
		return fmt.Sprintf("%s built %s from %s", Version, ts, br)
	}
	return fmt.Sprintf("built %s from %s", ts, br)
}
