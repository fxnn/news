package main

// Summarizer is the minimal interface for any LLM-based summarizer.
type Summarizer interface {
	Summarize(text string) (string, error)
}

// This file now only contains comments or potentially shared types/constants
// related to summarization if needed in the future. The global Summarize function
// has been removed in favor of explicit instantiation in main.go.
