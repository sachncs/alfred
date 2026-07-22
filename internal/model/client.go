// Package model defines the model.Client interface used by ChatAgent.
// Phase 1 ships a stub; the real SSE/Responses/Messages adapters land
// in Phase 5 (model-router MCP worker).
package model

import (
	"context"
	"errors"
)

// Request is a single model invocation.
type Request struct {
	SystemPrompt string     `json:"systemPrompt"`
	Messages     []Message  `json:"messages"`
	Tools        []ToolSpec `json:"tools,omitempty"`
	Model        string     `json:"model,omitempty"`
}

// Message is one entry in the conversation history.
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant", "tool"
	Content string `json:"content"`
	// Name is set when Role=="tool" to identify the originating tool.
	Name string `json:"name,omitempty"`
}

// ToolSpec declares a tool the model may invoke.
type ToolSpec struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Schema is the JSON Schema for the tool's input.
	Schema string `json:"schema"`
}

// StreamChunk is one event from a streaming model response.
type StreamChunk struct {
	// DeltaText is the incremental text added on this chunk (assistant text).
	DeltaText string `json:"deltaText,omitempty"`
	// ToolCall is set when the model wants to invoke a tool.
	ToolCall *ToolCall `json:"toolCall,omitempty"`
	// Done is true on the final chunk of the stream.
	Done bool `json:"done,omitempty"`
	// Error is set if the stream terminated abnormally.
	Error string `json:"error,omitempty"`
}

// ToolCall is a model-emitted tool invocation.
type ToolCall struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Input []byte `json:"input"` // raw JSON
}

// Client is the contract every model backend (OpenAI, Anthropic, stub)
// must satisfy.
type Client interface {
	// Stream sends Request and returns a channel of StreamChunks. The
	// channel MUST be closed by the implementation when done. The
	// implementation MUST honour context cancellation.
	Stream(ctx context.Context, req Request) (<-chan StreamChunk, error)
}

// ErrNotImplemented is returned by stub clients when a feature is not
// implemented yet.
var ErrNotImplemented = errors.New("model: not implemented")
