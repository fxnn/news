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
	if _, err := os.Stat(savedir); errors.Is(err, os.ErrNotExist) {
		return map[string]bool{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to access savedir: %w", err)
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
var ErrInvalidFilename = errors.New("invalid filename")

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
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check destination: %w", err)
	}

	srcPath := filepath.Join(storydir, filename)
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read story file: %w", err)
	}

	tmpFile, err := os.CreateTemp(savedir, ".save-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	n, err := tmpFile.Write(data)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if n != len(data) {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("short write: wrote %d of %d bytes", n, len(data))
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// Unsave removes a saved story from savedir.
func Unsave(savedir, filename string) error {
	if err := validateFilename(filename); err != nil {
		return err
	}

	path := filepath.Join(savedir, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove saved story: %w", err)
	}

	return nil
}

func validateFilename(filename string) error {
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") ||
		strings.Contains(filename, "..") || !strings.HasSuffix(filename, ".json") {
		return fmt.Errorf("%w: %s", ErrInvalidFilename, filename)
	}
	return nil
}
