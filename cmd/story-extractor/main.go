package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/logger"
)

func main() {
	maildir := flag.String("maildir", "", "Path to the Maildir directory (required)")
	storydir := flag.String("storydir", "", "Output directory for story files (optional, defaults to stdout)")
	configPath := flag.String("config", "", "Path to the TOML configuration file (required)")
	limit := flag.Int("limit", 0, "Maximum number of emails to process (optional, 0 = unlimited)")
	verbose := flag.Bool("verbose", false, "Enable detailed log output")

	flag.Parse()

	if *maildir == "" {
		fmt.Fprintln(os.Stderr, "Error: --maildir is required")
		flag.Usage()
		os.Exit(1)
	}

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --config is required")
		flag.Usage()
		os.Exit(1)
	}

	log := logger.New(*verbose)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log.Info("starting story extractor",
		"maildir", *maildir,
		"storydir", *storydir,
		"config", *configPath,
		"limit", *limit)
	log.Debug("LLM configuration loaded",
		"provider", cfg.LLM.Provider,
		"model", cfg.LLM.Model)
}
