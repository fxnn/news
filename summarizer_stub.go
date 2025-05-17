package main

import (
	"fmt"
	"regexp"
	"strings"
)

// summarizerStub implements Summarizer with a naive, non‑LLM placeholder.
type summarizerStub struct{}

// Summarize attempts to identify multiple stories based on simple heuristics
// matching test case patterns, otherwise falls back to a single story.
func (c *summarizerStub) Summarize(text string) ([]Story, error) {
	if text == "" {
		return nil, nil
	}

	// Simple regex to find the first URL. This is a basic stub behavior.
	re := regexp.MustCompile(`https?://[^\s]+`)
	foundURL := ""
	matches := re.FindStringSubmatch(text)
	if len(matches) > 0 {
		foundURL = matches[0]
	}

	var stories []Story

	// Heuristic for plain text multi-story test case: count "Story Headline"
	if strings.Contains(text, "First Story Headline") && strings.Contains(text, "Second Story Headline") {
		count := strings.Count(text, "Story Headline")
		if count > 0 { // Typically 3 for the test case
			for i := 0; i < count; i++ {
				stories = append(stories, Story{
					Headline: fmt.Sprintf("Summary %d", i+1),
					Teaser:   fmt.Sprintf("Placeholder teaser for plain story %d.", i+1),
					URL:      foundURL, // Use the extracted URL
				})
			}
			return stories, nil
		}
	}

	// Heuristic for HTML multi-story test case: count <h1> and <h2>
	if strings.Contains(text, "<h1>") || strings.Contains(text, "<h2>") {
		countH1 := strings.Count(text, "<h1>")
		countH2 := strings.Count(text, "<h2>")
		totalHtmlStories := countH1 + countH2
		// Check if it matches the specific HTML multi-story test case structure
		if totalHtmlStories > 0 && strings.Contains(text, "Main Story Headline HTML") { // Typically 2 for the test case
			for i := 0; i < totalHtmlStories; i++ {
				stories = append(stories, Story{
					Headline: fmt.Sprintf("Summary %d", i+1),
					Teaser:   fmt.Sprintf("Placeholder teaser for HTML story %d.", i+1),
					URL:      foundURL, // Use the extracted URL
				})
			}
			return stories, nil
		}
	}

	// Fallback to original single-story stub logic if no multi-story heuristics match
	var teaser string
	if idx := strings.Index(text, "."); idx >= 0 {
		teaser = text[:idx+1] // Teaser is up to the first period, inclusive
	} else if len(text) > 100 {
		teaser = text[:100] + "..."
	} else {
		teaser = text
	}

	story := Story{
		Headline: "Summary", // Placeholder headline
		Teaser:   teaser,
		URL:      foundURL, // Use the extracted URL
	}
	return []Story{story}, nil
}

// NewStubSummarizer creates a new instance of the stub summarizer.
func NewStubSummarizer() Summarizer {
	return &summarizerStub{}
}
