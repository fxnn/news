package version

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewCommand(t *testing.T) {
	prevTimestamp, prevBranch := BuildTimestamp, BuildBranch
	t.Cleanup(func() {
		BuildTimestamp = prevTimestamp
		BuildBranch = prevBranch
	})

	BuildTimestamp = "2025-06-01T08:00:00Z"
	BuildBranch = "feature/x"

	cmd := NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "2025-06-01T08:00:00Z")
	assert.Contains(t, buf.String(), "feature/x")
}

func TestNewCommand_RejectsArgs(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"foo"})

	err := cmd.Execute()
	assert.Error(t, err)
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
