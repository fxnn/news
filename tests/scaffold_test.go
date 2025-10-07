package scaffold

import (
	"os"
	"testing"
)

func TestGoModExists(t *testing.T) {
	// The test's working directory is the package directory (`tests`),
	// so we check for go.mod in the parent (project root).
	if _, err := os.Stat("../go.mod"); os.IsNotExist(err) {
		t.Fatal("go.mod does not exist in project root")
	}
}
