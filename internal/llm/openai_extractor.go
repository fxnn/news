package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/email"
	"github.com/fxnn/news/internal/story"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIExtractor uses OpenAI API to extract stories from emails
type OpenAIExtractor struct {
	client *openai.Client
	model  string
}

// NewOpenAIExtractor creates a new OpenAI-based story extractor
func NewOpenAIExtractor(cfg *config.LLM) *OpenAIExtractor {
	clientConfig := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		clientConfig.BaseURL = cfg.BaseURL
	}

	return &OpenAIExtractor{
		client: openai.NewClientWithConfig(clientConfig),
		model:  cfg.Model,
	}
}

// LLMResponse represents the JSON structure returned by the LLM
type LLMResponse struct {
	Stories []story.ExtractedStory `json:"stories"`
}

func (e *OpenAIExtractor) Extract(emailData *email.Email) ([]story.Story, error) {
	prompt := fmt.Sprintf(`Your task is to extract ALL news stories from this newsletter email. Read through the ENTIRE email body carefully and extract EVERY story that has a URL.

Subject: %s

Body:
%s

Return a JSON object with this exact structure:
{
  "stories": [
    {
      "headline": "Story headline",
      "teaser": "Content type prefix followed by summary (2-4 sentences)",
      "url": "https://example.com/article"
    }
  ]
}

CRITICAL INSTRUCTIONS:
- Extract ALL stories from the email body - not just the ones mentioned in the subject line
- Read through the ENTIRE email systematically from top to bottom
- Each story with a unique URL should be included
- Do NOT limit yourself to only a few stories - extract as many as exist

FORMATTING RULES:
- Write the headline and teaser in the same language as the original email
- Keep headlines SHORT: maximum 5-8 words
- Always start the teaser with a short content type label (1-2 words) followed by a period, e.g. "Article.", "Blog post.", "Podcast.", "Video.", "LinkedIn Post.", "GitHub Repo.", "Research Paper.", "News.", "Tutorial.", "Talk.", "Tool."
- If the newsletter already contains a summary paragraph describing the linked content, reuse that summary word-for-word after the content type prefix, regardless of length
- Otherwise, write teasers of 2-4 sentences. Prefer longer, more informative summaries over short ones.
- Each story MUST have a unique URL link to the actual article
- If there is only one URL in the email, create only one story
- Separate stories should have separate URLs - do not create multiple stories for a single URL

WHAT TO EXTRACT:
- Each story should be a MAIN article/post/resource being featured in the newsletter
- Extract the primary link for each distinct story/article
- Stories are typically presented as separate entries with their own headline and description

EXCLUSION RULES:
- Exclude order links, shopping links, or any paid content
- Exclude sponsored content, advertisements, and promotions. A story is sponsored when it is LABELED as such by the newsletter — look for markers like "(Sponsor)", "Sponsored", "Ad", "Partner Post", "Promoted", "Brought to you by", "In partnership with" used as labels near the headline or as section headers. Do NOT exclude articles that merely discuss topics like advertising, partnerships, or affiliate programs as editorial content
- Exclude anything about the newsletter service itself rather than actual news or content. This includes: subscription management, privacy policies, terms of service, cookie notices, data protection info (Datenschutzinformationen), objection to advertising (Widerspruch gegen Werbung), Impressum, unsubscribe links, contact/about pages of the newsletter, and any similar boilerplate
- Exclude social media links (follow us, share, tweet)
- Exclude footnote links, reference links, and citation links within story text
- Exclude "read more", "learn more", or supplementary links that are part of an existing story
- Only include actual news stories or articles with readable content
- If there are no valid stories with URLs, return {"stories": []}
`, emailData.Subject, emailData.Body)

	// Create context with 60 second timeout to prevent indefinite hangs
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := e.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: e.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
			MaxTokens:   4096, // Allow longer responses for emails with many stories
			Temperature: 0.3,  // Low temperature for consistent, focused extraction
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI API")
	}

	var llmResp LLMResponse
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &llmResp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Convert extracted stories to full stories with email metadata
	var stories []story.Story
	for _, extracted := range llmResp.Stories {
		s := story.Story{
			Headline:  extracted.Headline,
			Teaser:    extracted.Teaser,
			URL:       extracted.URL,
			FromEmail: emailData.FromEmail,
			FromName:  emailData.FromName,
			Date:      emailData.Date,
		}
		stories = append(stories, s)
	}

	return stories, nil
}
