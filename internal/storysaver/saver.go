package storysaver

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

var ErrAlreadySaved = errors.New("story is already saved")

// Save copies a story JSON file from storydir to savedir.
// Creates savedir if it does not exist. Uses atomic writes to prevent partial copies.
func Save(storydir, savedir, filename string) error {
	if err := validateFilename(filename); err != nil {
		return err
	}

	if err := os.MkdirAll(savedir, 0700); err != nil {
		return fmt.Errorf("failed to create savedir: %w", err)
	}

	destPath := filepath.Join(savedir, filename)
	if _, err := os.Stat(destPath); err == nil {
		return ErrAlreadySaved
	}

	srcPath := filepath.Join(storydir, filename)
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read story file: %w", err)
	}

	tmpPath := destPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func validateFilename(filename string) error {
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") ||
		strings.Contains(filename, "..") || !strings.HasSuffix(filename, ".json") {
		return fmt.Errorf("invalid filename: %s", filename)
	}
	return nil
}
