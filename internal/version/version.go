// Package version provides build-time version information, injected via ldflags.
package version

import "fmt"

// Set at build time via -ldflags.
var (
	BuildTimestamp string
	BuildBranch    string
)

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
