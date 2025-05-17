package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockSummarizer allows mocking the Summarizer interface for tests.
type mockSummarizer struct {
	SummarizeFunc func(text string) ([]Story, error)
}

func (m *mockSummarizer) Summarize(text string) ([]Story, error) {
	if m.SummarizeFunc != nil {
		return m.SummarizeFunc(text)
	}
	// Default behavior: return a single mock story if not customized.
	return []Story{{Headline: "Mock Story", Teaser: "Default mock teaser", URL: "http://example.com/mock"}}, nil
}

// mockEmailFetcher creates a mock EmailFetcher function for testing.
// It allows specifying the emails and error to return.
func mockEmailFetcher(emailsToReturn []Email, errorToReturn error) EmailFetcher {
	return func(server string, port int, username, password, folder string, days int, tls bool) ([]Email, error) {
		return emailsToReturn, errorToReturn
	}
}

func TestStoriesHandler(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	testCfg := config{ // Basic config for tests, specific fields overridden as needed
		server:   "test.imap.com",
		port:     993,
		username: "user",
		password: "password",
		folder:   "INBOX",
		days:     7,
	}

	story1 := Story{Headline: "Story 1", Teaser: "Teaser 1", URL: "http://example.com/1"}
	story2 := Story{Headline: "Story 2", Teaser: "Teaser 2", URL: "http://example.com/2"}

	email1 := Email{UID: 1, Subject: "Email 1", Body: "Body 1", Date: fixedTime, Stories: []Story{story1}}
	email2 := Email{UID: 2, Subject: "Email 2", Body: "Body 2", Date: fixedTime, Stories: []Story{story2}}
	emailWithSummarizationError := Email{UID: 3, Subject: "Email 3", Body: "Body 3", Date: fixedTime, SummarizationError: errors.New("summarization failed")}
	emailNoStories := Email{UID: 4, Subject: "Email 4", Body: "Body 4", Date: fixedTime, Stories: []Story{}}

	tests := []struct {
		name               string
		fetcher            EmailFetcher
		summarizer         Summarizer
		expectedStatusCode int
		expectedBody       string // Expected JSON string
	}{
		{
			name:               "fetch emails error",
			fetcher:            mockEmailFetcher(nil, errors.New("fetch failed")),
			summarizer:         &mockSummarizer{},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Failed to fetch or summarize emails: error fetching emails: fetch failed\n",
		},
		{
			name:               "no emails found",
			fetcher:            mockEmailFetcher([]Email{}, nil),
			summarizer:         &mockSummarizer{},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "[]\n",
		},
		{
			name:    "multiple emails with stories",
			fetcher: mockEmailFetcher([]Email{email1, email2}, nil), // Stories are pre-set in mock emails for simplicity here
			summarizer: &mockSummarizer{SummarizeFunc: func(text string) ([]Story, error) {
				if text == "Body 1" {
					return []Story{story1}, nil
				}
				if text == "Body 2" {
					return []Story{story2}, nil
				}
				return nil, nil
			}},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[{"Headline":"Story 1","Teaser":"Teaser 1","URL":"http://example.com/1"},{"Headline":"Story 2","Teaser":"Teaser 2","URL":"http://example.com/2"}]` + "\n",
		},
		{
			name:    "email with summarization error",
			fetcher: mockEmailFetcher([]Email{email1, emailWithSummarizationError}, nil),
			summarizer: &mockSummarizer{SummarizeFunc: func(text string) ([]Story, error) {
				if text == "Body 1" {
					return []Story{story1}, nil
				}
				if text == "Body 3" { // This email will have SummarizationError set by fetchAndSummarizeEmails
					return nil, errors.New("summarization failed")
				}
				return nil, nil
			}},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[{"Headline":"Story 1","Teaser":"Teaser 1","URL":"http://example.com/1"}]` + "\n",
		},
		{
			name:    "email generates no stories",
			fetcher: mockEmailFetcher([]Email{email1, emailNoStories}, nil),
			summarizer: &mockSummarizer{SummarizeFunc: func(text string) ([]Story, error) {
				if text == "Body 1" {
					return []Story{story1}, nil
				}
				if text == "Body 4" { // This email will have empty Stories set by fetchAndSummarizeEmails
					return []Story{}, nil
				}
				return nil, nil
			}},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[{"Headline":"Story 1","Teaser":"Teaser 1","URL":"http://example.com/1"}]` + "\n",
		},
		{
			name:    "all emails generate no stories",
			fetcher: mockEmailFetcher([]Email{emailNoStories}, nil),
			summarizer: &mockSummarizer{SummarizeFunc: func(text string) ([]Story, error) {
				return []Story{}, nil
			}},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "[]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the handler using the test's fetcher and summarizer
			// Note: newStoriesHandler is defined in main.go (will be added)
			handler := newStoriesHandler(testCfg, tt.summarizer, tt.fetcher)

			req := httptest.NewRequest("GET", "/stories", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v. Body: %s",
					status, tt.expectedStatusCode, rr.Body.String())
			}

			// Normalize newlines for body comparison, as http.Error might add a newline.
			gotBody := strings.ReplaceAll(rr.Body.String(), "\r\n", "\n")
			expectedBody := strings.ReplaceAll(tt.expectedBody, "\r\n", "\n")

			if gotBody != expectedBody {
				t.Errorf("handler returned unexpected body: got \n%q\n want \n%q",
					gotBody, expectedBody)
			}

			if tt.expectedStatusCode == http.StatusOK && rr.Header().Get("Content-Type") != "application/json" {
				t.Errorf("handler returned wrong content type: got %v want application/json", rr.Header().Get("Content-Type"))
			}
		})
	}
}

func TestCreateBodyPreview(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "Empty string",
			body: "",
			want: "",
		},
		{
			name: "Short string",
			body: "Hello world",
			want: "Hello world",
		},
		{
			name: "Exactly 20 chars",
			body: "12345678901234567890",
			want: "12345678901234567890",
		},
		{
			name: "Longer than 20 chars",
			body: "This is a string that is definitely longer than one hundred characters, designed to test the truncation logic effectively. It needs to be long enough.",
			want: "This is a string that is definitely longer than one hundred characters, designed to test the truncat...",
		},
		{
			name: "String with newline",
			body: "First line\nSecond line",
			want: "First line Second line",
		},
		{
			name: "String with carriage return",
			body: "First line\rSecond line",
			want: "First line Second line",
		},
		{
			name: "String with CRLF",
			body: "First line\r\nSecond line",
			want: "First line Second line",
		},
		{
			name: "String with multiple newlines",
			body: "Line 1\nLine 2\nLine 3 is exceptionally long, so long in fact that after replacing newlines with spaces, it will most certainly exceed the one hundred character limit for previews, thereby requiring truncation to be applied by the function under test.",
			want: "Line 1 Line 2 Line 3 is exceptionally long, so long in fact that after replacing newlines with space...", // Adjusted expectation after replacement and truncation
		},
		{
			name: "String with leading/trailing spaces preserved",
			body: "  Leading space ",
			want: "  Leading space ",
		},
		{
			name: "Long string with leading/trailing spaces",
			body: "   This is an extremely long string, much longer than one hundred characters, with leading and trailing spaces. The purpose is to verify that truncation works correctly and preserves leading spaces while cutting off the string at the 100-character mark from the start of actual content.   ",
			want: "   This is an extremely long string, much longer than one hundred characters, with leading and trail...", // Match actual desired output
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createBodyPreview(tt.body); got != tt.want {
				t.Errorf("createBodyPreview() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatEmailDetails(t *testing.T) {
	// Fixed date for consistent test output
	testDate := time.Date(2024, time.May, 15, 10, 30, 0, 0, time.UTC)
	// Sample body for consistent preview
	sampleBody := "This is the body of the email. It might be long, or it might be short."
	expectedPreview := "This is the body of the email. It might be long, or it might be short." // createBodyPreview will handle this

	tests := []struct {
		name  string
		email Email // Assumes Email struct will have Stories []Story and SummarizationError error
		want  string
	}{
		{
			name: "Email with multiple stories",
			email: Email{
				UID:     101,
				Date:    testDate,
				Subject: "Multiple Updates",
				From:    "sender@example.com",
				To:      "receiver@example.com",
				Body:    sampleBody,
				Stories: []Story{
					{Headline: "Story 1", Teaser: "Teaser for story 1.", URL: "http://example.com/story1"},
					{Headline: "Story 2", Teaser: "Teaser for story 2.", URL: "http://example.com/story2"},
				},
				SummarizationError: nil,
			},
			want: `
=== Message ===
UID: 101
Date: 2024-05-15 10:30:00 +0000 UTC
Subject: Multiple Updates
From: sender@example.com
To: receiver@example.com
Body Preview: ` + expectedPreview + `
--- Story 1 ---
Headline: Story 1
Teaser: Teaser for story 1.
URL: http://example.com/story1
--- Story 2 ---
Headline: Story 2
Teaser: Teaser for story 2.
URL: http://example.com/story2
`,
		},
		{
			name: "Email with one story, no URL",
			email: Email{
				UID:     102,
				Date:    testDate,
				Subject: "Single Update",
				From:    "sender@example.com",
				To:      "receiver@example.com",
				Body:    sampleBody,
				Stories: []Story{
					{Headline: "Important News", Teaser: "Just one important thing.", URL: ""},
				},
				SummarizationError: nil,
			},
			want: `
=== Message ===
UID: 102
Date: 2024-05-15 10:30:00 +0000 UTC
Subject: Single Update
From: sender@example.com
To: receiver@example.com
Body Preview: ` + expectedPreview + `
--- Story 1 ---
Headline: Important News
Teaser: Just one important thing.
URL: 
`,
		},
		{
			name: "Email with no stories (successful summarization, empty result)",
			email: Email{
				UID:                103,
				Date:               testDate,
				Subject:            "Empty Summary",
				From:               "sender@example.com",
				To:                 "receiver@example.com",
				Body:               sampleBody,
				Stories:            []Story{},
				SummarizationError: nil,
			},
			want: `
=== Message ===
UID: 103
Date: 2024-05-15 10:30:00 +0000 UTC
Subject: Empty Summary
From: sender@example.com
To: receiver@example.com
Body Preview: ` + expectedPreview + `
[No summary generated]
`,
		},
		{
			name: "Email with summarization error",
			email: Email{
				UID:                104,
				Date:               testDate,
				Subject:            "Failed Summary",
				From:               "sender@example.com",
				To:                 "receiver@example.com",
				Body:               sampleBody,
				Stories:            nil,
				SummarizationError: fmt.Errorf("summarizer timed out"),
			},
			want: `
=== Message ===
UID: 104
Date: 2024-05-15 10:30:00 +0000 UTC
Subject: Failed Summary
From: sender@example.com
To: receiver@example.com
Body Preview: ` + expectedPreview + `
Summarization Error: summarizer timed out
`,
		},
		{
			name: "Email with one story, empty headline",
			email: Email{
				UID:     105,
				Date:    testDate,
				Subject: "Update with no headline",
				From:    "sender@example.com",
				To:      "receiver@example.com",
				Body:    sampleBody,
				Stories: []Story{
					{Headline: "", Teaser: "Teaser for story.", URL: "http://example.com/storyX"},
				},
				SummarizationError: nil,
			},
			want: `
=== Message ===
UID: 105
Date: 2024-05-15 10:30:00 +0000 UTC
Subject: Update with no headline
From: sender@example.com
To: receiver@example.com
Body Preview: ` + expectedPreview + `
--- Story 1 ---
Headline: 
Teaser: Teaser for story.
URL: http://example.com/storyX
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We pass a pointer to email, as formatEmailDetails might expect *Email
			// and to avoid copying if Email struct becomes large.
			got := formatEmailDetails(&tt.email)
			// Normalize newlines and trim leading/trailing whitespace for comparison
			normalize := func(s string) string {
				s = strings.ReplaceAll(s, "\r\n", "\n")
				return strings.TrimSpace(s)
			}
			if normalize(got) != normalize(tt.want) {
				t.Errorf("formatEmailDetails() for %s:\nGOT:\n%s\nWANT:\n%s", tt.name, got, tt.want)
			}
		})
	}
}
