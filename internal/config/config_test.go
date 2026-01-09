package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `[llm]
provider = "openai"
model = "gpt-4"
api_key = "test-key"
base_url = "https://api.openai.com/v1"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.LLM.Provider != "openai" {
		t.Errorf("Provider = %v, want openai", cfg.LLM.Provider)
	}

	if cfg.LLM.Model != "gpt-4" {
		t.Errorf("Model = %v, want gpt-4", cfg.LLM.Model)
	}

	if cfg.LLM.APIKey != "test-key" {
		t.Errorf("APIKey = %v, want test-key", cfg.LLM.APIKey)
	}

	if cfg.LLM.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("BaseURL = %v, want https://api.openai.com/v1", cfg.LLM.BaseURL)
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/config.toml")
	if err == nil {
		t.Error("Load() expected error for nonexistent file, got nil")
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.toml")

	invalidContent := `[llm
provider = "openai"
`

	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() expected error for invalid TOML, got nil")
	}
}

func TestLoad_MissingAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "no-key.toml")

	content := `[llm]
provider = "openai"
model = "gpt-4"
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Clear environment variable to ensure test isolation
	oldEnv := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if oldEnv != "" {
			os.Setenv("OPENAI_API_KEY", oldEnv)
		}
	}()

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() expected error for missing API key, got nil")
	}
}

func TestLoad_EnvVarAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `[llm]
provider = "openai"
model = "gpt-4"
base_url = "https://api.openai.com/v1"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Set environment variable
	oldEnv := os.Getenv("OPENAI_API_KEY")
	os.Setenv("OPENAI_API_KEY", "env-test-key")
	defer func() {
		if oldEnv != "" {
			os.Setenv("OPENAI_API_KEY", oldEnv)
		} else {
			os.Unsetenv("OPENAI_API_KEY")
		}
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.LLM.APIKey != "env-test-key" {
		t.Errorf("APIKey = %v, want env-test-key (from environment)", cfg.LLM.APIKey)
	}
}

func TestLoad_EnvVarOverridesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `[llm]
provider = "openai"
model = "gpt-4"
api_key = "config-key"
base_url = "https://api.openai.com/v1"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Set environment variable - should override config file
	oldEnv := os.Getenv("OPENAI_API_KEY")
	os.Setenv("OPENAI_API_KEY", "env-override-key")
	defer func() {
		if oldEnv != "" {
			os.Setenv("OPENAI_API_KEY", oldEnv)
		} else {
			os.Unsetenv("OPENAI_API_KEY")
		}
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.LLM.APIKey != "env-override-key" {
		t.Errorf("APIKey = %v, want env-override-key (environment should override config)", cfg.LLM.APIKey)
	}
}
