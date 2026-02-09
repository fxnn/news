package storysaver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListSavedFilenames_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	filenames, err := ListSavedFilenames(tmpDir)
	if err != nil {
		t.Fatalf("ListSavedFilenames() unexpected error: %v", err)
	}

	if len(filenames) != 0 {
		t.Errorf("ListSavedFilenames() returned %d filenames, want 0", len(filenames))
	}
}

func TestListSavedFilenames_WithFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some JSON files
	for _, name := range []string{"story1.json", "story2.json"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("{}"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	// Create a non-JSON file that should be ignored
	if err := os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte("hi"), 0600); err != nil {
		t.Fatal(err)
	}

	filenames, err := ListSavedFilenames(tmpDir)
	if err != nil {
		t.Fatalf("ListSavedFilenames() unexpected error: %v", err)
	}

	if len(filenames) != 2 {
		t.Fatalf("ListSavedFilenames() returned %d filenames, want 2", len(filenames))
	}
	if !filenames["story1.json"] {
		t.Error("ListSavedFilenames() missing story1.json")
	}
	if !filenames["story2.json"] {
		t.Error("ListSavedFilenames() missing story2.json")
	}
}

func TestListSavedFilenames_NonExistentDir(t *testing.T) {
	_, err := ListSavedFilenames("/nonexistent/directory")
	if err == nil {
		t.Error("ListSavedFilenames() expected error for nonexistent directory, got nil")
	}
}
