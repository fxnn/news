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
func WriteStoriesToDir(dir, messageID string, date time.Time, stories []Story) error {
	sanitized := sanitizeMessageID(messageID)
	dateStr := date.Format("2006-01-02")

	for i, story := range stories {
		filename := fmt.Sprintf("%s_%s_%d.json", dateStr, sanitized, i+1)
		path := filepath.Join(dir, filename)

		data, err := json.MarshalIndent(story, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal story: %w", err)
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("failed to write story file: %w", err)
		}
	}

	return nil
}

// WriteStoriesToStdout writes stories as JSON to stdout
func WriteStoriesToStdout(stories []Story) error {
	for _, story := range stories {
		data, err := json.Marshal(story)
		if err != nil {
			return fmt.Errorf("failed to marshal story: %w", err)
		}
		fmt.Println(string(data))
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
