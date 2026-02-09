package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString_WithAllFields(t *testing.T) {
	prevTimestamp, prevBranch := BuildTimestamp, BuildBranch
	t.Cleanup(func() {
		BuildTimestamp = prevTimestamp
		BuildBranch = prevBranch
	})

	BuildTimestamp = "2025-01-15T10:30:00Z"
	BuildBranch = "main"

	result := String()
	assert.Contains(t, result, "2025-01-15T10:30:00Z")
	assert.Contains(t, result, "main")
}

func TestString_WithDefaults(t *testing.T) {
	prevTimestamp, prevBranch := BuildTimestamp, BuildBranch
	t.Cleanup(func() {
		BuildTimestamp = prevTimestamp
		BuildBranch = prevBranch
	})

	BuildTimestamp = ""
	BuildBranch = ""

	result := String()
	assert.Contains(t, result, "unknown")
}
