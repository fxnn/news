package main

import (
	"fmt"
	"os"

	"github.com/fxnn/news/internal/cli"
	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/email"
	"github.com/fxnn/news/internal/logger"
	"github.com/fxnn/news/internal/maildir"
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
		"debug", opts.Debug)
	log.Debug("LLM configuration loaded",
		"provider", cfg.LLM.Provider,
		"model", cfg.LLM.Model)

	// Read all email files from the Maildir
	emailPaths, err := maildir.Read(opts.Maildir)
	if err != nil {
		log.Error("failed to read maildir", "error", err)
		os.Exit(1)
	}

	log.Info("found emails", "count", len(emailPaths))

	// Apply limit if specified
	if opts.Limit > 0 && len(emailPaths) > opts.Limit {
		emailPaths = emailPaths[:opts.Limit]
		log.Info("limiting email processing", "limit", opts.Limit)
	}

	// Process each email
	processedCount := 0
	errorCount := 0

	for i, path := range emailPaths {
		log.Debug("processing email", "index", i+1, "path", path)

		file, err := os.Open(path)
		if err != nil {
			log.Warn("failed to open email file", "path", path, "error", err)
			errorCount++
			continue
		}

		parsedEmail, err := email.Parse(file)
		file.Close()

		if err != nil {
			log.Warn("failed to parse email", "path", path, "error", err)
			errorCount++
			continue
		}

		processedCount++

		if opts.Debug {
			log.Info("parsed email",
				"index", i+1,
				"subject", parsedEmail.Subject,
				"from_email", parsedEmail.FromEmail,
				"from_name", parsedEmail.FromName,
				"date", parsedEmail.Date.Format("2006-01-02 15:04:05"),
				"message_id", parsedEmail.MessageID,
				"body_length", len(parsedEmail.Body))
		}

		// TODO: Extract stories using LLM
		// TODO: Save stories to files or stdout
	}

	log.Info("processing complete",
		"total", len(emailPaths),
		"processed", processedCount,
		"errors", errorCount)
}
