// Package ollama provides a minimal streaming client for the Ollama chat API.
package ollama

import "encoding/json"

// Message is a single chat message exchanged with the model.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the payload sent to POST /api/chat.
type chatRequest struct {
	Model    string          `json:"model"`
	Messages []Message       `json:"messages"`
	Stream   bool            `json:"stream"`
	Format   json.RawMessage `json:"format,omitempty"`
	Options  map[string]any  `json:"options,omitempty"`
}

// chatResponse is a single streamed chunk from /api/chat.
type chatResponse struct {
	Model     string  `json:"model"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
	DoneReasn string  `json:"done_reason,omitempty"`
	Error     string  `json:"error,omitempty"`
}
