package storyreader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fxnn/news/internal/story"
)

// ReadStories reads all story JSON files from a directory
func ReadStories(dir string) ([]story.Story, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	pattern := filepath.Join(dir, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob story files: %w", err)
	}

	// Initialize with empty slice to ensure JSON encoding as [] not null
	stories := []story.Story{}

	for _, path := range matches {
		data, err := os.ReadFile(path) //nolint:gosec // G304: Paths from Glob pattern, constrained to storydir
		if err != nil {
			// Skip files we can't read
			continue
		}

		var s story.Story
		if err := json.Unmarshal(data, &s); err != nil {
			// Skip invalid JSON files
			continue
		}

		// Add filename for debugging
		s.Filename = filepath.Base(path)

		stories = append(stories, s)
	}

	// Sort stories by date (newest first)
	sort.Slice(stories, func(i, j int) bool {
		return stories[i].Date.After(stories[j].Date)
	})

	return stories, nil
}
