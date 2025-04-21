package main

import (
	"testing"
)

func TestSummarize(t *testing.T) {
	tests := []struct {
		name    string
		text    string
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
		// Add more test cases here if needed, e.g., for different languages, formats, error conditions.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Summarize(tt.text)

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
