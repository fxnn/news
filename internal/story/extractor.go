package story

import "github.com/fxnn/news/internal/email"

// ExtractedStory represents a story extracted by the LLM (without email metadata)
type ExtractedStory struct {
	Headline string
	Teaser   string
	URL      string
}

// Extractor extracts stories from email content using an LLM
type Extractor interface {
	Extract(email *email.Email) ([]Story, error)
}

// StubExtractor is a test implementation that returns predefined stories
type StubExtractor struct {
	Stories []ExtractedStory
}

func (s *StubExtractor) Extract(emailData *email.Email) ([]Story, error) {
	var stories []Story

	for _, extracted := range s.Stories {
		story := Story{
			Headline:  extracted.Headline,
			Teaser:    extracted.Teaser,
			URL:       extracted.URL,
			FromEmail: emailData.FromEmail,
			FromName:  emailData.FromName,
			Date:      emailData.Date,
		}
		stories = append(stories, story)
	}

	return stories, nil
}
