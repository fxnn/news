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

	if err := os.MkdirAll(curDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test email files
	email1 := filepath.Join(curDir, "1234567890.M123456P12345.host:2,S")
	email2 := filepath.Join(newDir, "1234567891.M123457P12346.host")

	if err := os.WriteFile(email1, []byte("test email 1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(email2, []byte("test email 2"), 0644); err != nil {
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
	if err := os.MkdirAll(filepath.Join(tmpDir, "cur"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "new"), 0755); err != nil {
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

	if err := os.MkdirAll(mainCur, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subfolderCur, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subfolderNew, 0755); err != nil {
		t.Fatal(err)
	}

	// Create emails in different locations
	email1 := filepath.Join(mainCur, "1234567890.M123456P12345.host:2,S")
	email2 := filepath.Join(subfolderCur, "1234567891.M123457P12346.host:2,S")
	email3 := filepath.Join(subfolderNew, "1234567892.M123458P12347.host")

	if err := os.WriteFile(email1, []byte("email 1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(email2, []byte("email 2"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(email3, []byte("email 3"), 0644); err != nil {
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

	if err := os.MkdirAll(curDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tmpMailDir, 0755); err != nil {
		t.Fatal(err)
	}

	email1 := filepath.Join(curDir, "1234567890.M123456P12345.host:2,S")
	emailTmp := filepath.Join(tmpMailDir, "1234567891.M123457P12346.host")

	if err := os.WriteFile(email1, []byte("email 1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(emailTmp, []byte("temp email"), 0644); err != nil {
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
