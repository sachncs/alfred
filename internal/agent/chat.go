package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/tool"
)

// ChatAgent is the default chat agent: it streams from a model.Client,
// invokes registered tools, and assembles a Turn.
//
// Phase 1 implements a single-pass loop (no compaction, no sub-agents).
// Phase 2 wires the full turn loop; this is the hello-world.
type ChatAgent struct {
	id           string
	state        *AgentState
	client       model.Client
	toolsMu      sync.RWMutex
	tools        map[string]tool.Tool
	systemPrompt string
}

// NewChatAgent constructs a ChatAgent with the given id, system prompt,
// and model client. No tools are registered initially.
func NewChatAgent(id, systemPrompt string, client model.Client) *ChatAgent {
	return &ChatAgent{
		id:           id,
		state:        NewAgentState(),
		client:       client,
		tools:        map[string]tool.Tool{},
		systemPrompt: systemPrompt,
	}
}

// ID implements Agent.
func (a *ChatAgent) ID() string { return a.id }

// State returns the agent's state (mostly for tests).
func (a *ChatAgent) State() *AgentState { return a.state }

// RegisterTool implements Agent.
func (a *ChatAgent) RegisterTool(t tool.Tool) {
	a.toolsMu.Lock()
	defer a.toolsMu.Unlock()
	a.tools[t.Name()] = t
}

// Tools implements Agent.
func (a *ChatAgent) Tools() []tool.Tool {
	a.toolsMu.RLock()
	defer a.toolsMu.RUnlock()
	out := make([]tool.Tool, 0, len(a.tools))
	for _, t := range a.tools {
		out = append(out, t)
	}
	return out
}

// Interrupt implements Agent.
func (a *ChatAgent) Interrupt(threadID contract.ThreadID) error {
	a.state.mu.Lock()
	defer a.state.mu.Unlock()
	if a.state.CurrentThreadID != threadID {
		return fmt.Errorf("no active turn on thread %s", threadID)
	}
	a.state.LastError = fmt.Errorf("interrupted")
	a.state.CurrentThreadID = ""
	a.state.CurrentTurnID = ""
	return nil
}

// Resume implements Agent.
func (a *ChatAgent) Resume(threadID contract.ThreadID) (*contract.Turn, error) {
	if threadID == "" {
		return nil, errors.New("empty thread id")
	}
	return &contract.Turn{
		ID:       contract.TurnID("resumed-1"),
		ThreadID: threadID,
		Status:   contract.TurnStatusCompleted,
	}, nil
}

// StartTurn implements Agent. It opens a stream from the model.Client,
// forwards the deltas as TurnItems, executes any tool calls, and
// returns the final Turn.
func (a *ChatAgent) StartTurn(ctx context.Context, thread *contract.Thread, input contract.UserInput) (*contract.Turn, error) {
	if thread == nil {
		return nil, errors.New("nil thread")
	}
	turnID := contract.TurnID(fmt.Sprintf("turn-%d", time.Now().UnixNano()))
	a.state.BeginTurn(thread.ID, turnID)

	turn := &contract.Turn{
		ID:       turnID,
		ThreadID: thread.ID,
		Status:   contract.TurnStatusRunning,
		Items:    []contract.TurnItem{},
	}

	// Append the user message as the first item.
	userText := input.Text
	if input.DisplayText != "" {
		userText = input.DisplayText
	}
	turn.Items = append(turn.Items, contract.TurnItem{
		ID:        contract.ItemID(fmt.Sprintf("it-%d-user", time.Now().UnixNano())),
		Kind:      contract.ItemKindUserMessage,
		CreatedAt: time.Now().UTC(),
		Text:      &userText,
	})

	req := model.Request{
		SystemPrompt: a.systemPrompt,
		Messages: []model.Message{
			{Role: "user", Content: userText},
		},
		Model: string(thread.ID),
	}

	ch, err := a.client.Stream(ctx, req)
	if err != nil {
		a.state.EndTurn(turnID, err)
		turn.Status = contract.TurnStatusFailed
		return turn, err
	}

	// Read the stream, accumulate text, dispatch tool calls.
	var assistantBuf []byte
	for chunk := range ch {
		if chunk.Error != "" {
			a.state.EndTurn(turnID, errors.New(chunk.Error))
			turn.Status = contract.TurnStatusFailed
			return turn, errors.New(chunk.Error)
		}
		if chunk.DeltaText != "" {
			assistantBuf = append(assistantBuf, chunk.DeltaText...)
		}
		if chunk.ToolCall != nil {
			// Append the assistant message so far, then the tool call.
			if len(assistantBuf) > 0 {
				t := string(assistantBuf)
				turn.Items = append(turn.Items, contract.TurnItem{
					ID:        contract.ItemID(fmt.Sprintf("it-%d-asst", time.Now().UnixNano())),
					Kind:      contract.ItemKindAssistantText,
					CreatedAt: time.Now().UTC(),
					Text:      &t,
				})
				assistantBuf = assistantBuf[:0]
			}
			tc := *chunk.ToolCall
			item := a.dispatchTool(ctx, turnID, tc)
			turn.Items = append(turn.Items, item)
		}
		if chunk.Done {
			break
		}
	}

	if len(assistantBuf) > 0 {
		t := string(assistantBuf)
		turn.Items = append(turn.Items, contract.TurnItem{
			ID:        contract.ItemID(fmt.Sprintf("it-%d-asst", time.Now().UnixNano())),
			Kind:      contract.ItemKindAssistantText,
			CreatedAt: time.Now().UTC(),
			Text:      &t,
		})
	}

	turn.Status = contract.TurnStatusCompleted
	a.state.EndTurn(turnID, nil)
	return turn, nil
}

// dispatchTool looks up the tool by name, invokes it with a Context,
// and returns a TurnItem of kind tool_call or tool_result.
func (a *ChatAgent) dispatchTool(ctx context.Context, turnID contract.TurnID, tc model.ToolCall) contract.TurnItem {
	a.toolsMu.RLock()
	t, ok := a.tools[tc.Name]
	a.toolsMu.RUnlock()

	now := time.Now().UTC()
	itemID := contract.ItemID(fmt.Sprintf("it-%d-tc", time.Now().UnixNano()))

	if !ok {
		errTxt := fmt.Sprintf("tool not found: %s", tc.Name)
		return contract.TurnItem{
			ID:        itemID,
			Kind:      contract.ItemKindToolResult,
			CreatedAt: now,
			ToolCall: &contract.ToolCall{
				ID:    contract.ToolCallID(tc.ID),
				Name:  tc.Name,
				Input: json.RawMessage(tc.Input),
				Error: errTxt,
			},
		}
	}

	toolCtx := tool.NewContext(ctx, "", string(turnID), string(contract.ToolCallID(tc.ID)), "")
	res, err := t.Execute(ctx, json.RawMessage(tc.Input), toolCtx)
	item := contract.TurnItem{
		ID:        itemID,
		Kind:      contract.ItemKindToolResult,
		CreatedAt: now,
		ToolCall: &contract.ToolCall{
			ID:    contract.ToolCallID(tc.ID),
			Name:  tc.Name,
			Input: json.RawMessage(tc.Input),
		},
	}
	if res != nil {
		item.ToolCall.Output = json.RawMessage(mustMarshal(res))
	}
	if err != nil {
		item.ToolCall.Error = err.Error()
	} else if res != nil && !res.OK {
		item.ToolCall.Error = res.Error
	}
	return item
}

func mustMarshal(r *tool.Result) []byte {
	if r == nil {
		return []byte("null")
	}
	b, err := json.Marshal(r)
	if err != nil {
		return []byte(`{"error":"marshal failed"}`)
	}
	return b
}
