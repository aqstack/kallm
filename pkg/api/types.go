// Package api provides OpenAI-compatible API types for kallm.
package api

import "time"

// ChatCompletionRequest represents an OpenAI chat completion request.
type ChatCompletionRequest struct {
	Model            string          `json:"model"`
	Messages         []Message       `json:"messages"`
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"top_p,omitempty"`
	N                *int            `json:"n,omitempty"`
	Stream           bool            `json:"stream,omitempty"`
	Stop             []string        `json:"stop,omitempty"`
	MaxTokens        *int            `json:"max_tokens,omitempty"`
	PresencePenalty  *float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64        `json:"frequency_penalty,omitempty"`
	User             string          `json:"user,omitempty"`
	Functions        []Function      `json:"functions,omitempty"`
	FunctionCall     interface{}     `json:"function_call,omitempty"`
	Tools            []Tool          `json:"tools,omitempty"`
	ToolChoice       interface{}     `json:"tool_choice,omitempty"`
	ResponseFormat   *ResponseFormat `json:"response_format,omitempty"`
	Seed             *int            `json:"seed,omitempty"`
}

// Message represents a chat message.
type Message struct {
	Role         string        `json:"role"`
	Content      interface{}   `json:"content"` // string or []ContentPart
	Name         string        `json:"name,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`
	ToolCallID   string        `json:"tool_call_id,omitempty"`
}

// ContentPart represents a multimodal content part.
type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image URL in a content part.
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// Function represents a function definition.
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// FunctionCall represents a function call.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool represents a tool definition.
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// ToolCall represents a tool call.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// ResponseFormat specifies the response format.
type ResponseFormat struct {
	Type string `json:"type"`
}

// ChatCompletionResponse represents an OpenAI chat completion response.
type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
}

// Choice represents a completion choice.
type Choice struct {
	Index        int      `json:"index"`
	Message      Message  `json:"message"`
	FinishReason string   `json:"finish_reason"`
	Logprobs     *Logprob `json:"logprobs,omitempty"`
}

// Logprob represents log probability information.
type Logprob struct {
	Content []TokenLogprob `json:"content,omitempty"`
}

// TokenLogprob represents log probability for a token.
type TokenLogprob struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
}

// Usage represents token usage statistics.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// EmbeddingRequest represents an OpenAI embedding request.
type EmbeddingRequest struct {
	Input          interface{} `json:"input"` // string or []string
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     *int        `json:"dimensions,omitempty"`
	User           string      `json:"user,omitempty"`
}

// EmbeddingResponse represents an OpenAI embedding response.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

// EmbeddingData represents embedding data for a single input.
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingUsage represents token usage for embeddings.
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// APIError represents an API error.
type APIError struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param,omitempty"`
	Code    *string `json:"code,omitempty"`
}

// CacheEntry represents a cached response with metadata.
type CacheEntry struct {
	Request    ChatCompletionRequest  `json:"request"`
	Response   ChatCompletionResponse `json:"response"`
	Embedding  []float64              `json:"embedding"`
	CreatedAt  time.Time              `json:"created_at"`
	ExpiresAt  time.Time              `json:"expires_at"`
	HitCount   int64                  `json:"hit_count"`
	LastHitAt  time.Time              `json:"last_hit_at"`
}

// CacheStats represents cache statistics.
type CacheStats struct {
	TotalEntries   int64   `json:"total_entries"`
	TotalHits      int64   `json:"total_hits"`
	TotalMisses    int64   `json:"total_misses"`
	HitRate        float64 `json:"hit_rate"`
	AvgSimilarity  float64 `json:"avg_similarity"`
	EstimatedSaved float64 `json:"estimated_saved_usd"`
}
