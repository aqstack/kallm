package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaEmbedder generates embeddings using a local Ollama instance.
type OllamaEmbedder struct {
	baseURL    string
	model      string
	dimensions int
	client     *http.Client
}

// OllamaConfig configures the Ollama embedder.
type OllamaConfig struct {
	BaseURL string
	Model   string
	Timeout time.Duration
}

// ollamaRequest is the request body for Ollama embeddings API.
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// ollamaResponse is the response from Ollama embeddings API.
type ollamaResponse struct {
	Embedding []float64 `json:"embedding"`
}

// NewOllamaEmbedder creates a new Ollama embedder.
func NewOllamaEmbedder(cfg *OllamaConfig) *OllamaEmbedder {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "nomic-embed-text"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Dimensions vary by model
	dimensions := 768 // default for nomic-embed-text
	switch cfg.Model {
	case "nomic-embed-text":
		dimensions = 768
	case "mxbai-embed-large":
		dimensions = 1024
	case "all-minilm":
		dimensions = 384
	}

	return &OllamaEmbedder{
		baseURL:    cfg.BaseURL,
		model:      cfg.Model,
		dimensions: dimensions,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Embed generates an embedding for the given text.
func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	reqBody := ollamaRequest{
		Model:  e.model,
		Prompt: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/api/embeddings", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama error (status %d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(ollamaResp.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return ollamaResp.Embedding, nil
}

// EmbedBatch generates embeddings for multiple texts.
// Ollama doesn't support batch embeddings natively, so we do them sequentially.
func (e *OllamaEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	results := make([][]float64, len(texts))

	for i, text := range texts {
		emb, err := e.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		results[i] = emb
	}

	return results, nil
}

// Dimensions returns the dimensionality of the embeddings.
func (e *OllamaEmbedder) Dimensions() int {
	return e.dimensions
}

// Model returns the model name used for embeddings.
func (e *OllamaEmbedder) Model() string {
	return e.model
}
