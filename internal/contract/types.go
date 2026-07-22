// Package contract defines the shared wire types exchanged between the
// renderer, runtime, and MCP workers. Every type here is a stable
// JSON-serializable contract — internal types live in their own packages.
package contract

import (
	"encoding/json"
	"time"
)

// ThreadID is a stable opaque identifier for a conversation thread.
type ThreadID string

// TurnID is a stable opaque identifier for a single agent turn.
type TurnID string

// ItemID is a stable opaque identifier for a single item inside a turn.
type ItemID string

// ToolCallID identifies an in-flight tool invocation.
type ToolCallID string

// ThreadStatus describes the lifecycle state of a thread.
type ThreadStatus string

const (
	ThreadStatusIdle     ThreadStatus = "idle"
	ThreadStatusRunning  ThreadStatus = "running"
	ThreadStatusWaiting  ThreadStatus = "waiting"
	ThreadStatusArchived ThreadStatus = "archived"
)

// TurnStatus describes the lifecycle state of a turn.
type TurnStatus string

const (
	TurnStatusQueued    TurnStatus = "queued"
	TurnStatusRunning   TurnStatus = "running"
	TurnStatusCompleted TurnStatus = "completed"
	TurnStatusFailed    TurnStatus = "failed"
	TurnStatusCancelled TurnStatus = "cancelled"
)

// ItemKind discriminates the union of TurnItem.
type ItemKind string

const (
	ItemKindUserMessage      ItemKind = "user_message"
	ItemKindAssistantText    ItemKind = "assistant_text"
	ItemKindAssistantReason  ItemKind = "assistant_reasoning"
	ItemKindToolCall         ItemKind = "tool_call"
	ItemKindToolResult       ItemKind = "tool_result"
	ItemKindApprovalRequest  ItemKind = "approval_request"
	ItemKindUserInputRequest ItemKind = "user_input_request"
	ItemKindCompaction       ItemKind = "compaction"
	ItemKindReview           ItemKind = "review"
	ItemKindError            ItemKind = "error"
)

// Thread is the top-level conversation container.
type Thread struct {
	ID        ThreadID       `json:"id"`
	Title     string         `json:"title"`
	Status    ThreadStatus   `json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// UserInput is the payload submitted by the renderer to start a turn.
type UserInput struct {
	Text        string          `json:"text"`
	DisplayText string          `json:"displayText,omitempty"`
	Attachments []AttachmentRef `json:"attachments,omitempty"`
}

// Turn is a single agent execution cycle within a thread.
type Turn struct {
	ID        TurnID     `json:"id"`
	ThreadID  ThreadID   `json:"threadId"`
	Status    TurnStatus `json:"status"`
	StartedAt time.Time  `json:"startedAt,omitempty"`
	EndedAt   time.Time  `json:"endedAt,omitempty"`
	Items     []TurnItem `json:"items"`
}

// AttachmentRef references an uploaded attachment by its ID.
type AttachmentRef struct {
	ID       string `json:"id"`
	MIMEType string `json:"mimeType"`
	Name     string `json:"name,omitempty"`
}

// TurnItem is the union of every item kind that can appear in a turn.
// Exactly one of the pointer fields is set, chosen by Kind.
type TurnItem struct {
	ID        ItemID            `json:"id"`
	Kind      ItemKind          `json:"kind"`
	CreatedAt time.Time         `json:"createdAt"`
	Status    string            `json:"status,omitempty"`
	Text      *string           `json:"text,omitempty"`
	ToolCall  *ToolCall         `json:"toolCall,omitempty"`
	Approval  *Approval         `json:"approval,omitempty"`
	UserInput *UserInputRequest `json:"userInput,omitempty"`
	Metadata  map[string]any    `json:"metadata,omitempty"`
}

// ToolCall is a request to invoke a tool and its result.
type ToolCall struct {
	ID     ToolCallID      `json:"id"`
	Name   string          `json:"name"`
	Input  json.RawMessage `json:"input"`
	Output json.RawMessage `json:"output,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// Approval is a request to approve a tool action.
type Approval struct {
	ID        string    `json:"id"`
	ToolName  string    `json:"toolName"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserInputRequest is a request for structured user input mid-turn.
type UserInputRequest struct {
	ID        string              `json:"id"`
	Questions []UserInputQuestion `json:"questions"`
	CreatedAt time.Time           `json:"createdAt"`
}

// UserInputQuestion is a single question with options.
type UserInputQuestion struct {
	ID      string            `json:"id"`
	Prompt  string            `json:"prompt"`
	Options []UserInputOption `json:"options,omitempty"`
}

// UserInputOption is one choice in a user input question.
type UserInputOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
