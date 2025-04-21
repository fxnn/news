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
	chain *chains.SummarizeChain
}

// NewLangChainSummarizer constructs a Summarizer backed by OpenAI via langchaingo.
func NewLangChainSummarizer() (Summarizer, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("environment variable OPENAI_API_KEY is required")
	}

	// Create an OpenAI LLM client
	llm := openai.New(
		openai.WithAPIKey(apiKey),
		// you can also tune Model, Temperature, MaxTokens, etc:
		// openai.WithModel("gpt-4"),
	)

	// Wrap it in a SummarizeChain
	sumChain := chains.NewSummarizeChain(llm)

	return &langChainSummarizer{chain: sumChain}, nil
}

// Summarize calls the underlying langchaingo SummarizeChain.
// It still returns "" for empty input to match your tests.
func (s *langChainSummarizer) Summarize(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	out, err := s.chain.Run(context.Background(), map[string]any{
		"input": text,
	})
	if err != nil {
		return "", err
	}

	// The chain returns an interface{}, but we know it's a string.
	summary, ok := out.(string)
	if !ok {
		return "", fmt.Errorf("unexpected SummarizeChain output type: %T", out)
	}
	return summary, nil
}

func init() {
	// If we can build the LangChain summarizer, swap it in for the stub.
	if lc, err := NewLangChainSummarizer(); err == nil {
		defaultLLM = lc
	}
}
