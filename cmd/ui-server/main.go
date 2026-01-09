package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fxnn/news/internal/storyreader"
)

//go:embed index.html
var indexHTML []byte

func main() {
	storydir := flag.String("storydir", "", "Path to the story directory (required)")
	port := flag.String("port", "8080", "Port to listen on")

	flag.Parse()

	if *storydir == "" {
		fmt.Fprintln(os.Stderr, "Error: --storydir is required")
		flag.Usage()
		os.Exit(1)
	}

	// Set up HTTP routes
	http.HandleFunc("/api/stories", func(w http.ResponseWriter, r *http.Request) {
		handleStories(w, r, *storydir)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})

	addr := ":" + *port
	log.Printf("Starting UI server on http://localhost%s\n", addr)
	log.Printf("Stories directory: %s\n", *storydir)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleStories(w http.ResponseWriter, r *http.Request, storydir string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stories, err := storyreader.ReadStories(storydir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read stories: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}
