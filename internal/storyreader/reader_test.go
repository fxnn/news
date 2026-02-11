package storyreader

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fxnn/news/internal/story"
)

func TestReadStories_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	stories, err := ReadStories(tmpDir)
	if err != nil {
		t.Fatalf("ReadStories() unexpected error: %v", err)
	}

	if len(stories) != 0 {
		t.Errorf("ReadStories() returned %d stories, want 0", len(stories))
	}
}

func TestReadStories_SingleStory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test story file
	testStory := story.Story{
		Headline:  "Test Headline",
		Teaser:    "Test teaser",
		URL:       "https://example.com/test",
		FromEmail: "test@example.com",
		FromName:  "Test User",
		Date:      time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
	}

	stories := []story.Story{testStory}
	messageID := "<test@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	if err := story.WriteStoriesToDir(tmpDir, messageID, date, stories); err != nil {
		t.Fatal(err)
	}

	// Read stories
	readStories, err := ReadStories(tmpDir)
	if err != nil {
		t.Fatalf("ReadStories() unexpected error: %v", err)
	}

	if len(readStories) != 1 {
		t.Fatalf("ReadStories() returned %d stories, want 1", len(readStories))
	}

	if readStories[0].Headline != "Test Headline" {
		t.Errorf("Headline = %v, want Test Headline", readStories[0].Headline)
	}
}

func TestReadStories_MultipleStories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test story files
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	stories1 := []story.Story{
		{
			Headline:  "Story 1",
			Teaser:    "Teaser 1",
			URL:       "https://example.com/1",
			FromEmail: "test@example.com",
			FromName:  "Test",
			Date:      date,
		},
		{
			Headline:  "Story 2",
			Teaser:    "Teaser 2",
			URL:       "https://example.com/2",
			FromEmail: "test@example.com",
			FromName:  "Test",
			Date:      date,
		},
	}

	if err := story.WriteStoriesToDir(tmpDir, "<test1@example.com>", date, stories1); err != nil {
		t.Fatal(err)
	}

	stories2 := []story.Story{
		{
			Headline:  "Story 3",
			Teaser:    "Teaser 3",
			URL:       "https://example.com/3",
			FromEmail: "test2@example.com",
			FromName:  "Test2",
			Date:      date,
		},
	}

	if err := story.WriteStoriesToDir(tmpDir, "<test2@example.com>", date, stories2); err != nil {
		t.Fatal(err)
	}

	// Read stories
	readStories, err := ReadStories(tmpDir)
	if err != nil {
		t.Fatalf("ReadStories() unexpected error: %v", err)
	}

	if len(readStories) != 3 {
		t.Fatalf("ReadStories() returned %d stories, want 3", len(readStories))
	}
}

func TestReadStories_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid JSON file
	invalidFile := filepath.Join(tmpDir, "2006-01-02_test@example.com_1.json")
	if err := os.WriteFile(invalidFile, []byte("invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Should skip invalid files and continue
	stories, err := ReadStories(tmpDir)
	if err != nil {
		t.Fatalf("ReadStories() unexpected error: %v", err)
	}

	if len(stories) != 0 {
		t.Errorf("ReadStories() returned %d stories, want 0 (invalid file skipped)", len(stories))
	}
}

func TestReadStories_NonExistentDir(t *testing.T) {
	_, err := ReadStories("/nonexistent/directory")
	if err == nil {
		t.Error("ReadStories() expected error for nonexistent directory, got nil")
	}
}
