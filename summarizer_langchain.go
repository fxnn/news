package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
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
		// openai.WithModel("gpt-4"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	// Use LoadStuffSummarizationChain to create the chain
	sumChain := chains.LoadStuffSummarizationChain(llm)

	return &langChainSummarizer{chain: sumChain}, nil
}

// Summarize calls the underlying langchaingo SummarizeChain.
// It still returns "" for empty input to match your tests.
func (s *langChainSummarizer) Summarize(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Use chains.Call which expects and returns map[string]any
	// The input key for summarization chains is typically "text"
	// The output key is typically "text" as well
	result, err := chains.Call(context.Background(), s.chain, map[string]any{
		"text": text, // Use "text" as the input key
	})
	if err != nil {
		return "", fmt.Errorf("summarization chain call failed: %w", err)
	}

	// Extract the summary string from the result map
	summary, ok := result["text"].(string)
	if !ok {
		// Log the actual result for debugging if the type assertion fails
		fmt.Printf("Debug: Unexpected result type or key. Result map: %v\n", result)
		return "", fmt.Errorf("unexpected output type from summarization chain: expected string under key 'text', got %T", result["text"])
	}
	return summary, nil
}

func init() {
	// If we can build the LangChain summarizer, swap it in for the stub.
	if lc, err := NewLangChainSummarizer(); err == nil {
		defaultLLM = lc
	}
}
