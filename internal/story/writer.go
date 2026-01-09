package story

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WriteStoriesToDir writes stories to individual JSON files in the specified directory
// Uses atomic file writes (temp file + rename) to prevent race conditions
func WriteStoriesToDir(dir, messageID string, date time.Time, stories []Story) error {
	sanitized := sanitizeMessageID(messageID)
	dateStr := date.Format("2006-01-02")

	for i, story := range stories {
		filename := fmt.Sprintf("%s_%s_%d.json", dateStr, sanitized, i+1)
		path := filepath.Join(dir, filename)

		// Check if file already exists (skip if present from concurrent process)
		if _, err := os.Stat(path); err == nil {
			continue // File already exists, skip
		}

		data, err := json.MarshalIndent(story, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal story: %w", err)
		}

		// Atomic write: write to temp file, then rename
		// This prevents partial writes and race conditions
		tmpPath := path + ".tmp"

		// Use 0600 permissions (owner read/write only) for privacy
		// Newsletter content may contain sensitive information
		if err := os.WriteFile(tmpPath, data, 0600); err != nil {
			return fmt.Errorf("failed to write temp file: %w", err)
		}

		// Atomic rename - if this fails, another process won the race
		if err := os.Rename(tmpPath, path); err != nil {
			// Clean up temp file
			os.Remove(tmpPath)
			// Check if target file now exists (another process created it)
			if _, statErr := os.Stat(path); statErr == nil {
				continue // File exists now, another process won the race
			}
			return fmt.Errorf("failed to rename temp file: %w", err)
		}
	}

	return nil
}

// sanitizeMessageID removes angle brackets and replaces filesystem-unsafe characters
func sanitizeMessageID(messageID string) string {
	// Remove angle brackets
	s := strings.TrimPrefix(messageID, "<")
	s = strings.TrimSuffix(s, ">")

	// Replace filesystem-unsafe characters with underscore
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	return replacer.Replace(s)
}
