package main

import (
	"os"
	"testing"
)

func TestSummarizeImplementations(t *testing.T) {
	// Define the test cases once
	testCases := []struct {
		name               string
		text               string
		wantErr            error
		expectedNumStories int  // Expected number of stories in the output slice.
		expectURLInStory   bool // Whether the output stories are expected to have a non-empty URL.
	}{
		{
			name:               "Empty text",
			text:               "",
			wantErr:            nil,
			expectedNumStories: 0,
			expectURLInStory:   false, // No stories, so URL presence is moot.
		},
		{
			name: "Plain text - Single story, with URL in input",
			text: "This is a reasonably long email body that requires summarization. " +
				"It discusses the project status. More info at http://example.com/project-status.",
			wantErr:            nil,
			expectedNumStories: 1, // Both stub and current LLM produce 1 story.
			expectURLInStory:   true,
		},
		{
			name:               "Plain text - Short story, with URL in input",
			text:               "OK. Read more http://example.com/ok.",
			wantErr:            nil,
			expectedNumStories: 1,
			expectURLInStory:   true,
		},
		{
			name:               "Plain text - Single story, no URL in input",
			text:               "This is a simple statement. It stands alone.",
			wantErr:            nil,
			expectedNumStories: 1,
			expectURLInStory:   false,
		},
		{
			name:               "HTML text - Single story, with URL in input",
			text:               "<p>This is <b>HTML</b> content. Learn more <a href='http://example.com/html-story'>here</a>.</p>",
			wantErr:            nil,
			expectedNumStories: 1,
			expectURLInStory:   true,
		},
		{
			name:               "HTML text - Single story, no URL in input",
			text:               "<div><p>Just a piece of HTML. Indeed.</p></div>",
			wantErr:            nil,
			expectedNumStories: 1,
			expectURLInStory:   false,
		},
		{
			name: "Plain text - Multiple stories from newsletter format",
			text: `
First Story Headline

This is the body of the first story. It's a paragraph of text.
Learn more: http://example.com/first-story

Second Story Headline

This is the body of the second story. It also has some interesting content.
No specific URL for this one in the text.

Third Story Headline

And here's the third story. This one has a URL too.
Check it out: http://example.com/third-story
`,
			wantErr:            nil,
			expectedNumStories: 3, // We want to drive the capability to find 3 stories.
			expectURLInStory:   true,
		},
		{
			name: "HTML text - Multiple stories from newsletter format",
			text: `
<div>
  <h1>Main Story Headline HTML</h1>
  <p>This is the first paragraph of the main HTML story. It has some  جذاب content. <a href="http://example.com/main-html">Read more...</a></p>
  <p>Another paragraph for the first story.</p>
</div>
<div>
  <h2>Secondary Story HTML</h2>
  <p>This is the content for the second story in HTML. It might be shorter. Check out <a href="http://example.com/secondary-html">this link</a> for fun.</p>
</div>
`,
			wantErr:            nil,
			expectedNumStories: 2, // We want to drive the capability to find 2 stories from this HTML.
			expectURLInStory:   true,
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

					// Check the number of stories returned.
					if len(got) != tt.expectedNumStories {
						t.Errorf("Summarize() for %s, got %d stories, want %d stories for text: %q. Got: %v", sName, len(got), tt.expectedNumStories, tt.text, got)
						// If the number of stories is not as expected, further checks on story content might be misleading or cause panics.
						// However, if we got more stories than expected (0), but expected some, we can still check the ones we got.
						// If we expected stories but got 0, we should return.
						if tt.expectedNumStories == 0 || len(got) == 0 && tt.expectedNumStories > 0 {
							return
						}
					}

					// If no stories were expected and none were returned, the test passes for this part.
					if tt.expectedNumStories == 0 {
						return
					}

					// If stories are expected, check their content.
					for i, story := range got {
						if story.Headline == "" {
							t.Errorf("Summarize() for %s, story %d has empty Headline for text: %q", sName, i, tt.text)
						}
						if story.Teaser == "" {
							t.Errorf("Summarize() for %s, story %d has empty Teaser for text: %q", sName, i, tt.text)
						}

						hasURL := story.URL != ""
						if tt.expectURLInStory && !hasURL {
							t.Errorf("Summarize() for %s, story %d URL is empty, want non-empty URL for text: %q", sName, i, tt.text)
						}
						if !tt.expectURLInStory && hasURL {
							t.Errorf("Summarize() for %s, story %d URL is %q, want empty URL for text: %q", sName, i, story.URL, tt.text)
						}
					}

				}) // End of t.Run for test case
			} // End of loop over test cases
		}) // End of t.Run for summarizer type
	} // End of loop over summarizers
}
