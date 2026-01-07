package cli

import (
	"flag"
	"fmt"
)

type Options struct {
	Maildir  string
	Storydir string
	Config   string
	Limit    int
	Verbose  bool
	Debug    bool
}

func ParseOptions(args []string) (*Options, error) {
	fs := flag.NewFlagSet("story-extractor", flag.ContinueOnError)

	opts := &Options{}

	fs.StringVar(&opts.Maildir, "maildir", "", "Path to the Maildir directory (required)")
	fs.StringVar(&opts.Storydir, "storydir", "", "Output directory for story files (optional, defaults to stdout)")
	fs.StringVar(&opts.Config, "config", "", "Path to the TOML configuration file (required)")
	fs.IntVar(&opts.Limit, "limit", 0, "Maximum number of emails to process (optional, 0 = unlimited)")
	fs.BoolVar(&opts.Verbose, "verbose", false, "Enable detailed log output")
	fs.BoolVar(&opts.Debug, "debug", false, "Enable debug mode (logs all found emails)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if opts.Maildir == "" {
		return nil, fmt.Errorf("--maildir is required")
	}

	if opts.Config == "" {
		return nil, fmt.Errorf("--config is required")
	}

	return opts, nil
}
