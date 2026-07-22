package runtime

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/tool"
)

type stubTool struct {
	name   string
	result *tool.Result
}

func (t *stubTool) Name() string                                        { return t.name }
func (t *stubTool) Description() string                                 { return t.name + " tool" }
func (t *stubTool) Schema() json.RawMessage                            { return json.RawMessage(`{}`) }
func (t *stubTool) Execute(_ context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	return t.result, nil
}

type stubClient struct {
	responses []responseSeq
}

type responseSeq struct {
	text      string
	toolCalls []model.ToolCall
}

func (c *stubClient) Stream(_ context.Context, _ model.Request) (<-chan model.StreamChunk, error) {
	ch := make(chan model.StreamChunk, 10)
	go func() {
		defer close(ch)
		for _, r := range c.responses {
			if r.text != "" {
				ch <- model.StreamChunk{DeltaText: r.text}
			}
			for i, tc := range r.toolCalls {
				_ = i
				ch <- model.StreamChunk{ToolCall: &tc}
			}
			ch <- model.StreamChunk{Done: true}
		}
	}()
	return ch, nil
}

func TestTurnLoopNoTools(t *testing.T) {
	client := &stubClient{responses: []responseSeq{
		{text: "Hello!"},
	}}
	loop := NewTurnLoop(client, nil, nil, nil)

	thread := &contract.Thread{ID: "thr1"}
	turn, err := loop.RunTurn(context.Background(), thread, contract.UserInput{Text: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if turn.Status != contract.TurnStatusCompleted {
		t.Errorf("status = %s, want completed", turn.Status)
	}
	if len(turn.Items) != 2 { // user + assistant
		t.Errorf("items = %d, want 2", len(turn.Items))
	}
}

func TestTurnLoopWithToolCall(t *testing.T) {
	client := &stubClient{responses: []responseSeq{
		{toolCalls: []model.ToolCall{{ID: "tc1", Name: "echo", Input: json.RawMessage(`{"msg":"hi"}`)}}},
		{text: "Done"},
	}}
	echoTool := &stubTool{name: "echo", result: &tool.Result{OK: true, Content: `{"out":"hi"}`}}
	loop := NewTurnLoop(client, []tool.Tool{echoTool}, nil, nil)

	thread := &contract.Thread{ID: "thr1"}
	turn, err := loop.RunTurn(context.Background(), thread, contract.UserInput{Text: "echo hi"})
	if err != nil {
		t.Fatal(err)
	}
	if turn.Status != contract.TurnStatusCompleted {
		t.Errorf("status = %s, want completed", turn.Status)
	}
	// user + tool result + assistant
	if len(turn.Items) < 2 {
		t.Errorf("items = %d, want >= 2", len(turn.Items))
	}
}

func TestTurnLoopToolNotFound(t *testing.T) {
	client := &stubClient{responses: []responseSeq{
		{toolCalls: []model.ToolCall{{ID: "tc1", Name: "nosuch", Input: json.RawMessage(`{}`)}}},
		{text: "ok"},
	}}
	loop := NewTurnLoop(client, nil, nil, nil)

	thread := &contract.Thread{ID: "thr1"}
	turn, err := loop.RunTurn(context.Background(), thread, contract.UserInput{Text: "do it"})
	if err != nil {
		t.Fatal(err)
	}
	// Tool not found produces an error in ToolCall, but loop continues.
	var foundError bool
	for _, it := range turn.Items {
		if it.ToolCall != nil && it.ToolCall.Error != "" {
			foundError = true
		}
	}
	if !foundError {
		t.Error("expected tool not found error in items")
	}
}

func TestTurnLoopCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := &stubClient{responses: []responseSeq{
		{text: "Hello"},
	}}
	loop := NewTurnLoop(client, nil, nil, nil)

	thread := &contract.Thread{ID: "thr1"}
	_, err := loop.RunTurn(ctx, thread, contract.UserInput{Text: "hi"})
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}
