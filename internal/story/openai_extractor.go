package story

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/email"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIExtractor uses OpenAI API to extract stories from emails
type OpenAIExtractor struct {
	client *openai.Client
	model  string
}

// NewOpenAIExtractor creates a new OpenAI-based story extractor
func NewOpenAIExtractor(cfg *config.LLMConfig) *OpenAIExtractor {
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
	Stories []ExtractedStory `json:"stories"`
}

func (e *OpenAIExtractor) Extract(emailData *email.Email) ([]Story, error) {
	prompt := fmt.Sprintf(`Extract news stories from this email. For each story, provide a headline, teaser, and URL.

Subject: %s

Body:
%s

Return a JSON object with this exact structure:
{
  "stories": [
    {
      "headline": "Story headline",
      "teaser": "Brief teaser text (1-2 sentences)",
      "url": "https://example.com/article"
    }
  ]
}

IMPORTANT RULES:
- Write the headline and teaser in the same language as the original email
- Keep headlines SHORT: maximum 5-8 words
- Each story MUST have a unique URL link to the actual article
- If there is only one URL in the email, create only one story
- Separate stories should have separate URLs - do not create multiple stories for a single URL
- Exclude order links, shopping links, or any paid content
- Exclude promotional content or advertisements
- Only include actual news stories or articles with readable content
- If there are no valid stories with URLs, return {"stories": []}
`, emailData.Subject, emailData.Body)

	resp, err := e.client.CreateChatCompletion(
		context.Background(),
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
	var stories []Story
	for _, extracted := range llmResp.Stories {
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
