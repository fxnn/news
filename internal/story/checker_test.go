package story

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoriesExist_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	messageID := "<test123@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	exists, err := StoriesExist(tmpDir, messageID, date)
	if err != nil {
		t.Fatalf("StoriesExist() unexpected error: %v", err)
	}

	if exists {
		t.Error("StoriesExist() = true, want false (no files)")
	}
}

func TestStoriesExist_WithMatchingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a matching story file
	filename := "2006-01-02_test123@example.com_1.json"
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	messageID := "<test123@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	exists, err := StoriesExist(tmpDir, messageID, date)
	if err != nil {
		t.Fatalf("StoriesExist() unexpected error: %v", err)
	}

	if !exists {
		t.Error("StoriesExist() = false, want true (matching file exists)")
	}
}

func TestStoriesExist_WithMultipleMatchingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple matching story files
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(tmpDir, "2006-01-02_test@example.com_"+string(rune('0'+i))+".json")
		if err := os.WriteFile(filename, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	messageID := "<test@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	exists, err := StoriesExist(tmpDir, messageID, date)
	if err != nil {
		t.Fatalf("StoriesExist() unexpected error: %v", err)
	}

	if !exists {
		t.Error("StoriesExist() = false, want true (multiple matching files exist)")
	}
}

func TestStoriesExist_DifferentMessageID(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with different message ID
	filename := "2006-01-02_other@example.com_1.json"
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	messageID := "<test@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	exists, err := StoriesExist(tmpDir, messageID, date)
	if err != nil {
		t.Fatalf("StoriesExist() unexpected error: %v", err)
	}

	if exists {
		t.Error("StoriesExist() = true, want false (different message ID)")
	}
}

func TestStoriesExist_DifferentDate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with different date
	filename := "2006-01-03_test@example.com_1.json"
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	messageID := "<test@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	exists, err := StoriesExist(tmpDir, messageID, date)
	if err != nil {
		t.Fatalf("StoriesExist() unexpected error: %v", err)
	}

	if exists {
		t.Error("StoriesExist() = true, want false (different date)")
	}
}

func TestStoriesExist_NonExistentDir(t *testing.T) {
	messageID := "<test@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	exists, err := StoriesExist("/nonexistent/directory", messageID, date)
	if err != nil {
		t.Fatalf("StoriesExist() unexpected error: %v", err)
	}

	if exists {
		t.Error("StoriesExist() = true, want false (directory doesn't exist)")
	}
}
