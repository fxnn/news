package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

// ParsedStory defines the structure we expect the LLM to return for each story.
// Tags are for the outputparser.Defined.
type ParsedStory struct {
	Headline string `json:"headline" describe:"The headline of the story"`
	Teaser   string `json:"teaser" describe:"A brief teaser for the story"`
}

// langChainSummarizer implements our Summarizer interface
type langChainSummarizer struct {
	chain  chains.Chain
	parser schema.OutputParser[[]ParsedStory] // Parser for the LLM's structured output
}

// NewLangChainSummarizer constructs a Summarizer backed by OpenAI via langchaingo.
// It now attempts to extract multiple stories in JSON format.
func NewLangChainSummarizer() (Summarizer, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("environment variable OPENAI_API_KEY is required")
	}

	llm, err := openai.New(
		openai.WithToken(apiKey),
		openai.WithModel("gpt-4o-mini"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	// Create an output parser for a slice of ParsedStory structs
	parser, err := outputparser.NewDefined([]ParsedStory{})
	if err != nil {
		return nil, fmt.Errorf("failed to create output parser: %w", err)
	}
	formatInstructions := parser.GetFormatInstructions()

	// Define the prompt template, incorporating format instructions for JSON output
	// The LLM is instructed to identify multiple stories and format them as JSON.
	promptTemplateString := fmt.Sprintf(`%s

Identify distinct news stories or topics in the following text. For each story, provide a concise headline and a brief teaser.
Output all identified stories according to the JSON schema provided above.
If no distinct stories are found, or the text is too short to summarize, return an empty JSON array [].

Text:
"{{.text}}"

JSON Output:`, formatInstructions)

	prompt := prompts.NewPromptTemplate(promptTemplateString, []string{"text"})
	llmChain := chains.NewLLMChain(llm, prompt)
	stuffChain := chains.NewStuffDocuments(llmChain)
	stuffChain.DocumentVariableName = "text" // Ensure StuffDocuments uses "text" for the LLMChain

	return &langChainSummarizer{chain: stuffChain, parser: parser}, nil
}

// Summarize calls the underlying langchaingo chain and parses the structured output.
func (s *langChainSummarizer) Summarize(text string) ([]Story, error) {
	if text == "" {
		return nil, nil
	}

	docs := []schema.Document{{PageContent: text}}
	input := map[string]any{"input_documents": docs}

	result, err := chains.Call(context.Background(), s.chain, input)
	if err != nil {
		return nil, fmt.Errorf("summarization chain call failed: %w", err)
	}

	llmOutputText, ok := result["text"].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected output type from summarization chain: expected string under key 'text', got %T. Full result: %v", result["text"], result)
	}

	// Parse the LLM's JSON output string into []ParsedStory
	parsedLLMStories, err := s.parser.Parse(llmOutputText)
	if err != nil {
		// Log the problematic text for debugging
		fmt.Printf("Debug: Failed to parse LLM output. Output text: %s\n", llmOutputText)
		return nil, fmt.Errorf("failed to parse LLM output into structured stories: %w", err)
	}

	// Convert []ParsedStory to []Story
	var stories []Story
	for _, ps := range parsedLLMStories {
		stories = append(stories, Story{
			Headline: ps.Headline,
			Teaser:   ps.Teaser,
			URL:      "", // URL extraction is not part of this prompt
		})
	}

	return stories, nil
}
