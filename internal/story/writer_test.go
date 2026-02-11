package story

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriter_WriteToDir(t *testing.T) {
	tmpDir := t.TempDir()

	stories := []Story{
		{
			Headline:  "Test Story",
			Teaser:    "Test teaser",
			URL:       "https://example.com",
			FromEmail: "test@example.com",
			FromName:  "Test User",
			Date:      time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		},
	}

	messageID := "<test123@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	err := WriteStoriesToDir(tmpDir, messageID, date, stories)
	if err != nil {
		t.Fatalf("WriteStoriesToDir() unexpected error: %v", err)
	}

	// Check file was created with correct name
	expectedFilename := "2006-01-02_test123@example.com_1.json"
	expectedPath := filepath.Join(tmpDir, expectedFilename)

	if _, statErr := os.Stat(expectedPath); os.IsNotExist(statErr) {
		t.Errorf("Expected file %s was not created", expectedPath)
	}

	// Read and verify content
	content, err := os.ReadFile(expectedPath) //nolint:gosec // G304: Reading test file we just created
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var readStory Story
	if err := json.Unmarshal(content, &readStory); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if readStory.Headline != "Test Story" {
		t.Errorf("Headline = %v, want Test Story", readStory.Headline)
	}
}

func TestWriter_WriteMultipleStories(t *testing.T) {
	tmpDir := t.TempDir()

	stories := []Story{
		{
			Headline:  "Story 1",
			Teaser:    "Teaser 1",
			URL:       "https://example.com/1",
			FromEmail: "test@example.com",
			FromName:  "Test",
			Date:      time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		},
		{
			Headline:  "Story 2",
			Teaser:    "Teaser 2",
			URL:       "https://example.com/2",
			FromEmail: "test@example.com",
			FromName:  "Test",
			Date:      time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		},
	}

	messageID := "<multi@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	err := WriteStoriesToDir(tmpDir, messageID, date, stories)
	if err != nil {
		t.Fatalf("WriteStoriesToDir() unexpected error: %v", err)
	}

	// Check both files were created
	file1 := filepath.Join(tmpDir, "2006-01-02_multi@example.com_1.json")
	file2 := filepath.Join(tmpDir, "2006-01-02_multi@example.com_2.json")

	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Errorf("Expected file %s was not created", file1)
	}

	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Errorf("Expected file %s was not created", file2)
	}
}

func TestWriter_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	stories := []Story{
		{
			Headline:  "Test Story",
			Teaser:    "Test teaser",
			URL:       "https://example.com",
			FromEmail: "test@example.com",
			FromName:  "Test User",
			Date:      time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		},
	}

	messageID := "<test@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	err := WriteStoriesToDir(tmpDir, messageID, date, stories)
	if err != nil {
		t.Fatalf("WriteStoriesToDir() unexpected error: %v", err)
	}

	// Check file permissions are 0600 (owner read/write only)
	expectedFilename := "2006-01-02_test@example.com_1.json"
	expectedPath := filepath.Join(tmpDir, expectedFilename)

	fileInfo, err := os.Stat(expectedPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	mode := fileInfo.Mode()
	expectedMode := os.FileMode(0600)

	if mode.Perm() != expectedMode {
		t.Errorf("File permissions = %04o, want %04o (owner read/write only for privacy)",
			mode.Perm(), expectedMode)
	}
}

func TestWriter_SanitizeMessageID(t *testing.T) {
	tests := []struct {
		name      string
		messageID string
		want      string
	}{
		{
			name:      "with brackets",
			messageID: "<test123@example.com>",
			want:      "test123@example.com",
		},
		{
			name:      "without brackets",
			messageID: "test123@example.com",
			want:      "test123@example.com",
		},
		{
			name:      "with special chars",
			messageID: "<test/123@example.com>",
			want:      "test_123@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeMessageID(tt.messageID)
			if got != tt.want {
				t.Errorf("sanitizeMessageID(%v) = %v, want %v", tt.messageID, got, tt.want)
			}
		})
	}
}
