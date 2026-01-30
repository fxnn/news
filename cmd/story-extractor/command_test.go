package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fxnn/news/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractorCmd_RequiredFlags(t *testing.T) {
	// Test missing maildir
	v := viper.New()
	config.SetupStoryExtractor(v)
	cmd := NewStoryExtractorCmd(v, func(cfg *config.StoryExtractor) error {
		return nil
	})
	cmd.SetArgs([]string{"--storydir", "/tmp", "--config", "/dev/null"}) // config is optional but we provide dummy to avoid file read error?
	// Actually LoadStoryExtractor fails if file not found when cfgFile is set.
	// If cfgFile is empty, it looks for defaults.
	// We should probably just pass args.

	// We need config file to be optional or present. Defaults check home/params.
	// But LoadStoryExtractor logic: if cfgFile != "", reads it.
	// If not provided, it searches. If not found, it continues without error (unless it fails to parse).

	// To test "Missing Flags", we just set args.
	cmd.SetArgs([]string{"--storydir", "/tmp/stories"})
	// llm.api_key is also required.
	t.Setenv("STORY_EXTRACTOR_LLM_API_KEY", "dummy-key")

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maildir is required")
}

func TestExtractorCmd_FlagsPrecedence(t *testing.T) {
	v := viper.New()
	config.SetupStoryExtractor(v)

	var capturedCfg *config.StoryExtractor
	cmd := NewStoryExtractorCmd(v, func(cfg *config.StoryExtractor) error {
		capturedCfg = cfg
		return nil
	})

	cmd.SetArgs([]string{
		"--maildir", "/flag/maildir",
		"--storydir", "/flag/storydir",
		"--verbose",
	})
	t.Setenv("STORY_EXTRACTOR_LLM_API_KEY", "dummy-key")

	err := cmd.Execute()
	require.NoError(t, err)
	require.NotNil(t, capturedCfg)

	assert.Equal(t, "/flag/maildir", capturedCfg.Maildir)
	assert.Equal(t, "/flag/storydir", capturedCfg.Storydir)
	assert.True(t, capturedCfg.Verbose)
}

func TestExtractorCmd_EnvVarPrecedence(t *testing.T) {
	v := viper.New()
	config.SetupStoryExtractor(v)

	var capturedCfg *config.StoryExtractor
	cmd := NewStoryExtractorCmd(v, func(cfg *config.StoryExtractor) error {
		capturedCfg = cfg
		return nil
	})

	// Set env vars
	t.Setenv("STORY_EXTRACTOR_MAILDIR", "/env/maildir")
	t.Setenv("STORY_EXTRACTOR_STORYDIR", "/env/storydir")
	t.Setenv("STORY_EXTRACTOR_VERBOSE", "true")
	t.Setenv("STORY_EXTRACTOR_LLM_API_KEY", "env-key")

	// No flags provided
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	require.NotNil(t, capturedCfg)

	assert.Equal(t, "/env/maildir", capturedCfg.Maildir)
	assert.Equal(t, "/env/storydir", capturedCfg.Storydir)
	assert.Equal(t, "env-key", capturedCfg.LLM.APIKey)
	assert.True(t, capturedCfg.Verbose)
}

func TestExtractorCmd_ConfigFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "story-extractor.toml")
	configContent := `
maildir = "/file/maildir"
storydir = "/file/storydir"
verbose = true
[llm]
api_key = "file-key"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	v := viper.New()
	config.SetupStoryExtractor(v)

	var capturedCfg *config.StoryExtractor
	cmd := NewStoryExtractorCmd(v, func(cfg *config.StoryExtractor) error {
		capturedCfg = cfg
		return nil
	})

	cmd.SetArgs([]string{"--config", configFile})

	err = cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "/file/maildir", capturedCfg.Maildir)
	assert.Equal(t, "/file/storydir", capturedCfg.Storydir)
	assert.Equal(t, "file-key", capturedCfg.LLM.APIKey)
	assert.True(t, capturedCfg.Verbose)
}

func TestExtractorCmd_PrecedenceOrder(t *testing.T) {
	// File < Env < Flag

	// 1. Create Config File
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "precedence.toml")
	err := os.WriteFile(configFile, []byte(`maildir = "file"`), 0644)
	require.NoError(t, err)

	v := viper.New()
	config.SetupStoryExtractor(v)

	var capturedCfg *config.StoryExtractor
	cmd := NewStoryExtractorCmd(v, func(cfg *config.StoryExtractor) error {
		capturedCfg = cfg
		return nil
	})

	// 2. Set Env Var (Should override file)
	t.Setenv("STORY_EXTRACTOR_MAILDIR", "env")
	t.Setenv("STORY_EXTRACTOR_STORYDIR", "env-story") // Provide required
	t.Setenv("STORY_EXTRACTOR_LLM_API_KEY", "key")

	// 3. Set Flag (Should override env)
	cmd.SetArgs([]string{"--maildir", "flag", "--config", configFile})

	err = cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "flag", capturedCfg.Maildir)
}
