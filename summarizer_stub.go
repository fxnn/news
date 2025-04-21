package main

import "strings"

// Summarizer is the minimal interface for any LLM-based summarizer.
type Summarizer interface {
	Summarize(text string) (string, error)
}

// summarizerStub implements Summarizer with a naive, non‑LLM placeholder.
type summarizerStub struct{}

// Summarize returns the first sentence (up to the first '.') or
// truncates at 100 chars, never empty if input is non‑empty.
func (c *summarizerStub) Summarize(text string) (string, error) {
	if text == "" {
		return "", nil
	}
	if idx := strings.Index(text, "."); idx >= 0 {
		// Return up to the first period, inclusive
		return text[:idx+1], nil
	}
	if len(text) > 100 {
		return text[:100] + "...", nil
	}
	return text, nil
}

// NewStubSummarizer creates a new instance of the stub summarizer.
func NewStubSummarizer() Summarizer {
	return &summarizerStub{}
}
