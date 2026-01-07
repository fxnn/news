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

func TestLoad_MinimalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal.toml")

	minimalContent := `[llm]
provider = "openai"
model = "gpt-4"
`

	if err := os.WriteFile(configPath, []byte(minimalContent), 0644); err != nil {
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

	if cfg.LLM.APIKey != "" {
		t.Errorf("APIKey = %v, want empty string", cfg.LLM.APIKey)
	}

	if cfg.LLM.BaseURL != "" {
		t.Errorf("BaseURL = %v, want empty string", cfg.LLM.BaseURL)
	}
}
