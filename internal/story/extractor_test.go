package story

import (
	"testing"
	"time"

	"github.com/fxnn/news/internal/email"
)

func TestExtractor_Extract_SingleStory(t *testing.T) {
	extractor := &StubExtractor{
		Stories: []ExtractedStory{
			{
				Headline: "Test Headline",
				Teaser:   "Test teaser text",
				URL:      "https://example.com/article",
			},
		},
	}

	emailData := &email.Email{
		Subject:   "Newsletter Subject",
		Body:      "Newsletter body content",
		FromEmail: "sender@example.com",
		FromName:  "John Doe",
		Date:      time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		MessageID: "<test123@example.com>",
	}

	stories, err := extractor.Extract(emailData)
	if err != nil {
		t.Fatalf("Extract() unexpected error: %v", err)
	}

	if len(stories) != 1 {
		t.Fatalf("Extract() returned %d stories, want 1", len(stories))
	}

	story := stories[0]

	if story.Headline != "Test Headline" {
		t.Errorf("Headline = %v, want Test Headline", story.Headline)
	}

	if story.Teaser != "Test teaser text" {
		t.Errorf("Teaser = %v, want Test teaser text", story.Teaser)
	}

	if story.URL != "https://example.com/article" {
		t.Errorf("URL = %v, want https://example.com/article", story.URL)
	}

	if story.FromEmail != "sender@example.com" {
		t.Errorf("FromEmail = %v, want sender@example.com", story.FromEmail)
	}

	if story.FromName != "John Doe" {
		t.Errorf("FromName = %v, want John Doe", story.FromName)
	}

	expectedDate := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	if !story.Date.Equal(expectedDate) {
		t.Errorf("Date = %v, want %v", story.Date, expectedDate)
	}
}

func TestExtractor_Extract_MultipleStories(t *testing.T) {
	extractor := &StubExtractor{
		Stories: []ExtractedStory{
			{
				Headline: "Story 1",
				Teaser:   "Teaser 1",
				URL:      "https://example.com/1",
			},
			{
				Headline: "Story 2",
				Teaser:   "Teaser 2",
				URL:      "https://example.com/2",
			},
		},
	}

	emailData := &email.Email{
		Subject:   "Newsletter",
		Body:      "Body",
		FromEmail: "sender@example.com",
		FromName:  "Sender",
		Date:      time.Now(),
		MessageID: "<test@example.com>",
	}

	stories, err := extractor.Extract(emailData)
	if err != nil {
		t.Fatalf("Extract() unexpected error: %v", err)
	}

	if len(stories) != 2 {
		t.Fatalf("Extract() returned %d stories, want 2", len(stories))
	}
}

func TestExtractor_Extract_NoStories(t *testing.T) {
	extractor := &StubExtractor{
		Stories: []ExtractedStory{},
	}

	emailData := &email.Email{
		Subject:   "Newsletter",
		Body:      "Body",
		FromEmail: "sender@example.com",
		FromName:  "Sender",
		Date:      time.Now(),
		MessageID: "<test@example.com>",
	}

	stories, err := extractor.Extract(emailData)
	if err != nil {
		t.Fatalf("Extract() unexpected error: %v", err)
	}

	if len(stories) != 0 {
		t.Errorf("Extract() returned %d stories, want 0", len(stories))
	}
}
