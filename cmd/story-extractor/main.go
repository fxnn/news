package main

import (
	"fmt"
	"os"

	"github.com/fxnn/news/internal/cli"
	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/extractor"
	"github.com/fxnn/news/internal/llm"
	"github.com/fxnn/news/internal/logger"
)

func main() {
	opts, err := cli.ParseOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(opts.Verbose)

	cfg, err := config.Load(opts.Config)
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log.Info("starting story extractor",
		"maildir", opts.Maildir,
		"storydir", opts.Storydir,
		"config", opts.Config,
		"limit", opts.Limit,
		"log_headers", opts.LogHeaders,
		"log_bodies", opts.LogBodies,
		"log_stories", opts.LogStories)
	log.Debug("LLM configuration loaded",
		"provider", cfg.LLM.Provider,
		"model", cfg.LLM.Model)

	// Create story extractor
	storyExtractor := llm.NewOpenAIExtractor(&cfg.LLM)

	// Create and run processor
	processor := extractor.NewProcessor(opts, log, storyExtractor)
	_, err = processor.Run()
	if err != nil {
		log.Error("processing failed", "error", err)
		os.Exit(1)
	}
}
