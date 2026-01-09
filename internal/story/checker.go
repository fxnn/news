package story

import (
	"fmt"
	"path/filepath"
	"time"
)

// StoriesExist checks if any story files already exist for the given email
func StoriesExist(dir, messageID string, date time.Time) (bool, error) {
	sanitized := sanitizeMessageID(messageID)
	dateStr := date.Format("2006-01-02")

	// Build glob pattern: <date>_<message-id>_*.json
	pattern := filepath.Join(dir, fmt.Sprintf("%s_%s_*.json", dateStr, sanitized))

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false, fmt.Errorf("failed to check for existing stories: %w", err)
	}

	return len(matches) > 0, nil
}
