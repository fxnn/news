package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fxnn/news/internal/story"
)

func TestHandleStories_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test stories
	testStories := []story.Story{
		{
			Headline:  "Test Story 1",
			Teaser:    "Teaser 1",
			URL:       "https://example.com/1",
			FromEmail: "test@example.com",
			FromName:  "Test User",
			Date:      time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		},
		{
			Headline:  "Test Story 2",
			Teaser:    "Teaser 2",
			URL:       "https://example.com/2",
			FromEmail: "test@example.com",
			FromName:  "Test User",
			Date:      time.Date(2006, 1, 3, 15, 4, 5, 0, time.UTC),
		},
	}

	messageID := "<test@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	if err := story.WriteStoriesToDir(tmpDir, messageID, date, testStories); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/stories", nil)
	w := httptest.NewRecorder()

	handleStories(w, req, tmpDir)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var stories []story.Story
	if err := json.NewDecoder(resp.Body).Decode(&stories); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(stories) != 2 {
		t.Errorf("Got %d stories, want 2", len(stories))
	}
}

func TestHandleStories_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	req := httptest.NewRequest(http.MethodGet, "/api/stories", nil)
	w := httptest.NewRecorder()

	handleStories(w, req, tmpDir)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var stories []story.Story
	if err := json.NewDecoder(resp.Body).Decode(&stories); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(stories) != 0 {
		t.Errorf("Got %d stories, want 0", len(stories))
	}
}

func TestHandleStories_NonExistentDirectory(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/stories", nil)
	w := httptest.NewRecorder()

	handleStories(w, req, "/nonexistent/directory")

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestHandleStories_MethodNotAllowed(t *testing.T) {
	tmpDir := t.TempDir()

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/stories", nil)
			w := httptest.NewRecorder()

			handleStories(w, req, tmpDir)

			resp := w.Result()
			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("Status = %d, want %d for method %s", resp.StatusCode, http.StatusMethodNotAllowed, method)
			}
		})
	}
}

func TestHandleStories_SortedByDate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create stories with different dates
	olderStory := []story.Story{
		{
			Headline:  "Older Story",
			Teaser:    "Older teaser",
			URL:       "https://example.com/old",
			FromEmail: "test@example.com",
			FromName:  "Test",
			Date:      time.Date(2006, 1, 1, 15, 4, 5, 0, time.UTC),
		},
	}

	newerStory := []story.Story{
		{
			Headline:  "Newer Story",
			Teaser:    "Newer teaser",
			URL:       "https://example.com/new",
			FromEmail: "test@example.com",
			FromName:  "Test",
			Date:      time.Date(2006, 1, 5, 15, 4, 5, 0, time.UTC),
		},
	}

	// Write older story first
	if err := story.WriteStoriesToDir(tmpDir, "<old@example.com>", olderStory[0].Date, olderStory); err != nil {
		t.Fatal(err)
	}

	// Write newer story
	if err := story.WriteStoriesToDir(tmpDir, "<new@example.com>", newerStory[0].Date, newerStory); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/stories", nil)
	w := httptest.NewRecorder()

	handleStories(w, req, tmpDir)

	var stories []story.Story
	if err := json.NewDecoder(w.Body).Decode(&stories); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(stories) != 2 {
		t.Fatalf("Got %d stories, want 2", len(stories))
	}

	// Verify newest first
	if stories[0].Headline != "Newer Story" {
		t.Errorf("First story headline = %q, want %q", stories[0].Headline, "Newer Story")
	}

	if stories[1].Headline != "Older Story" {
		t.Errorf("Second story headline = %q, want %q", stories[1].Headline, "Older Story")
	}
}
