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
	maildir := filepath.Join(tmpDir, "mail")
	storydir := filepath.Join(tmpDir, "stories")
	configContent := "maildir = " + quote(maildir) + "\nstorydir = " + quote(storydir) + `
verbose = true

[llm]
provider = "anthropic"
model = "claude-3-opus"
api_key = "test-key"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
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
	if err := os.Setenv("STORY_EXTRACTOR_LLM_PROVIDER", "gemini"); err != nil {
		t.Fatalf("Failed to set STORY_EXTRACTOR_LLM_PROVIDER: %v", err)
	}
	if err := os.Setenv("STORY_EXTRACTOR_VERBOSE", "true"); err != nil {
		t.Fatalf("Failed to set STORY_EXTRACTOR_VERBOSE: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("STORY_EXTRACTOR_LLM_PROVIDER"); err != nil {
			t.Errorf("Failed to unset STORY_EXTRACTOR_LLM_PROVIDER: %v", err)
		}
	}()
	defer func() {
		if err := os.Unsetenv("STORY_EXTRACTOR_VERBOSE"); err != nil {
			t.Errorf("Failed to unset STORY_EXTRACTOR_VERBOSE: %v", err)
		}
	}()

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
	if err := os.Setenv("UI_SERVER_PORT", "9090"); err != nil {
		t.Fatalf("Failed to set UI_SERVER_PORT: %v", err)
	}
	if err := os.Setenv("UI_SERVER_VERBOSE", "true"); err != nil {
		t.Fatalf("Failed to set UI_SERVER_VERBOSE: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("UI_SERVER_PORT"); err != nil {
			t.Errorf("Failed to unset UI_SERVER_PORT: %v", err)
		}
	}()
	defer func() {
		if err := os.Unsetenv("UI_SERVER_VERBOSE"); err != nil {
			t.Errorf("Failed to unset UI_SERVER_VERBOSE: %v", err)
		}
	}()

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

func TestLoadUiServer_SavedirFromEnvVar(t *testing.T) {
	savedir := filepath.Join(t.TempDir(), "saved-stories")
	t.Setenv("UI_SERVER_SAVEDIR", savedir)

	v := viper.New()
	SetupUiServer(v)
	cfg, err := LoadUiServer(v, "")
	if err != nil {
		t.Fatalf("LoadUiServer() error = %v", err)
	}

	if cfg.Savedir != savedir {
		t.Errorf("Savedir = %v, want %v", cfg.Savedir, savedir)
	}
}

func TestLoadUiServer_SavedirFromConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	storydir := filepath.Join(tmpDir, "stories")
	savedir := filepath.Join(tmpDir, "saved")
	configContent := "storydir = " + quote(storydir) + "\nsavedir = " + quote(savedir) + "\nport = 8080\n"
	configPath := filepath.Join(tmpDir, "ui-server.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatal(err)
	}

	v := viper.New()
	SetupUiServer(v)
	cfg, err := LoadUiServer(v, configPath)
	if err != nil {
		t.Fatalf("LoadUiServer() error = %v", err)
	}

	if cfg.Savedir != savedir {
		t.Errorf("Savedir = %v, want %v", cfg.Savedir, savedir)
	}
}

// quote wraps a path in single quotes for TOML values (TOML literal strings),
// so that backslashes in Windows paths are not treated as escapes.
func quote(s string) string {
	return "'" + s + "'"
}
