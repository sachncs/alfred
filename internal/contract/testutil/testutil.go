// Package testutil provides shared helpers for building fake domain
// objects in tests. Keep this package dependency-free so any test
// package can import it without circular import issues.
package testutil

import (
	"time"

	"github.com/alfred/alfred/internal/contract"
)

// NewThread returns a fully-populated Thread for tests.
func NewThread(id contract.ThreadID, title string) *contract.Thread {
	now := time.Now().UTC()
	return &contract.Thread{
		ID:        id,
		Title:     title,
		Status:    contract.ThreadStatusIdle,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewTurn returns a turn with the given id, thread, and status.
func NewTurn(id contract.TurnID, threadID contract.ThreadID, status contract.TurnStatus) *contract.Turn {
	return &contract.Turn{
		ID:       id,
		ThreadID: threadID,
		Status:   status,
		Items:    []contract.TurnItem{},
	}
}

// NewUserMessageItem returns a TurnItem of kind user_message.
func NewUserMessageItem(id contract.ItemID, text string) contract.TurnItem {
	t := text
	return contract.TurnItem{
		ID:        id,
		Kind:      contract.ItemKindUserMessage,
		CreatedAt: time.Now().UTC(),
		Text:      &t,
	}
}

// NewAssistantTextItem returns a TurnItem of kind assistant_text.
func NewAssistantTextItem(id contract.ItemID, text string) contract.TurnItem {
	t := text
	return contract.TurnItem{
		ID:        id,
		Kind:      contract.ItemKindAssistantText,
		CreatedAt: time.Now().UTC(),
		Text:      &t,
	}
}

// NewToolResultItem returns a TurnItem of kind tool_result.
func NewToolResultItem(id contract.ItemID, name string, output, inputJSON []byte) contract.TurnItem {
	return contract.TurnItem{
		ID:        id,
		Kind:      contract.ItemKindToolResult,
		CreatedAt: time.Now().UTC(),
		ToolCall: &contract.ToolCall{
			ID:     contract.ToolCallID(id),
			Name:   name,
			Input:  inputJSON,
			Output: output,
		},
	}
}

// NewUserInput returns a UserInput with the given text.
func NewUserInput(text string) contract.UserInput {
	return contract.UserInput{Text: text}
}
