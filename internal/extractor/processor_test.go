package extractor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/logger"
	"github.com/fxnn/news/internal/story"
)

func TestProcessor_Run_EmptyMaildir(t *testing.T) {
	tmpMaildir := t.TempDir()
	tmpStorydir := t.TempDir()

	// Create Maildir structure
	if err := os.MkdirAll(filepath.Join(tmpMaildir, "cur"), 0o750); err != nil {
		t.Fatalf("Failed to create cur directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpMaildir, "new"), 0o750); err != nil {
		t.Fatalf("Failed to create new directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpMaildir, "tmp"), 0o750); err != nil {
		t.Fatalf("Failed to create tmp directory: %v", err)
	}

	cfg := &config.StoryExtractor{
		Maildir:  tmpMaildir,
		Storydir: tmpStorydir,
		Verbose:  false,
	}

	log := logger.New(false)
	extractor := &story.StubExtractor{
		Stories: []story.ExtractedStory{
			{Headline: "Test", Teaser: "Test", URL: "https://example.com"},
		},
	}

	processor := NewProcessor(cfg, log, extractor)
	result, err := processor.Run()

	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("Total = %d, want 0", result.Total)
	}

	if result.Processed != 0 {
		t.Errorf("Processed = %d, want 0", result.Processed)
	}
}

func TestProcessor_Run_SingleEmail(t *testing.T) {
	tmpMaildir := t.TempDir()
	tmpStorydir := t.TempDir()

	// Create Maildir structure
	curDir := filepath.Join(tmpMaildir, "cur")
	if err := os.MkdirAll(curDir, 0o750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create a test email
	emailContent := `From: Test User <test@example.com>
To: user@example.com
Subject: Test Newsletter
Date: Mon, 02 Jan 2006 15:04:05 -0700
Message-ID: <test123@example.com>

This is a test email body.
`
	emailPath := filepath.Join(curDir, "test.eml")
	if err := os.WriteFile(emailPath, []byte(emailContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.StoryExtractor{
		Maildir:  tmpMaildir,
		Storydir: tmpStorydir,
		Verbose:  false,
	}

	log := logger.New(false)
	extractor := &story.StubExtractor{
		Stories: []story.ExtractedStory{
			{Headline: "Test Story", Teaser: "A test story", URL: "https://example.com/test"},
		},
	}

	processor := NewProcessor(cfg, log, extractor)
	result, err := processor.Run()

	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}

	if result.Processed != 1 {
		t.Errorf("Processed = %d, want 1", result.Processed)
	}

	if result.Errors != 0 {
		t.Errorf("Errors = %d, want 0", result.Errors)
	}

	// Verify story was written
	matches, err := filepath.Glob(filepath.Join(tmpStorydir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 story file, got %d", len(matches))
	}
}

func TestProcessor_Run_SkipsExistingStories(t *testing.T) {
	tmpMaildir := t.TempDir()
	tmpStorydir := t.TempDir()

	// Create Maildir structure
	curDir := filepath.Join(tmpMaildir, "cur")
	if err := os.MkdirAll(curDir, 0o750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create a test email
	emailContent := `From: Test User <test@example.com>
To: user@example.com
Subject: Test Newsletter
Date: Mon, 02 Jan 2006 15:04:05 -0700
Message-ID: <test123@example.com>

This is a test email body.
`
	emailPath := filepath.Join(curDir, "test.eml")
	if err := os.WriteFile(emailPath, []byte(emailContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.StoryExtractor{
		Maildir:  tmpMaildir,
		Storydir: tmpStorydir,
		Verbose:  false,
	}

	log := logger.New(false)
	extractor := &story.StubExtractor{
		Stories: []story.ExtractedStory{
			{Headline: "Test Story", Teaser: "A test story", URL: "https://example.com/test"},
		},
	}

	// First run - should process the email
	processor := NewProcessor(cfg, log, extractor)
	result1, err := processor.Run()
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if result1.Processed != 1 {
		t.Errorf("First run: Processed = %d, want 1", result1.Processed)
	}

	// Second run - should skip the email
	result2, err := processor.Run()
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if result2.Skipped != 1 {
		t.Errorf("Second run: Skipped = %d, want 1", result2.Skipped)
	}

	if result2.Processed != 0 {
		t.Errorf("Second run: Processed = %d, want 0", result2.Processed)
	}
}

func TestProcessor_Run_WithLimit(t *testing.T) {
	tmpMaildir := t.TempDir()
	tmpStorydir := t.TempDir()

	// Create Maildir structure
	curDir := filepath.Join(tmpMaildir, "cur")
	if err := os.MkdirAll(curDir, 0o750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create multiple test emails
	for i := 1; i <= 5; i++ {
		emailContent := `From: Test User <test@example.com>
To: user@example.com
Subject: Test Newsletter
Date: Mon, 02 Jan 2006 15:04:05 -0700
Message-ID: <test` + string(rune('0'+i)) + `@example.com>

This is a test email body.
`
		emailPath := filepath.Join(curDir, "test"+string(rune('0'+i))+".eml")
		if err := os.WriteFile(emailPath, []byte(emailContent), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.StoryExtractor{
		Maildir:  tmpMaildir,
		Storydir: tmpStorydir,
		Limit:    2,
		Verbose:  false,
	}

	log := logger.New(false)
	extractor := &story.StubExtractor{
		Stories: []story.ExtractedStory{
			{Headline: "Test Story", Teaser: "A test story", URL: "https://example.com/test"},
		},
	}

	processor := NewProcessor(cfg, log, extractor)
	result, err := processor.Run()

	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("Total = %d, want 2 (limit applied)", result.Total)
	}

	if result.Processed != 2 {
		t.Errorf("Processed = %d, want 2", result.Processed)
	}
}

func TestProcessor_Run_InvalidEmail(t *testing.T) {
	tmpMaildir := t.TempDir()
	tmpStorydir := t.TempDir()

	// Create Maildir structure
	curDir := filepath.Join(tmpMaildir, "cur")
	if err := os.MkdirAll(curDir, 0o750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create an invalid email file
	emailPath := filepath.Join(curDir, "invalid.eml")
	if err := os.WriteFile(emailPath, []byte("not a valid email"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.StoryExtractor{
		Maildir:  tmpMaildir,
		Storydir: tmpStorydir,
		Verbose:  false,
	}

	log := logger.New(false)
	extractor := &story.StubExtractor{
		Stories: []story.ExtractedStory{
			{Headline: "Test Story", Teaser: "A test story", URL: "https://example.com/test"},
		},
	}

	processor := NewProcessor(cfg, log, extractor)
	result, err := processor.Run()

	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}

	if result.Errors != 1 {
		t.Errorf("Errors = %d, want 1", result.Errors)
	}

	if result.Processed != 0 {
		t.Errorf("Processed = %d, want 0", result.Processed)
	}
}

func TestProcessor_Run_MultipleEmailsMixedResults(t *testing.T) {
	tmpMaildir := t.TempDir()
	tmpStorydir := t.TempDir()

	// Create Maildir structure
	curDir := filepath.Join(tmpMaildir, "cur")
	if err := os.MkdirAll(curDir, 0o750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create a valid email
	validEmail := `From: Test User <test@example.com>
To: user@example.com
Subject: Valid Newsletter
Date: Mon, 02 Jan 2006 15:04:05 -0700
Message-ID: <valid@example.com>

This is a valid email.
`
	validPath := filepath.Join(curDir, "valid.eml")
	if err := os.WriteFile(validPath, []byte(validEmail), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an invalid email
	invalidPath := filepath.Join(curDir, "invalid.eml")
	if err := os.WriteFile(invalidPath, []byte("invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create another valid email and pre-process it (for skip test)
	skipEmail := `From: Skip User <skip@example.com>
To: user@example.com
Subject: Skip Newsletter
Date: Tue, 03 Jan 2006 15:04:05 -0700
Message-ID: <skip@example.com>

This should be skipped.
`
	skipPath := filepath.Join(curDir, "skip.eml")
	if err := os.WriteFile(skipPath, []byte(skipEmail), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre-create story for the skip email
	date := time.Date(2006, 1, 3, 15, 4, 5, 0, time.UTC)
	stories := []story.Story{
		{Headline: "Skip", Teaser: "Skip", URL: "https://example.com", Date: date},
	}
	if err := story.WriteStoriesToDir(tmpStorydir, "<skip@example.com>", date, stories); err != nil {
		t.Fatal(err)
	}

	cfg := &config.StoryExtractor{
		Maildir:  tmpMaildir,
		Storydir: tmpStorydir,
		Verbose:  false,
	}

	log := logger.New(false)
	extractor := &story.StubExtractor{
		Stories: []story.ExtractedStory{
			{Headline: "Test Story", Teaser: "A test story", URL: "https://example.com/test"},
		},
	}

	processor := NewProcessor(cfg, log, extractor)
	result, err := processor.Run()

	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Total = %d, want 3", result.Total)
	}

	if result.Processed != 1 {
		t.Errorf("Processed = %d, want 1", result.Processed)
	}

	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}

	if result.Errors != 1 {
		t.Errorf("Errors = %d, want 1", result.Errors)
	}
}
