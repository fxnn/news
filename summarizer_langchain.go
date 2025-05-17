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
	URL      string `json:"url" describe:"The primary URL associated with the story, if any. If multiple URLs are relevant, pick the most prominent one. If no URL is present in the story, this can be an empty string."`
}

// StoryListContainer is the top-level structure the LLM is expected to return,
// containing a list of stories.
type StoryListContainer struct {
	Stories []ParsedStory `json:"stories" describe:"A list of identified news stories"`
}

// langChainSummarizer implements our Summarizer interface
type langChainSummarizer struct {
	chain  chains.Chain
	parser schema.OutputParser[StoryListContainer] // Parser for the LLM's structured output
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

	// Create an output parser for the StoryListContainer struct.
	// The LLM will be instructed to return a JSON object matching this structure.
	parser, err := outputparser.NewDefined(StoryListContainer{})
	if err != nil {
		return nil, fmt.Errorf("failed to create output parser: %w", err)
	}
	formatInstructions := parser.GetFormatInstructions()

	// Define the prompt template, incorporating format instructions for JSON output
	// The LLM is instructed to identify multiple stories and format them as JSON.
	promptTemplateString := fmt.Sprintf(`%s

Your task is to identify distinct news stories or topics in the provided text.
For each identified story, provide a concise headline, a brief teaser, and the primary URL associated with the story if one is present.
Format your entire output as a JSON object according to the schema provided above.

IMPORTANT INSTRUCTIONS:
- If the input text contains ANY discernible information, even if it's just a single sentence or a very short statement, you MUST treat it as a story. Provide at least one entry in the "stories" array.
- Do NOT return an empty "stories" array if there is any meaningful content to summarize, no matter how short.
- Only if the input text is completely empty, consists only of whitespace, or is absolute gibberish with no extractable meaning, should you return a JSON object with an empty "stories" array (e.g., {"stories": []}).
- In all other cases, you must populate the "stories" array with at least one story.

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

	// Parse the LLM's JSON output string into a StoryListContainer
	storyContainer, err := s.parser.Parse(llmOutputText)
	if err != nil {
		// Log the problematic text for debugging
		fmt.Printf("Debug: Failed to parse LLM output. Output text was: <<<\n%s\n>>>\n", llmOutputText)
		return nil, fmt.Errorf("failed to parse LLM output into StoryListContainer: %w", err)
	}

	// Convert []ParsedStory from the container to []Story
	var stories []Story
	for _, ps := range storyContainer.Stories {
		stories = append(stories, Story{
			Headline: ps.Headline,
			Teaser:   ps.Teaser,
			URL:      ps.URL, // Map the URL from ParsedStory
		})
	}

	return stories, nil
}
