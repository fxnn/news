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
		`Write a concise summary of the following text.
Follow these instructions strictly:
- Use only prose and complete sentences.
- Do not use bullet points or lists.
- Do not mention the format of the original text (e.g., "This is an HTML email").
- Do not include metadata like dates unless they are part of the core narrative.
- Focus solely on extracting the key facts and stories presented in the text.

Text:
"{{.text}}"

CONCISE SUMMARY:`,
		[]string{"text"},
	)

	// Create an LLMChain with the custom prompt
	llmChain := chains.NewLLMChain(llm, prompt)

	// Create the StuffDocuments chain using the LLMChain
	// This chain knows how to handle input documents and use the LLMChain
	stuffChain := chains.NewStuffDocuments(llmChain)

	return &langChainSummarizer{chain: stuffChain}, nil
}

// Summarize calls the underlying langchaingo SummarizeChain.
// It still returns "" for empty input to match your tests.
func (s *langChainSummarizer) Summarize(text string) (string, error) {
	if text == "" {
		return "", nil
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
	// The default output key for NewStuffDocuments is "output_text".
	summary, ok := result["output_text"].(string)
	if !ok {
		// Log the actual result for debugging if the type assertion fails
		fmt.Printf("Debug: Unexpected result type or key. Result map: %v\n", result)
		return "", fmt.Errorf("unexpected output type from summarization chain: expected string under key 'output_text', got %T", result["output_text"])
	}
	return summary, nil
}
