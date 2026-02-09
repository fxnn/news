package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

	handleStories(w, req, tmpDir, "")

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

	handleStories(w, req, tmpDir, "")

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

	handleStories(w, req, "/nonexistent/directory", "")

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

			handleStories(w, req, tmpDir, "")

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

	handleStories(w, req, tmpDir, "")

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

func TestHandleStories_AnnotatesSavedStories(t *testing.T) {
	storydir := t.TempDir()
	savedir := t.TempDir()

	testStories := []story.Story{
		{
			Headline:  "Saved Story",
			Teaser:    "Teaser 1",
			URL:       "https://example.com/1",
			FromEmail: "test@example.com",
			FromName:  "Test",
			Date:      time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		},
		{
			Headline:  "Unsaved Story",
			Teaser:    "Teaser 2",
			URL:       "https://example.com/2",
			FromEmail: "test@example.com",
			FromName:  "Test",
			Date:      time.Date(2006, 1, 3, 15, 4, 5, 0, time.UTC),
		},
	}

	messageID := "<test@example.com>"
	date := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	if err := story.WriteStoriesToDir(storydir, messageID, date, testStories); err != nil {
		t.Fatal(err)
	}

	// Copy only the first story to savedir to mark it as saved
	firstStoryFilename := "2006-01-02_test@example.com_1.json"
	data, err := os.ReadFile(filepath.Join(storydir, firstStoryFilename))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(savedir, firstStoryFilename), data, 0600); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/stories", nil)
	w := httptest.NewRecorder()

	handleStories(w, req, storydir, savedir)

	var stories []storyResponse
	if err := json.NewDecoder(w.Body).Decode(&stories); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(stories) != 2 {
		t.Fatalf("Got %d stories, want 2", len(stories))
	}

	// Find each story and check saved status
	for _, s := range stories {
		if s.Headline == "Saved Story" && !s.Saved {
			t.Error("Saved Story should have saved=true")
		}
		if s.Headline == "Unsaved Story" && s.Saved {
			t.Error("Unsaved Story should have saved=false")
		}
	}
}

func TestHandleSaveStory_Success(t *testing.T) {
	storydir := t.TempDir()
	savedir := t.TempDir()

	content := []byte(`{"headline":"Test"}`)
	if err := os.WriteFile(filepath.Join(storydir, "story.json"), content, 0600); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/stories/story.json/save", nil)
	req.SetPathValue("filename", "story.json")
	w := httptest.NewRecorder()

	handleSaveStory(w, req, storydir, savedir)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusCreated)
	}

	if _, err := os.Stat(filepath.Join(savedir, "story.json")); err != nil {
		t.Errorf("saved file not found: %v", err)
	}
}

func TestHandleSaveStory_NotFound(t *testing.T) {
	storydir := t.TempDir()
	savedir := t.TempDir()

	req := httptest.NewRequest(http.MethodPost, "/api/stories/nonexistent.json/save", nil)
	req.SetPathValue("filename", "nonexistent.json")
	w := httptest.NewRecorder()

	handleSaveStory(w, req, storydir, savedir)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleSaveStory_AlreadySaved(t *testing.T) {
	storydir := t.TempDir()
	savedir := t.TempDir()

	content := []byte(`{"headline":"Test"}`)
	if err := os.WriteFile(filepath.Join(storydir, "story.json"), content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(savedir, "story.json"), content, 0600); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/stories/story.json/save", nil)
	req.SetPathValue("filename", "story.json")
	w := httptest.NewRecorder()

	handleSaveStory(w, req, storydir, savedir)

	if w.Code != http.StatusConflict {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestHandleSaveStory_PathTraversal(t *testing.T) {
	storydir := t.TempDir()
	savedir := t.TempDir()

	req := httptest.NewRequest(http.MethodPost, "/api/stories/../evil.json/save", nil)
	req.SetPathValue("filename", "../evil.json")
	w := httptest.NewRecorder()

	handleSaveStory(w, req, storydir, savedir)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUnsaveStory_Success(t *testing.T) {
	savedir := t.TempDir()

	if err := os.WriteFile(filepath.Join(savedir, "story.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/stories/story.json/save", nil)
	req.SetPathValue("filename", "story.json")
	w := httptest.NewRecorder()

	handleUnsaveStory(w, req, savedir)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}

	if _, err := os.Stat(filepath.Join(savedir, "story.json")); !os.IsNotExist(err) {
		t.Error("file should have been removed")
	}
}

func TestHandleUnsaveStory_NotSaved(t *testing.T) {
	savedir := t.TempDir()

	req := httptest.NewRequest(http.MethodDelete, "/api/stories/nonexistent.json/save", nil)
	req.SetPathValue("filename", "nonexistent.json")
	w := httptest.NewRecorder()

	handleUnsaveStory(w, req, savedir)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleUnsaveStory_PathTraversal(t *testing.T) {
	savedir := t.TempDir()

	req := httptest.NewRequest(http.MethodDelete, "/api/stories/../evil.json/save", nil)
	req.SetPathValue("filename", "../evil.json")
	w := httptest.NewRecorder()

	handleUnsaveStory(w, req, savedir)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
