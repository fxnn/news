package main

import (
	"os"
	"testing"
)

func TestSummarizeImplementations(t *testing.T) {
	// Define the test cases once
	testCases := []struct {
		name string
		text string
		// For stub, we can check wantStories. For LLM, we check checkNonEmptyStories and basic structure.
		wantStories         []Story
		wantErr             error
		checkNonEmptyStories bool // Flag to check if the output slice of stories should be non-empty
	}{
		{
			name:                "Empty text",
			text:                "",
			wantStories:         nil, // Or []Story{}
			wantErr:             nil,
			checkNonEmptyStories: false, // Expect empty slice of stories for empty text
		},
		{
			name: "Normal text - single story expected (for stub)",
			text: "This is a reasonably long email body that requires summarization. " +
				"It discusses the project status, upcoming deadlines, and action items for the team. " +
				"We need to ensure the summary captures the key points without being too verbose. More info at http://example.com/project-status",
			// wantStories will be specific for the stub, for LLM we just check if it's non-empty and fields are populated
			wantErr:             nil,
			checkNonEmptyStories: true, // Expect a non-empty slice of stories
		},
		{
			name:                "Short text - single story expected (for stub)",
			text:                "OK. Read more http://example.com/ok",
			wantErr:             nil,
			checkNonEmptyStories: true, // Expect a non-empty slice of stories even for short text
		},
		// Add more test cases here if needed
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

					// Check if the slice of stories should be non-empty
					if tt.checkNonEmptyStories {
						if got == nil || len(got) == 0 {
							t.Errorf("Summarize() got an empty or nil slice of stories, want non-empty for text: %q", tt.text)
							return // Avoid further checks on nil/empty slice
						}
						// For non-empty, also check if stories have populated fields (basic check for LLM)
						for i, story := range got {
							if story.Headline == "" {
								t.Errorf("Summarize() story %d has empty Headline for text: %q", i, tt.text)
							}
							if story.Teaser == "" {
								t.Errorf("Summarize() story %d has empty Teaser for text: %q", i, tt.text)
							}
							// URL can sometimes be empty if not found, so this check might be too strict for all cases.
							// For now, let's assume a URL is usually expected if a story is present.
							// if story.URL == "" {
							//  t.Errorf("Summarize() story %d has empty URL for text: %q", i, tt.text)
							// }
						}
					}

					// Check if the slice of stories should be empty
					if !tt.checkNonEmptyStories && (got != nil && len(got) > 0) {
						t.Errorf("Summarize() got = %v, want empty or nil slice of stories for empty text", got)
					}

					// Specific checks for stub - this part will need actual implementation of stub to match
					if sName == "Stub" && tt.wantStories != nil {
						// This is a placeholder for more detailed comparison if needed for the stub.
						// For now, the length check and non-empty field checks above cover a lot.
						// We might compare tt.wantStories with `got` directly if stub provides deterministic output.
						if len(got) != len(tt.wantStories) {
							t.Errorf("Summarize() for Stub, got %d stories, want %d stories for text: %q", len(got), len(tt.wantStories), tt.text)
						}
						// Further checks for content can be added here if tt.wantStories is defined for the stub.
					}

				}) // End of t.Run for test case
			} // End of loop over test cases
		}) // End of t.Run for summarizer type
	} // End of loop over summarizers
}
