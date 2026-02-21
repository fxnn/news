package maildir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRead_BasicMaildir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create basic Maildir structure
	curDir := filepath.Join(tmpDir, "cur")
	newDir := filepath.Join(tmpDir, "new")

	if err := os.MkdirAll(curDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(newDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// Create test email files
	email1 := filepath.Join(curDir, "1234567890.M123456P12345.host:2,S")
	email2 := filepath.Join(newDir, "1234567891.M123457P12346.host")

	if err := os.WriteFile(email1, []byte("test email 1"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(email2, []byte("test email 2"), 0o600); err != nil {
		t.Fatal(err)
	}

	paths, err := Read(tmpDir)
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("Read() returned %d paths, want 2", len(paths))
	}
}

func TestRead_EmptyMaildir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty Maildir structure
	if err := os.MkdirAll(filepath.Join(tmpDir, "cur"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "new"), 0o750); err != nil {
		t.Fatal(err)
	}

	paths, err := Read(tmpDir)
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}

	if len(paths) != 0 {
		t.Errorf("Read() returned %d paths, want 0", len(paths))
	}
}

func TestRead_NonExistentDir(t *testing.T) {
	_, err := Read("/nonexistent/maildir")
	if err == nil {
		t.Error("Read() expected error for nonexistent directory, got nil")
	}
}

func TestRead_NestedMaildir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested Maildir structure (.subfolder/cur, .subfolder/new)
	mainCur := filepath.Join(tmpDir, "cur")
	subfolderCur := filepath.Join(tmpDir, ".subfolder", "cur")
	subfolderNew := filepath.Join(tmpDir, ".subfolder", "new")

	if err := os.MkdirAll(mainCur, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subfolderCur, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subfolderNew, 0o750); err != nil {
		t.Fatal(err)
	}

	// Create emails in different locations
	email1 := filepath.Join(mainCur, "1234567890.M123456P12345.host:2,S")
	email2 := filepath.Join(subfolderCur, "1234567891.M123457P12346.host:2,S")
	email3 := filepath.Join(subfolderNew, "1234567892.M123458P12347.host")

	if err := os.WriteFile(email1, []byte("email 1"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(email2, []byte("email 2"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(email3, []byte("email 3"), 0o600); err != nil {
		t.Fatal(err)
	}

	paths, err := Read(tmpDir)
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}

	if len(paths) != 3 {
		t.Errorf("Read() returned %d paths, want 3", len(paths))
	}
}

func TestRead_IgnoresTmpDir(t *testing.T) {
	tmpDir := t.TempDir()

	curDir := filepath.Join(tmpDir, "cur")
	tmpMailDir := filepath.Join(tmpDir, "tmp")

	if err := os.MkdirAll(curDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tmpMailDir, 0o750); err != nil {
		t.Fatal(err)
	}

	email1 := filepath.Join(curDir, "1234567890.M123456P12345.host:2,S")
	emailTmp := filepath.Join(tmpMailDir, "1234567891.M123457P12346.host")

	if err := os.WriteFile(email1, []byte("email 1"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(emailTmp, []byte("temp email"), 0o600); err != nil {
		t.Fatal(err)
	}

	paths, err := Read(tmpDir)
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}

	if len(paths) != 1 {
		t.Errorf("Read() returned %d paths, want 1 (tmp dir should be ignored)", len(paths))
	}
}

func TestRead_ReturnsNewestFirst(t *testing.T) {
	tmpDir := t.TempDir()

	curDir := filepath.Join(tmpDir, "cur")
	if err := os.MkdirAll(curDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// Maildir filenames start with Unix timestamps by convention
	oldFile := filepath.Join(curDir, "1000000000.M100.host:2,S")
	midFile := filepath.Join(curDir, "1500000000.M200.host:2,S")
	newFile := filepath.Join(curDir, "2000000000.M300.host:2,S")

	for _, f := range []string{oldFile, midFile, newFile} {
		if err := os.WriteFile(f, []byte("email"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	paths, err := Read(tmpDir)
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}

	if len(paths) != 3 {
		t.Fatalf("Read() returned %d paths, want 3", len(paths))
	}

	if filepath.Base(paths[0]) != "2000000000.M300.host:2,S" {
		t.Errorf("paths[0] = %s, want newest file first", filepath.Base(paths[0]))
	}
	if filepath.Base(paths[2]) != "1000000000.M100.host:2,S" {
		t.Errorf("paths[2] = %s, want oldest file last", filepath.Base(paths[2]))
	}
}
