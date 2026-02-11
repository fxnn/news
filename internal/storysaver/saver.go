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
	info, err := os.Stat(savedir)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]bool{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to access savedir: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("savedir is not a directory: %s", savedir)
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

// ErrAlreadySaved is returned when attempting to save a story that already exists.
var ErrAlreadySaved = errors.New("story is already saved")

// ErrInvalidFilename is returned when the filename contains path separators or is empty.
var ErrInvalidFilename = errors.New("invalid filename")

// Save copies a story JSON file from storydir to savedir.
// Creates savedir if it does not exist. Uses atomic writes to prevent partial copies.
func Save(storydir, savedir, filename string) error {
	if err := validateFilename(filename); err != nil {
		return err
	}

	if err := os.MkdirAll(savedir, 0o700); err != nil {
		return fmt.Errorf("failed to create savedir: %w", err)
	}

	destPath := filepath.Join(savedir, filename)
	if _, err := os.Stat(destPath); err == nil {
		return ErrAlreadySaved
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check destination: %w", err)
	}

	srcPath := filepath.Join(storydir, filename)
	data, err := os.ReadFile(srcPath) //nolint:gosec // G304: Filename validated above (no path separators)
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
		_ = tmpFile.Close() // Best effort cleanup
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if n != len(data) {
		_ = tmpFile.Close() // Best effort cleanup
		_ = os.Remove(tmpPath)
		return fmt.Errorf("short write: wrote %d of %d bytes", n, len(data))
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath) // Best effort cleanup
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		_ = os.Remove(tmpPath) // Best effort cleanup
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
	// Disallow directory traversal attempts in the name itself.
	if strings.Contains(filename, "..") {
		return fmt.Errorf("%w: %s", ErrInvalidFilename, filename)
	}

	// Only allow simple base filenames (no path separators or extra components).
	if filename != filepath.Base(filename) {
		return fmt.Errorf("%w: %s", ErrInvalidFilename, filename)
	}

	// Reject absolute paths on any OS.
	if filepath.IsAbs(filename) {
		return fmt.Errorf("%w: %s", ErrInvalidFilename, filename)
	}

	// Reject Windows volume-relative or absolute paths like "C:evil.json" or "C:\evil.json".
	if v := filepath.VolumeName(filename); v != "" {
		return fmt.Errorf("%w: %s", ErrInvalidFilename, filename)
	}

	// Require a .json extension.
	if filepath.Ext(filename) != ".json" {
		return fmt.Errorf("%w: %s", ErrInvalidFilename, filename)
	}

	return nil
}
