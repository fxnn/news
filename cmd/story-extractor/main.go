package main

import (
	"fmt"
	"os"

	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/extractor"
	"github.com/fxnn/news/internal/llm"
	"github.com/fxnn/news/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	var cfgFile string
	v := viper.New()

	cmd := &cobra.Command{
		Use:   "story-extractor",
		Short: "Extract stories from emails",
		RunE: func(cmd *cobra.Command, args []string) error {
			config.SetupStoryExtractor(v)

			cfg, err := config.LoadStoryExtractor(v, cfgFile)
			if err != nil {
				return err
			}

			// Validate required
			if cfg.Maildir == "" {
				return fmt.Errorf("maildir is required (via flag, config, or env)")
			}
			if cfg.Storydir == "" {
				return fmt.Errorf("storydir is required")
			}
			if cfg.LLM.APIKey == "" {
				return fmt.Errorf("llm.api_key is required (via config or STORY_EXTRACTOR_LLM_API_KEY env var)")
			}

			// Initialize dependencies
			log := logger.New(cfg.Verbose)
			log.Info("starting story extractor", "maildir", cfg.Maildir, "storydir", cfg.Storydir)

			storyExtractor := llm.NewOpenAIExtractor(&cfg.LLM)

			processor := extractor.NewProcessor(cfg, log, storyExtractor)
			result, err := processor.Run()
			if err != nil {
				log.Error("processing failed", "error", err)
				return err
			}

			if result.Errors > 0 {
				return fmt.Errorf("processing completed with %d errors", result.Errors)
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&cfgFile, "config", "", "config file (default is $HOME/config.toml)")
	f.String("maildir", "", "Path to the Maildir directory")
	f.String("storydir", "", "Output directory for story files")
	f.Int("limit", 0, "Limit number of emails to process")
	f.Bool("verbose", false, "Enable verbose output")
	f.Bool("log-headers", false, "Log email headers")
	f.Bool("log-bodies", false, "Log email bodies")
	f.Bool("log-stories", false, "Log extracted stories")

	v.BindPFlag("maildir", f.Lookup("maildir"))
	v.BindPFlag("storydir", f.Lookup("storydir"))
	v.BindPFlag("limit", f.Lookup("limit"))
	v.BindPFlag("verbose", f.Lookup("verbose"))
	v.BindPFlag("log_headers", f.Lookup("log-headers"))
	v.BindPFlag("log_bodies", f.Lookup("log-bodies"))
	v.BindPFlag("log_stories", f.Lookup("log-stories"))

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
