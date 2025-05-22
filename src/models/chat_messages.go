package models

import "time"

// OllamaChatMessage represents a single message in an Ollama chat request.
type OllamaChatMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"` // For multimodal support if needed in future
}

// OllamaChatRequest represents the request body for Ollama's /api/chat.
type OllamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []OllamaChatMessage `json:"messages"`
	Stream   bool                `json:"stream,omitempty"`
	// Options map[string]interface{} `json:"options,omitempty"` // OpenAI doesn't use this directly
}

// OpenAIChatMessage matches the structure for messages in OpenAI API.
type OpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIChatRequest matches the request structure for OpenAI's /v1/chat/completions.
type OpenAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []OpenAIChatMessage `json:"messages"`
	Stream   bool                `json:"stream,omitempty"`
	// Other OpenAI specific parameters like temperature, max_tokens etc. can be added if needed.
}

// OpenAIChatChoice represents a choice in an OpenAI chat response.
type OpenAIChatChoice struct {
	Index        int               `json:"index"`
	Message      OpenAIChatMessage `json:"message,omitempty"` // For non-streaming
	Delta        OpenAIChatMessage `json:"delta,omitempty"`   // For streaming
	FinishReason string            `json:"finish_reason"`
}

// OpenAIChatResponse is for non-streaming responses.
type OpenAIChatResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []OpenAIChatChoice `json:"choices"`
	// Usage   OpenAIUsage        `json:"usage"` // Optional: if usage data is needed
}

// OpenAIStreamChunk is for streaming responses (the structure of the data part of an SSE).
type OpenAIStreamChunk struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []OpenAIChatChoice `json:"choices"` // Delta will be populated here
}

// OllamaChatResponse represents a non-streaming response from Ollama.
// This structure is based on the example final chunk when not streaming,
// but for a single aggregated response.
type OllamaChatResponse struct {
	Model     string            `json:"model"`
	CreatedAt string            `json:"created_at"`
	Message   OllamaChatMessage `json:"message"` // The complete assistant message
	Done      bool              `json:"done"`
	// Optional fields like total_duration, etc., if they can be obtained or are relevant.
	// For simplicity, we'll focus on the core message.
}

// OllamaStreamChunk represents a streaming chunk in Ollama format.
type OllamaStreamChunk struct {
	Model     string            `json:"model"`
	CreatedAt string            `json:"created_at"`
	Message   OllamaChatMessage `json:"message"` // Contains the delta content
	Done      bool              `json:"done"`    // False until the last chunk
	// Optional: DoneReason string `json:"done_reason,omitempty"`
}
