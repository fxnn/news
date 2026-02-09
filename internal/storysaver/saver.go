package storysaver

import (
	"fmt"
	"os"
	"path/filepath"
)

// ListSavedFilenames returns a set of JSON filenames present in the savedir.
// Used to determine which stories have been saved for later.
func ListSavedFilenames(savedir string) (map[string]bool, error) {
	if _, err := os.Stat(savedir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", savedir)
	}

	pattern := filepath.Join(savedir, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob saved files: %w", err)
	}

	filenames := make(map[string]bool, len(matches))
	for _, path := range matches {
		filenames[filepath.Base(path)] = true
	}

	return filenames, nil
}
