package extractor

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/email"
	"github.com/fxnn/news/internal/maildir"
	"github.com/fxnn/news/internal/story"
)

// Processor orchestrates the story extraction workflow
type Processor struct {
	cfg       *config.StoryExtractor
	log       *slog.Logger
	extractor story.Extractor
}

// Result holds the processing results
type Result struct {
	Total     int
	Processed int
	Skipped   int
	Errors    int
}

// NewProcessor creates a new story extraction processor
func NewProcessor(cfg *config.StoryExtractor, log *slog.Logger, extractor story.Extractor) *Processor {
	return &Processor{
		cfg:       cfg,
		log:       log,
		extractor: extractor,
	}
}

// Run executes the story extraction workflow
func (p *Processor) Run() (*Result, error) {
	// Read all email files from the Maildir
	emailPaths, err := maildir.Read(p.cfg.Maildir)
	if err != nil {
		return nil, fmt.Errorf("failed to read maildir: %w", err)
	}

	p.log.Info("found emails", "count", len(emailPaths))

	// Apply limit if specified
	if p.cfg.Limit > 0 && len(emailPaths) > p.cfg.Limit {
		emailPaths = emailPaths[:p.cfg.Limit]
		p.log.Info("limiting email processing", "limit", p.cfg.Limit)
	}

	result := &Result{
		Total: len(emailPaths),
	}

	// Process each email
	for i, path := range emailPaths {
		p.log.Debug("processing email", "index", i+1, "path", path)

		if err := p.processEmail(i, path); err != nil {
			if err == errSkipped {
				result.Skipped++
			} else {
				p.log.Warn("failed to process email", "path", path, "error", err)
				result.Errors++
			}
			continue
		}

		result.Processed++
	}

	p.log.Info("processing complete",
		"total", result.Total,
		"processed", result.Processed,
		"skipped", result.Skipped,
		"errors", result.Errors)

	return result, nil
}

var errSkipped = fmt.Errorf("email skipped")

func (p *Processor) processEmail(index int, path string) error {
	// Open and parse email
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	parsedEmail, err := email.Parse(file)
	file.Close()

	if err != nil {
		return fmt.Errorf("failed to parse email: %w", err)
	}

	// Check if stories already exist (incremental processing)
	exists, err := story.StoriesExist(p.cfg.Storydir, parsedEmail.MessageID, parsedEmail.Date)
	if err != nil {
		p.log.Warn("failed to check for existing stories", "path", path, "error", err)
	} else if exists {
		p.log.Debug("skipping email (stories already exist)", "path", path, "message_id", parsedEmail.MessageID)
		return errSkipped
	}

	// Log email details if requested
	if p.cfg.LogHeaders || p.cfg.LogBodies {
		logArgs := []any{
			"index", index + 1,
			"subject", parsedEmail.Subject,
			"from_email", parsedEmail.FromEmail,
			"from_name", parsedEmail.FromName,
			"date", parsedEmail.Date.Format("2006-01-02 15:04:05"),
			"message_id", parsedEmail.MessageID,
			"body_length", len(parsedEmail.Body),
		}

		if p.cfg.LogBodies {
			logArgs = append(logArgs, "body", parsedEmail.Body)
		}

		p.log.Debug("parsed email", logArgs...)
	}

	// Extract stories using LLM
	startTime := time.Now()
	stories, err := p.extractor.Extract(parsedEmail)
	duration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("failed to extract stories: %w", err)
	}

	p.log.Info("extracted stories", "path", path, "count", len(stories), "duration_ms", duration.Milliseconds())

	// Log stories if requested
	if p.cfg.LogStories {
		for i, s := range stories {
			p.log.Debug("story",
				"index", i+1,
				"headline", s.Headline,
				"teaser", s.Teaser,
				"url", s.URL)
		}
	}

	// Save stories to directory
	err = story.WriteStoriesToDir(p.cfg.Storydir, parsedEmail.MessageID, parsedEmail.Date, stories)
	if err != nil {
		return fmt.Errorf("failed to write stories: %w", err)
	}

	return nil
}
