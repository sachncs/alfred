package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/tool"
)

// TurnLoop drives the agent turn: stream from model, dispatch tool calls,
// repeat until the model stops calling tools.
type TurnLoop struct {
	client model.Client
	tools  map[string]tool.Tool
	store  ThreadStore
	events func(contract.ThreadID, SSEEvent)
}

// NewTurnLoop creates a TurnLoop with the given model client and tools.
func NewTurnLoop(client model.Client, tools []tool.Tool, ts ThreadStore, evFn func(contract.ThreadID, SSEEvent)) *TurnLoop {
	tm := make(map[string]tool.Tool, len(tools))
	for _, t := range tools {
		tm[t.Name()] = t
	}
	return &TurnLoop{client: client, tools: tm, store: ts, events: evFn}
}

// RunTurn executes a complete turn, generating a new turn ID.
func (l *TurnLoop) RunTurn(ctx context.Context, thread *contract.Thread, input contract.UserInput) (*contract.Turn, error) {
	return l.RunTurnWithID(ctx, thread, input, "")
}

// RunTurnWithID executes a turn using the given turn ID. If turnID is empty,
// a new one is generated.
func (l *TurnLoop) RunTurnWithID(ctx context.Context, thread *contract.Thread, input contract.UserInput, turnID contract.TurnID) (*contract.Turn, error) {
	if turnID == "" {
		turnID = contract.TurnID(fmt.Sprintf("turn-%d", time.Now().UnixNano()))
	}
	turn := &contract.Turn{
		ID:        turnID,
		ThreadID:  thread.ID,
		Status:    contract.TurnStatusRunning,
		StartedAt: time.Now().UTC(),
		Items:     []contract.TurnItem{},
	}

	// Append user message.
	userText := input.Text
	if input.DisplayText != "" {
		userText = input.DisplayText
	}
	turn.Items = append(turn.Items, contract.TurnItem{
		ID:        contract.ItemID(fmt.Sprintf("it-%d-user", time.Now().UnixNano())),
		Kind:      contract.ItemKindUserMessage,
		CreatedAt: time.Now().UTC(),
		Text:      &userText,
		Metadata:  map[string]any{"turnId": string(turnID)},
	})

	l.emit(thread.ID, SSEEvent{Event: "turn.started", Data: turn})

	// Build message history for the model.
	messages := []model.Message{
		{Role: "user", Content: userText},
	}

	// Tool-call loop — max 20 iterations to prevent infinite loops.
	for i := 0; i < 20; i++ {
		if ctx.Err() != nil {
			turn.Status = contract.TurnStatusFailed
			return turn, ctx.Err()
		}

		// Convert tools to model ToolSpecs.
		specs := make([]model.ToolSpec, 0, len(l.tools))
		for _, t := range l.tools {
			specs = append(specs, model.ToolSpec{
				Name:        t.Name(),
				Description: t.Description(),
				Schema:      string(t.Schema()),
			})
		}

		req := model.Request{
			Messages: messages,
			Tools:    specs,
		}

		ch, err := l.client.Stream(ctx, req)
		if err != nil {
			turn.Status = contract.TurnStatusFailed
			return turn, err
		}

		var assistantBuf []byte
		var toolCalls []model.StreamChunk

		for chunk := range ch {
			if chunk.Error != "" {
				turn.Status = contract.TurnStatusFailed
				return turn, fmt.Errorf("model error: %s", chunk.Error)
			}
			if chunk.DeltaText != "" {
				assistantBuf = append(assistantBuf, chunk.DeltaText...)
			}
			if chunk.ToolCall != nil {
				toolCalls = append(toolCalls, chunk)
			}
			if chunk.Done {
				break
			}
		}

		// Append assistant text if any.
		if len(assistantBuf) > 0 {
			t := string(assistantBuf)
			turn.Items = append(turn.Items, contract.TurnItem{
				ID:        contract.ItemID(fmt.Sprintf("it-%d-asst", time.Now().UnixNano())),
				Kind:      contract.ItemKindAssistantText,
				CreatedAt: time.Now().UTC(),
				Text:      &t,
				Metadata:  map[string]any{"turnId": string(turnID)},
			})
			messages = append(messages, model.Message{Role: "assistant", Content: t})
		}

		// No tool calls — we're done.
		if len(toolCalls) == 0 {
			turn.Status = contract.TurnStatusCompleted
			turn.EndedAt = time.Now().UTC()
			l.emit(thread.ID, SSEEvent{Event: "turn.completed", Data: turn})
			return turn, nil
		}

		// Dispatch tool calls.
		for _, chunk := range toolCalls {
			tc := chunk.ToolCall
			item := l.dispatchTool(ctx, turnID, *tc)
			turn.Items = append(turn.Items, item)

			// Feed tool result back to model.
			toolResult := ""
			if item.ToolCall != nil && item.ToolCall.Output != nil {
				toolResult = string(item.ToolCall.Output)
			}
			messages = append(messages, model.Message{
				Role:    "tool",
				Content: toolResult,
				Name:    tc.Name,
			})

			l.emit(thread.ID, SSEEvent{Event: "tool.executed", Data: item})
		}
	}

	turn.Status = contract.TurnStatusCompleted
	turn.EndedAt = time.Now().UTC()
	return turn, nil
}

func (l *TurnLoop) dispatchTool(ctx context.Context, turnID contract.TurnID, tc model.ToolCall) contract.TurnItem {
	t, ok := l.tools[tc.Name]
	now := time.Now().UTC()
	itemID := contract.ItemID(fmt.Sprintf("it-%d-tc", time.Now().UnixNano()))

	if !ok {
		return contract.TurnItem{
			ID:        itemID,
			Kind:      contract.ItemKindToolResult,
			CreatedAt: now,
			Metadata:  map[string]any{"turnId": string(turnID)},
			ToolCall: &contract.ToolCall{
				ID:    contract.ToolCallID(tc.ID),
				Name:  tc.Name,
				Input: json.RawMessage(tc.Input),
				Error: fmt.Sprintf("tool not found: %s", tc.Name),
			},
		}
	}

	toolCtx := tool.NewContext(ctx, "", string(turnID), string(contract.ToolCallID(tc.ID)), "")
	res, err := t.Execute(ctx, json.RawMessage(tc.Input), toolCtx)

	item := contract.TurnItem{
		ID:        itemID,
		Kind:      contract.ItemKindToolResult,
		CreatedAt: now,
		Metadata:  map[string]any{"turnId": string(turnID)},
		ToolCall: &contract.ToolCall{
			ID:    contract.ToolCallID(tc.ID),
			Name:  tc.Name,
			Input: json.RawMessage(tc.Input),
		},
	}
	if res != nil {
		b, _ := json.Marshal(res)
		item.ToolCall.Output = b
	}
	if err != nil {
		item.ToolCall.Error = err.Error()
	} else if res != nil && !res.OK {
		item.ToolCall.Error = res.Error
	}
	return item
}

func (l *TurnLoop) emit(threadID contract.ThreadID, ev SSEEvent) {
	if l.events != nil {
		l.events(threadID, ev)
	}
}
