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
	prompt := buildPrompt(emailData.Subject, emailData.Body)

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
			MaxCompletionTokens: 4096, // Allow longer responses for emails with many stories
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
