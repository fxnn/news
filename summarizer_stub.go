package main

import "strings"

// summarizerStub implements Summarizer with a naive, non‑LLM placeholder.
type summarizerStub struct{}

// Summarize returns a single Story. The teaser is generated based on the first sentence
// (up to the first '.') or truncates at 100 chars. It returns nil for empty input.
func (c *summarizerStub) Summarize(text string) ([]Story, error) {
	if text == "" {
		return nil, nil
	}

	var teaser string
	if idx := strings.Index(text, "."); idx >= 0 {
		// Teaser is up to the first period, inclusive
		teaser = text[:idx+1]
	} else if len(text) > 100 {
		teaser = text[:100] + "..."
	} else {
		teaser = text
	}

	story := Story{
		Headline: "Summary", // Placeholder headline
		Teaser:   teaser,
		URL:      "", // Placeholder URL, stub doesn't extract URLs
	}
	return []Story{story}, nil
}

// NewStubSummarizer creates a new instance of the stub summarizer.
func NewStubSummarizer() Summarizer {
	return &summarizerStub{}
}
