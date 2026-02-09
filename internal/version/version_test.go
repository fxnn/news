package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString_WithAllFields(t *testing.T) {
	BuildTimestamp = "2025-01-15T10:30:00Z"
	BuildBranch = "main"
	t.Cleanup(func() {
		BuildTimestamp = ""
		BuildBranch = ""
	})

	result := String()
	assert.Contains(t, result, "2025-01-15T10:30:00Z")
	assert.Contains(t, result, "main")
}

func TestString_WithDefaults(t *testing.T) {
	BuildTimestamp = ""
	BuildBranch = ""

	result := String()
	assert.Contains(t, result, "unknown")
}
