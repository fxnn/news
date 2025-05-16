package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

// langChainSummarizer implements our Summarizer interface
type langChainSummarizer struct {
	// Use the generic chains.Chain interface
	chain chains.Chain
}

// NewLangChainSummarizer constructs a Summarizer backed by OpenAI via langchaingo.
func NewLangChainSummarizer() (Summarizer, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("environment variable OPENAI_API_KEY is required")
	}

	// Create an OpenAI LLM client
	// Note: openai.New returns (llm, error)
	llm, err := openai.New(
		openai.WithToken(apiKey), // Use WithToken for the API key
		// you can also tune Model, Temperature, MaxTokens, etc:
		openai.WithModel("gpt-4o-mini"), // Specify the desired model
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	// Define the custom prompt template
	// Note: The input variable must be "text" for chains.NewStuffDocuments by default.
	prompt := prompts.NewPromptTemplate(
		`Create a very brief teaser highlighting the key insights from the following text.
Follow these instructions strictly:
- Use only prose and complete sentences.
- Do not use bullet points or lists.
- Do not mention the format of the original text (e.g., "This is an HTML email").
- Do not include any metadata like dates or user names unless they are essential to the core insight.
- Focus solely on the most important facts or stories. Keep it extremely short.
- Write the teaser in the same language as the original text.

Text:
"{{.text}}"

BRIEF TEASER:`,
		[]string{"text"},
	)

	// Create an LLMChain with the custom prompt
	llmChain := chains.NewLLMChain(llm, prompt)

	// Create the StuffDocuments chain using the LLMChain
	// This chain knows how to handle input documents and use the LLMChain
	stuffChain := chains.NewStuffDocuments(llmChain)
	// Ensure the variable name used by StuffDocuments matches the LLMChain's prompt input variable.
	// Although "text" is the default, we set it explicitly for clarity and robustness.
	stuffChain.DocumentVariableName = "text"

	return &langChainSummarizer{chain: stuffChain}, nil
}

// Summarize calls the underlying langchaingo chain.
// It returns nil for empty input. For non-empty input, it returns a single Story
// where the Teaser is the summarized text.
func (s *langChainSummarizer) Summarize(text string) ([]Story, error) {
	if text == "" {
		return nil, nil
	}

	// Prepare the input for the StuffDocuments chain
	// The default input key is "input_documents", expecting a slice of schema.Document.
	// The default output key is "output_text".
	docs := []schema.Document{
		{PageContent: text},
	}
	input := map[string]any{
		"input_documents": docs,
	}

	// Call the chain
	result, err := chains.Call(context.Background(), s.chain, input)
	if err != nil {
		return "", fmt.Errorf("summarization chain call failed: %w", err)
	}

	// Extract the summary string from the result map
	// The output key from the underlying LLMChain is "text".
	summaryText, ok := result["text"].(string)
	if !ok {
		// Log the actual result for debugging if the type assertion fails
		fmt.Printf("Debug: Unexpected result type or key. Result map: %v\n", result)
		return nil, fmt.Errorf("unexpected output type from summarization chain: expected string under key 'text', got %T", result["text"])
	}

	// For now, assume the LLM returns a single story's teaser.
	// Headline and URL would need more sophisticated extraction or prompting.
	story := Story{
		Headline: "Summary", // Placeholder headline
		Teaser:   summaryText,
		URL:      "", // Placeholder URL, not extracted by current prompt
	}
	return []Story{story}, nil
}
