package main

import (
	"os"
	"testing"
)

func TestSummarizeImplementations(t *testing.T) {
	// Define the test cases once
	testCases := []struct {
		name             string
		text             string
		wantStoriesStub  []Story // Precise expectation for Stub. Length indicates expected story count.
		wantErr          error
		expectAnyStory   bool    // If true, expect len(got) > 0. If false, expect len(got) == 0 or got == nil.
	}{
		{
			name:            "Empty text",
			text:            "",
			wantStoriesStub: nil,
			wantErr:         nil,
			expectAnyStory:  false,
		},
		{
			name: "Plain text - Single story, with URL",
			text: "This is a reasonably long email body that requires summarization. " +
				"It discusses the project status. More info at http://example.com/project-status.",
			wantStoriesStub: []Story{{
				Headline: "Summary",
				Teaser:   "This is a reasonably long email body that requires summarization.",
				URL:      "",
			}},
			wantErr:        nil,
			expectAnyStory: true,
		},
		{
			name: "Plain text - Short story, with URL",
			text: "OK. Read more http://example.com/ok.",
			wantStoriesStub: []Story{{
				Headline: "Summary",
				Teaser:   "OK.",
				URL:      "",
			}},
			wantErr:        nil,
			expectAnyStory: true,
		},
		{
			name: "Plain text - Single story, no URL",
			text: "This is a simple statement. It stands alone.",
			wantStoriesStub: []Story{{
				Headline: "Summary",
				Teaser:   "This is a simple statement.",
				URL:      "",
			}},
			wantErr:        nil,
			expectAnyStory: true,
		},
		{
			name: "HTML text - Single story, with URL",
			text: "<p>This is <b>HTML</b> content. Learn more <a href='http://example.com/html-story'>here</a>.</p>",
			wantStoriesStub: []Story{{
				Headline: "Summary",
				Teaser:   "<p>This is <b>HTML</b> content.", // Stub takes first sentence of raw HTML
				URL:      "",
			}},
			wantErr:        nil,
			expectAnyStory: true,
		},
		{
			name: "HTML text - Single story, no URL",
			text: "<div><p>Just a piece of HTML. Indeed.</p></div>",
			wantStoriesStub: []Story{{
				Headline: "Summary",
				Teaser:   "<div><p>Just a piece of HTML.", // Stub takes first sentence of raw HTML
				URL:      "",
			}},
			wantErr:        nil,
			expectAnyStory: true,
		},
		{
			name: "Plain text - Multiple segments (potential for multiple stories)",
			text: "First topic is about apples. They are good. Second topic is about bananas. They are yellow. Find out more at http://fruits.com.",
			wantStoriesStub: []Story{{
				Headline: "Summary",
				Teaser:   "First topic is about apples.",
				URL:      "",
			}},
			wantErr:        nil,
			expectAnyStory: true, // LLM might produce >1 story, or 1 combined story. We check for at least one.
		},
		{
			name: "HTML text - Multiple segments (potential for multiple stories)",
			text: "<h1>Topic One</h1><p>Content for topic one. <a href='http://example.com/one'>Link1</a></p><h2>Topic Two</h2><p>Content for topic two. No link here.</p>",
			// Stub will take the first sentence from the raw HTML string.
			// The first '.' appears after "Content for topic one."
			wantStoriesStub: []Story{{
				Headline: "Summary",
				Teaser:   "<h1>Topic One</h1><p>Content for topic one.",
				URL:      "",
			}},
			wantErr:        nil,
			expectAnyStory: true,
		},
	}

	// Define the summarizers to test
	summarizers := map[string]Summarizer{
		"Stub": NewStubSummarizer(),
	}

	// Attempt to add LangChain summarizer, skip if API key is missing
	lcSummarizer, err := NewLangChainSummarizer()
	if err == nil {
		summarizers["LangChain"] = lcSummarizer
	} else if os.Getenv("OPENAI_API_KEY") == "" {
		t.Log("Skipping LangChain summarizer tests: OPENAI_API_KEY not set")
	} else {
		// API key is set, but creation failed for another reason
		t.Fatalf("Failed to create LangChain summarizer even though OPENAI_API_KEY is set: %v", err)
	}

	// Run tests for each summarizer
	for sName, summarizer := range summarizers {
		t.Run(sName, func(t *testing.T) {
			// Run each test case for the current summarizer
			for _, tt := range testCases {
				t.Run(tt.name, func(t *testing.T) {
					got, err := summarizer.Summarize(tt.text) // Call the method on the instance

					// Check for unexpected errors
					if err != tt.wantErr {
						// If we expected a specific error (like ErrSummarizationNotImplemented in the future)
						// and got a different one, fail.
						// If we expected no error (wantErr == nil) and got one, fail.
						t.Errorf("Summarize() error = %v, wantErr %v", err, tt.wantErr)
						return
					}

					// If an error was expected, we don't need to check the output
					if tt.wantErr != nil {
						return
					}

					// General story expectation check (applies to all summarizers)
					if tt.expectAnyStory {
						if got == nil || len(got) == 0 {
							t.Errorf("Summarize() got an empty or nil slice of stories, want non-empty for text: %q", tt.text)
							return // Avoid further checks on nil/empty slice
						}
						// Basic checks for all stories returned (headline, teaser non-empty)
						for i, story := range got {
							if story.Headline == "" {
								t.Errorf("Summarize() story %d has empty Headline for text: %q (Summarizer: %s)", i, tt.text, sName)
							}
							if story.Teaser == "" {
								t.Errorf("Summarize() story %d has empty Teaser for text: %q (Summarizer: %s)", i, tt.text, sName)
							}
							// Story.URL is not reliably populated by current LLM prompt or stub,
							// so a generic check for non-empty URL is too strict here.
							// Specific check for stub's empty URL is done below.
						}
					} else { // !tt.expectAnyStory (i.e., expect no stories)
						if got != nil && len(got) > 0 {
							t.Errorf("Summarize() got %v, want empty or nil slice of stories for text: %q (Summarizer: %s)", got, tt.text, sName)
						}
					}

					// Specific checks for Stub (number of stories and content)
					if sName == "Stub" {
						expectedNumStoriesStub := 0
						if tt.wantStoriesStub != nil {
							expectedNumStoriesStub = len(tt.wantStoriesStub)
						}

						if len(got) != expectedNumStoriesStub {
							t.Errorf("Summarize() for Stub, got %d stories, want %d stories for text: %q. Got: %v, Want: %v", len(got), expectedNumStoriesStub, tt.text, got, tt.wantStoriesStub)
						} else if tt.wantStoriesStub != nil { // Lengths match, now check content if expected stories are defined
							for i := range got {
								if got[i].Headline != tt.wantStoriesStub[i].Headline {
									t.Errorf("Summarize() for Stub, story %d Headline got %q, want %q for text: %q", i, got[i].Headline, tt.wantStoriesStub[i].Headline, tt.text)
								}
								if got[i].Teaser != tt.wantStoriesStub[i].Teaser {
									t.Errorf("Summarize() for Stub, story %d Teaser got %q, want %q for text: %q", i, got[i].Teaser, tt.wantStoriesStub[i].Teaser, tt.text)
								}
								if got[i].URL != tt.wantStoriesStub[i].URL {
									t.Errorf("Summarize() for Stub, story %d URL got %q, want %q for text: %q", i, got[i].URL, tt.wantStoriesStub[i].URL, tt.text)
								}
							}
						}
					}
					// For LangChain, tt.expectAnyStory and the general loop for non-empty headline/teaser cover current expectations.
					// More specific checks for LLM (e.g., min number of stories) could be added here if needed.

				}) // End of t.Run for test case
			} // End of loop over test cases
		}) // End of t.Run for summarizer type
	} // End of loop over summarizers
}
