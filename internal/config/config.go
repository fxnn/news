package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// StoryExtractor configuration for the extraction CLI tool
type StoryExtractor struct {
	LLM        LLM    `mapstructure:"llm"`
	Maildir    string `mapstructure:"maildir"`
	Storydir   string `mapstructure:"storydir"`
	Limit      int    `mapstructure:"limit"`
	Verbose    bool   `mapstructure:"verbose"`
	LogHeaders bool   `mapstructure:"log_headers"`
	LogBodies  bool   `mapstructure:"log_bodies"`
	LogStories bool   `mapstructure:"log_stories"`
}

// UiServer configuration for the web server
type UiServer struct {
	Storydir string `mapstructure:"storydir"`
	Savedir  string `mapstructure:"savedir"`
	Port     int    `mapstructure:"port"`
	Verbose  bool   `mapstructure:"verbose"`
}

// LLM represents the configuration for a Large Language Model provider.
type LLM struct {
	Provider string `mapstructure:"provider"`
	Model    string `mapstructure:"model"`
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
}

// SetupStoryExtractor configures defaults for the story extractor
func SetupStoryExtractor(v *viper.Viper) {
	v.SetDefault("llm.provider", "openai")
	v.SetDefault("llm.model", "gpt-4o-mini")
	v.SetDefault("llm.api_key", "")
	v.SetDefault("llm.base_url", "https://api.openai.com/v1")
	v.SetDefault("verbose", false)

	v.SetEnvPrefix("STORY_EXTRACTOR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
}

// SetupUiServer configures defaults for the UI server
func SetupUiServer(v *viper.Viper) {
	v.SetDefault("savedir", "")
	v.SetDefault("port", 8080)
	v.SetDefault("verbose", false)

	v.SetEnvPrefix("UI_SERVER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
}

func loadConfig(v *viper.Viper, cfgFile, configName string, target interface{}) error {
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME")
		v.SetConfigName(configName)
		v.SetConfigType("toml")
	}

	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return fmt.Errorf("error reading config file: %w", err)
		}
		if cfgFile != "" {
			return fmt.Errorf("error reading config file %s: %w", cfgFile, err)
		}
	}

	if err := v.Unmarshal(target); err != nil {
		return fmt.Errorf("unable to decode config: %w", err)
	}
	return nil
}

// LoadStoryExtractor loads configuration for the extractor
func LoadStoryExtractor(v *viper.Viper, cfgFile string) (*StoryExtractor, error) {
	var cfg StoryExtractor
	if err := loadConfig(v, cfgFile, "story-extractor", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadUiServer loads configuration for the server
func LoadUiServer(v *viper.Viper, cfgFile string) (*UiServer, error) {
	var cfg UiServer
	if err := loadConfig(v, cfgFile, "ui-server", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
