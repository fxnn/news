package cli

import (
	"flag"
	"fmt"
)

type Options struct {
	Maildir    string
	Storydir   string
	Config     string
	Limit      int
	Verbose    bool
	LogHeaders bool
	LogBodies  bool
	LogStories bool
}

func ParseOptions(args []string) (*Options, error) {
	fs := flag.NewFlagSet("story-extractor", flag.ContinueOnError)

	opts := &Options{}

	fs.StringVar(&opts.Maildir, "maildir", "", "Path to the Maildir directory (required)")
	fs.StringVar(&opts.Storydir, "storydir", "", "Output directory for story files (required)")
	fs.StringVar(&opts.Config, "config", "", "Path to the TOML configuration file (required)")
	fs.IntVar(&opts.Limit, "limit", 0, "Maximum number of emails to process (optional, 0 = unlimited)")
	fs.BoolVar(&opts.Verbose, "verbose", false, "Enable detailed log output")
	fs.BoolVar(&opts.LogHeaders, "log-headers", false, "Log parsed email headers (requires --verbose)")
	fs.BoolVar(&opts.LogBodies, "log-bodies", false, "Log parsed email headers and bodies (requires --verbose)")
	fs.BoolVar(&opts.LogStories, "log-stories", false, "Log extracted stories (requires --verbose)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if opts.Maildir == "" {
		return nil, fmt.Errorf("--maildir is required")
	}

	if opts.Storydir == "" {
		return nil, fmt.Errorf("--storydir is required")
	}

	if opts.Config == "" {
		return nil, fmt.Errorf("--config is required")
	}

	return opts, nil
}
