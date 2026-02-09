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
	nonexistentDir := filepath.Join(t.TempDir(), "does-not-exist")
	filenames, err := ListSavedFilenames(nonexistentDir)
	if err != nil {
		t.Fatalf("ListSavedFilenames() unexpected error: %v", err)
	}
	if len(filenames) != 0 {
		t.Errorf("ListSavedFilenames() returned %d filenames, want 0", len(filenames))
	}
}

func TestSave_CopiesFile(t *testing.T) {
	storydir := t.TempDir()
	savedir := t.TempDir()

	content := []byte(`{"headline":"Test"}`)
	if err := os.WriteFile(filepath.Join(storydir, "story.json"), content, 0600); err != nil {
		t.Fatal(err)
	}

	if err := Save(storydir, savedir, "story.json"); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	saved, err := os.ReadFile(filepath.Join(savedir, "story.json"))
	if err != nil {
		t.Fatalf("saved file not found: %v", err)
	}
	if string(saved) != string(content) {
		t.Errorf("saved content = %q, want %q", saved, content)
	}
}

func TestSave_FileNotFound(t *testing.T) {
	storydir := t.TempDir()
	savedir := t.TempDir()

	err := Save(storydir, savedir, "nonexistent.json")
	if err == nil {
		t.Error("Save() expected error for nonexistent file, got nil")
	}
}

func TestSave_AlreadyExists(t *testing.T) {
	storydir := t.TempDir()
	savedir := t.TempDir()

	content := []byte(`{"headline":"Test"}`)
	if err := os.WriteFile(filepath.Join(storydir, "story.json"), content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(savedir, "story.json"), content, 0600); err != nil {
		t.Fatal(err)
	}

	err := Save(storydir, savedir, "story.json")
	if err == nil {
		t.Error("Save() expected error for already saved file, got nil")
	}
}

func TestSave_RejectsPathTraversal(t *testing.T) {
	storydir := t.TempDir()
	savedir := t.TempDir()

	for _, filename := range []string{"../etc/passwd", "foo/bar.json", "..\\evil.json"} {
		err := Save(storydir, savedir, filename)
		if err == nil {
			t.Errorf("Save(%q) expected error for path traversal, got nil", filename)
		}
	}
}

func TestSave_LeavesNoTempFiles(t *testing.T) {
	storydir := t.TempDir()
	savedir := t.TempDir()

	content := []byte(`{"headline":"Test"}`)
	if err := os.WriteFile(filepath.Join(storydir, "story.json"), content, 0600); err != nil {
		t.Fatal(err)
	}

	if err := Save(storydir, savedir, "story.json"); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	entries, err := os.ReadDir(savedir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.Name() != "story.json" {
			t.Errorf("unexpected file in savedir: %s", entry.Name())
		}
	}
}

func TestSave_CreatesSavedirIfNotExists(t *testing.T) {
	storydir := t.TempDir()
	savedir := filepath.Join(t.TempDir(), "new-subdir")

	content := []byte(`{"headline":"Test"}`)
	if err := os.WriteFile(filepath.Join(storydir, "story.json"), content, 0600); err != nil {
		t.Fatal(err)
	}

	if err := Save(storydir, savedir, "story.json"); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(savedir, "story.json")); err != nil {
		t.Errorf("saved file does not exist: %v", err)
	}
}

func TestUnsave_RemovesFile(t *testing.T) {
	savedir := t.TempDir()

	if err := os.WriteFile(filepath.Join(savedir, "story.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := Unsave(savedir, "story.json"); err != nil {
		t.Fatalf("Unsave() unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(savedir, "story.json")); !os.IsNotExist(err) {
		t.Error("Unsave() file still exists after removal")
	}
}

func TestUnsave_FileNotFound(t *testing.T) {
	savedir := t.TempDir()

	err := Unsave(savedir, "nonexistent.json")
	if err == nil {
		t.Error("Unsave() expected error for nonexistent file, got nil")
	}
}

func TestUnsave_RejectsPathTraversal(t *testing.T) {
	savedir := t.TempDir()

	for _, filename := range []string{"../etc/passwd", "foo/bar.json", "..\\evil.json"} {
		err := Unsave(savedir, filename)
		if err == nil {
			t.Errorf("Unsave(%q) expected error for path traversal, got nil", filename)
		}
	}
}
