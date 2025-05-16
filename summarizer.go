package main

// Story represents a single news item or section extracted from an email.
type Story struct {
	Headline string
	Teaser   string
	URL      string
}

// Summarizer is the minimal interface for any LLM-based summarizer.
// It now returns a slice of Story objects.
type Summarizer interface {
	Summarize(text string) ([]Story, error)
}

// This file now only contains comments or potentially shared types/constants
// related to summarization if needed in the future. The global Summarize function
// has been removed in favor of explicit instantiation in main.go.
