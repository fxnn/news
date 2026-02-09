package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/logger"
	"github.com/fxnn/news/internal/story"
	"github.com/fxnn/news/internal/storyreader"
	"github.com/fxnn/news/internal/storysaver"
	"github.com/fxnn/news/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed index.html
var indexHTML []byte

func main() {
	v := viper.New()
	config.SetupUiServer(v)

	cmd := NewUiServerCmd(v, nil)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type RunServerFunc func(cfg *config.UiServer) error

func NewUiServerCmd(v *viper.Viper, runFn RunServerFunc) *cobra.Command {
	var cfgFile string

	cmd := &cobra.Command{
		Use:   "ui-server",
		Short: "Start the UI server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadUiServer(v, cfgFile)
			if err != nil {
				return err
			}

			if cfg.Storydir == "" {
				return fmt.Errorf("storydir is required")
			}

			if cfg.Savedir == "" {
				return fmt.Errorf("savedir is required")
			}

			// Execute injected run function (for testing) or default logic
			if runFn != nil {
				return runFn(cfg)
			}

			log := logger.New(cfg.Verbose)
			addr := fmt.Sprintf(":%d", cfg.Port)
			log.Info("Starting UI server", "addr", addr, "storydir", cfg.Storydir)

			mux := http.NewServeMux()

			mux.HandleFunc("/api/stories", func(w http.ResponseWriter, r *http.Request) {
				handleStories(w, r, cfg.Storydir, cfg.Savedir)
			})

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(indexHTML)
			})

			if err := http.ListenAndServe(addr, mux); err != nil {
				return err
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&cfgFile, "config", "", "config file (default: ./ui-server.toml or $HOME/ui-server.toml)")
	f.String("storydir", "", "Path to stories")
	f.String("savedir", "", "Path to saved stories")
	f.Int("port", 8080, "Port to listen on")
	f.Bool("verbose", false, "Enable verbose output")

	v.BindPFlag("storydir", f.Lookup("storydir"))
	v.BindPFlag("savedir", f.Lookup("savedir"))
	v.BindPFlag("port", f.Lookup("port"))
	v.BindPFlag("verbose", f.Lookup("verbose"))

	cmd.AddCommand(version.NewCommand())

	return cmd
}

// storyResponse wraps a story with its saved status for the API response.
// Keeps the save concern in the UI layer, separate from the shared Story model.
type storyResponse struct {
	story.Story
	Saved bool `json:"saved"`
}

func handleStories(w http.ResponseWriter, r *http.Request, storydir, savedir string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stories, err := storyreader.ReadStories(storydir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read stories: %v", err), http.StatusInternalServerError)
		return
	}

	savedSet := map[string]bool{}
	if savedir != "" {
		savedSet, _ = storysaver.ListSavedFilenames(savedir)
	}

	response := make([]storyResponse, len(stories))
	for i, s := range stories {
		response[i] = storyResponse{Story: s, Saved: savedSet[s.Filename]}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
