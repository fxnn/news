package main

import (
	"os"
	"testing"
)

func TestSummarizeImplementations(t *testing.T) {
	// Define the test cases once
	testCases := []struct {
		name       string
		text       string
		want    string
		// We don't check for specific want string because LLM output is non-deterministic
		wantErr    error
		checkEmpty bool // Flag to check if the output should be non-empty
	}{
		{
			name:       "Empty text",
			text:       "",
			wantErr:    nil,
			checkEmpty: false, // Expect empty summary for empty text
		},
		{
			name: "Normal text",
			text: "This is a reasonably long email body that requires summarization. " +
				"It discusses the project status, upcoming deadlines, and action items for the team. " +
				"We need to ensure the summary captures the key points without being too verbose.",
			wantErr:    nil,
			checkEmpty: true, // Expect a non-empty summary
		},
		{
			name:       "Short text",
			text:       "OK",
			wantErr:    nil,
			checkEmpty: true, // Expect a non-empty summary even for short text
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

			// If an error was expected, we don't need to check the output string
			if tt.wantErr != nil {
				return
			}

			// Check if the summary should be non-empty
			if tt.checkEmpty && got == "" {
				t.Errorf("Summarize() got an empty summary, want non-empty for text: %q", tt.text)
			}

			// Check if the summary should be empty (only for the empty text case)
			if !tt.checkEmpty && got != "" {
				t.Errorf("Summarize() got = %q, want empty summary for empty text", got)
			}

			// Optional: Check if summary is shorter than original (might be flaky)
			// if tt.checkEmpty && len(got) >= len(tt.text) && len(tt.text) > 50 { // Only check for reasonably long texts
			// 	t.Errorf("Summarize() summary length %d >= original length %d for text: %q", len(got), len(tt.text), tt.text)
			// }
		})
	}
}
