package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fxnn/news/internal/cli"
	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/email"
	"github.com/fxnn/news/internal/llm"
	"github.com/fxnn/news/internal/logger"
	"github.com/fxnn/news/internal/maildir"
	"github.com/fxnn/news/internal/story"
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
	extractor := llm.NewOpenAIExtractor(&cfg.LLM)

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
	skippedCount := 0

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

		// Check if stories already exist (incremental processing)
		exists, err := story.StoriesExist(opts.Storydir, parsedEmail.MessageID, parsedEmail.Date)
		if err != nil {
			log.Warn("failed to check for existing stories", "path", path, "error", err)
		} else if exists {
			log.Debug("skipping email (stories already exist)", "path", path, "message_id", parsedEmail.MessageID)
			skippedCount++
			continue
		}

		processedCount++

		if opts.LogHeaders || opts.LogBodies {
			logArgs := []any{
				"index", i + 1,
				"subject", parsedEmail.Subject,
				"from_email", parsedEmail.FromEmail,
				"from_name", parsedEmail.FromName,
				"date", parsedEmail.Date.Format("2006-01-02 15:04:05"),
				"message_id", parsedEmail.MessageID,
				"body_length", len(parsedEmail.Body),
			}

			if opts.LogBodies {
				logArgs = append(logArgs, "body", parsedEmail.Body)
			}

			log.Debug("parsed email", logArgs...)
		}

		// Extract stories using LLM
		startTime := time.Now()
		stories, err := extractor.Extract(parsedEmail)
		duration := time.Since(startTime)

		if err != nil {
			log.Warn("failed to extract stories", "path", path, "error", err)
			errorCount++
			continue
		}

		log.Info("extracted stories", "path", path, "count", len(stories), "duration_ms", duration.Milliseconds())

		// Log stories if requested
		if opts.LogStories {
			for i, s := range stories {
				log.Debug("story",
					"index", i+1,
					"headline", s.Headline,
					"teaser", s.Teaser,
					"url", s.URL)
			}
		}

		// Save stories to directory
		err = story.WriteStoriesToDir(opts.Storydir, parsedEmail.MessageID, parsedEmail.Date, stories)
		if err != nil {
			log.Warn("failed to write stories to directory", "path", path, "error", err)
			errorCount++
			continue
		}
	}

	log.Info("processing complete",
		"total", len(emailPaths),
		"processed", processedCount,
		"skipped", skippedCount,
		"errors", errorCount)
}
