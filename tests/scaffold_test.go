package scaffold

import (
	"os"
	"os/exec"
	"testing"
)

func TestGoModExists(t *testing.T) {
	// The test's working directory is the package directory (`tests`),
	// so we check for go.mod in the parent (project root).
	if _, err := os.Stat("../go.mod"); os.IsNotExist(err) {
		t.Fatal("go.mod does not exist in project root")
	}
}

func TestMainPackageBuilds(t *testing.T) {
	cmd := exec.Command("go", "build", "../cmd/email-story-extractor")
	if err := cmd.Run(); err != nil {
		t.Fatalf("building main package failed: %v", err)
	}
}

func TestCIWorkflowExists(t *testing.T) {
	if _, err := os.Stat("../.github/workflows/go.yml"); os.IsNotExist(err) {
		t.Fatal(".github/workflows/go.yml does not exist in project root")
	}
}
