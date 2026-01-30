package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadStoryExtractor_Defaults(t *testing.T) {
	v := viper.New()
	SetupStoryExtractor(v)

	// No config file, just defaults
	cfg, err := LoadStoryExtractor(v, "")
	if err != nil {
		t.Fatalf("LoadStoryExtractor() error = %v", err)
	}

	if cfg.LLM.Provider != "openai" {
		t.Errorf("LLM.Provider = %v, want openai", cfg.LLM.Provider)
	}
	if cfg.LLM.Model != "gpt-4o-mini" {
		t.Errorf("LLM.Model = %v, want gpt-4o-mini", cfg.LLM.Model)
	}
	if cfg.Verbose != false {
		t.Errorf("Verbose = %v, want false", cfg.Verbose)
	}
}

func TestLoadStoryExtractor_ConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
maildir = "/tmp/mail"
storydir = "/tmp/stories"
verbose = true

[llm]
provider = "anthropic"
model = "claude-3-opus"
api_key = "test-key"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	v := viper.New()
	SetupStoryExtractor(v)
	cfg, err := LoadStoryExtractor(v, configPath)
	if err != nil {
		t.Fatalf("LoadStoryExtractor() error = %v", err)
	}

	if cfg.LLM.Provider != "anthropic" {
		t.Errorf("LLM.Provider = %v, want anthropic", cfg.LLM.Provider)
	}
	if cfg.Verbose != true {
		t.Errorf("Verbose = %v, want true", cfg.Verbose)
	}
}

func TestLoadStoryExtractor_EnvVars(t *testing.T) {
	os.Setenv("STORY_EXTRACTOR_LLM_PROVIDER", "gemini")
	os.Setenv("STORY_EXTRACTOR_VERBOSE", "true")
	defer os.Unsetenv("STORY_EXTRACTOR_LLM_PROVIDER")
	defer os.Unsetenv("STORY_EXTRACTOR_VERBOSE")

	v := viper.New()
	SetupStoryExtractor(v)
	cfg, err := LoadStoryExtractor(v, "")
	if err != nil {
		t.Fatalf("LoadStoryExtractor() error = %v", err)
	}

	if cfg.LLM.Provider != "gemini" {
		t.Errorf("LLM.Provider = %v, want gemini", cfg.LLM.Provider)
	}
	if cfg.Verbose != true {
		t.Errorf("Verbose = %v, want true", cfg.Verbose)
	}
}

func TestLoadUiServer_EnvVars(t *testing.T) {
	os.Setenv("UI_SERVER_PORT", "9090")
	os.Setenv("UI_SERVER_VERBOSE", "true")
	defer os.Unsetenv("UI_SERVER_PORT")
	defer os.Unsetenv("UI_SERVER_VERBOSE")

	v := viper.New()
	SetupUiServer(v)
	cfg, err := LoadUiServer(v, "")
	if err != nil {
		t.Fatalf("LoadUiServer() error = %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Port = %v, want 9090", cfg.Port)
	}
	if cfg.Verbose != true {
		t.Errorf("Verbose = %v, want true", cfg.Verbose)
	}
}
