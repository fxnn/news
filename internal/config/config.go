package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	LLM LLMConfig `toml:"llm"`
}

type LLMConfig struct {
	Provider string `toml:"provider"`
	Model    string `toml:"model"`
	APIKey   string `toml:"api_key"`
	BaseURL  string `toml:"base_url"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TOML config: %w", err)
	}

	// Allow environment variable override for API key (more secure)
	if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
		cfg.LLM.APIKey = envKey
	}

	// Validate that we have an API key from either source
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("API key is required (set in config file or OPENAI_API_KEY environment variable)")
	}

	return &cfg, nil
}
