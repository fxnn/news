package maildir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Read recursively reads all email files from a Maildir directory.
// It looks for files in 'cur' and 'new' subdirectories and ignores 'tmp'.
func Read(maildirPath string) ([]string, error) {
	var emails []string

	err := filepath.WalkDir(maildirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip tmp directories
		if d.IsDir() && d.Name() == "tmp" {
			return filepath.SkipDir
		}

		// Only process regular files
		if d.IsDir() {
			return nil
		}

		// Check if file is in a 'cur' or 'new' directory
		dir := filepath.Dir(path)
		if strings.HasSuffix(dir, "/cur") || strings.HasSuffix(dir, "/new") ||
			strings.HasSuffix(dir, string(filepath.Separator)+"cur") ||
			strings.HasSuffix(dir, string(filepath.Separator)+"new") {
			emails = append(emails, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to read maildir: %w", err)
	}

	return emails, nil
}
