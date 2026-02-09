package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/version"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerCmd_Version(t *testing.T) {
	prevTimestamp, prevBranch := version.BuildTimestamp, version.BuildBranch
	t.Cleanup(func() {
		version.BuildTimestamp = prevTimestamp
		version.BuildBranch = prevBranch
	})
	version.BuildTimestamp = "2025-01-15T10:30:00Z"
	version.BuildBranch = "feature/test"

	v := viper.New()
	config.SetupUiServer(v)
	cmd := NewUiServerCmd(v, nil)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "2025-01-15T10:30:00Z")
	assert.Contains(t, output, "feature/test")
}

func TestServerCmd_VersionRejectsArgs(t *testing.T) {
	v := viper.New()
	config.SetupUiServer(v)
	cmd := NewUiServerCmd(v, nil)
	cmd.SetArgs([]string{"version", "foo"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestServerCmd_RequiredInput(t *testing.T) {
	v := viper.New()
	config.SetupUiServer(v)

	cmd := NewUiServerCmd(v, func(cfg *config.UiServer) error {
		return nil
	})

	// Missing storydir
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storydir is required")
}

func TestServerCmd_Configuration(t *testing.T) {
	// Create config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "ui-server.toml")
	err := os.WriteFile(configFile, []byte(`
storydir = "/file/stories"
port = 9090
verbose = true
`), 0644)
	require.NoError(t, err)

	v := viper.New()
	config.SetupUiServer(v)

	var capturedCfg *config.UiServer
	cmd := NewUiServerCmd(v, func(cfg *config.UiServer) error {
		capturedCfg = cfg
		return nil
	})

	t.Setenv("UI_SERVER_PORT", "9999") // Should override file

	cmd.SetArgs([]string{
		"--config", configFile,
		"--verbose=false", // Should override file
	})

	err = cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "/file/stories", capturedCfg.Storydir)
	assert.Equal(t, 9999, capturedCfg.Port) // Env overrides file
	assert.False(t, capturedCfg.Verbose)    // Flag overrides file
}
