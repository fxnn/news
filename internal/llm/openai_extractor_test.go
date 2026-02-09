package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fxnn/news/internal/config"
	"github.com/fxnn/news/internal/email"
)

func newFakeOpenAIServer(t *testing.T, bodyCh chan<- map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if bodyCh != nil {
			bodyCh <- body
		}
		if err := json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": `{"stories":[]}`}},
			},
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
}

func TestExtract_UsesMaxCompletionTokens(t *testing.T) {
	bodyCh := make(chan map[string]any, 1)
	server := newFakeOpenAIServer(t, bodyCh)
	defer server.Close()

	cfg := &config.LLM{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-4o",
	}
	extractor := NewOpenAIExtractor(cfg)

	_, err := extractor.Extract(&email.Email{
		Subject: "Test",
		Body:    "Test body",
	})
	if err != nil {
		t.Fatalf("Extract() unexpected error: %v", err)
	}

	requestBody := <-bodyCh
	if _, ok := requestBody["max_tokens"]; ok {
		t.Error("request body contains deprecated 'max_tokens' field, should use 'max_completion_tokens' instead")
	}
	if _, ok := requestBody["max_completion_tokens"]; !ok {
		t.Error("request body missing 'max_completion_tokens' field")
	}
}

func TestExtract_WorksWithReasoningModels(t *testing.T) {
	server := newFakeOpenAIServer(t, nil)
	defer server.Close()

	cfg := &config.LLM{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-5-mini",
	}
	extractor := NewOpenAIExtractor(cfg)

	_, err := extractor.Extract(&email.Email{
		Subject: "Test",
		Body:    "Test body",
	})
	if err != nil {
		t.Fatalf("Extract() with gpt-5-mini should succeed, got error: %v", err)
	}
}
